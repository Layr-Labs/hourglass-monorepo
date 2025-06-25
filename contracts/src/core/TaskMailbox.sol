// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {
    IBN254CertificateVerifier,
    IBN254CertificateVerifierTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";

import {IAVSTaskHook} from "../interfaces/avs/l2/IAVSTaskHook.sol";
import {ITaskMailbox} from "../interfaces/core/ITaskMailbox.sol";
import {TaskMailboxStorage} from "./TaskMailboxStorage.sol";

/**
 * @title TaskMailbox
 * @author Layr Labs, Inc.
 * @notice Contract for managing the lifecycle of tasks that are executed by operator sets of task-based AVSs.
 */
contract TaskMailbox is Ownable, ReentrancyGuard, TaskMailboxStorage {
    using SafeERC20 for IERC20;
    using SafeCast for *;

    /**
     * @notice Constructor for TaskMailbox
     * @param _owner The owner of the contract
     * @param _certificateVerifiers Array of certificate verifier configs
     */
    constructor(address _owner, CertificateVerifierConfig[] memory _certificateVerifiers) Ownable() {
        _transferOwnership(_owner);

        for (uint256 i = 0; i < _certificateVerifiers.length; i++) {
            _setCertificateVerifier(_certificateVerifiers[i].curveType, _certificateVerifiers[i].verifier);
        }
    }

    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */

    /// @inheritdoc ITaskMailbox
    function setCertificateVerifier(
        IKeyRegistrarTypes.CurveType curveType,
        address certificateVerifier
    ) external onlyOwner {
        _setCertificateVerifier(curveType, certificateVerifier);
    }

    /// @inheritdoc ITaskMailbox
    function registerExecutorOperatorSet(OperatorSet memory operatorSet, bool isRegistered) external {
        // TODO: Only OperatorSetOwner can register executor operator set.

        _registerExecutorOperatorSet(operatorSet, isRegistered);
    }

    /// @inheritdoc ITaskMailbox
    function setExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet,
        ExecutorOperatorSetTaskConfig memory config
    ) external {
        // TODO: Only OperatorSetOwner can set config.

        // TODO: Do we need to make taskHook ERC165 compliant? and check for ERC165 interface support?
        // TODO: Double check if any other config checks are needed.

        require(config.curveType != IKeyRegistrarTypes.CurveType.NONE, InvalidCurveType());
        require(config.taskHook != IAVSTaskHook(address(0)), InvalidAddressZero());
        require(config.taskSLA > 0, TaskSLAIsZero());

        // If executor operator set is not registered, register it.
        if (!isExecutorOperatorSetRegistered[operatorSet.key()]) {
            _registerExecutorOperatorSet(operatorSet, true);
        }

        executorOperatorSetTaskConfigs[operatorSet.key()] = config;
        emit ExecutorOperatorSetTaskConfigSet(msg.sender, operatorSet.avs, operatorSet.id, config);
    }

    /// @inheritdoc ITaskMailbox
    function createTask(
        TaskParams memory taskParams
    ) external nonReentrant returns (bytes32) {
        // TODO: `Created` status cannot be enum value 0 since that is the default value. Figure out how to handle this.

        require(taskParams.payload.length > 0, PayloadIsEmpty());
        require(
            isExecutorOperatorSetRegistered[taskParams.executorOperatorSet.key()], ExecutorOperatorSetNotRegistered()
        );

        ExecutorOperatorSetTaskConfig memory taskConfig =
            executorOperatorSetTaskConfigs[taskParams.executorOperatorSet.key()];
        require(
            taskConfig.curveType != IKeyRegistrarTypes.CurveType.NONE && address(taskConfig.taskHook) != address(0)
                && taskConfig.taskSLA > 0,
            ExecutorOperatorSetTaskConfigNotSet()
        );

        // Pre-task submission checks: AVS can validate the caller, operator set and task payload
        taskConfig.taskHook.validatePreTaskCreation(msg.sender, taskParams.executorOperatorSet, taskParams.payload);

        bytes32 taskHash = keccak256(abi.encode(globalTaskCount, address(this), block.chainid, taskParams));
        globalTaskCount = globalTaskCount + 1;

        tasks[taskHash] = Task(
            msg.sender,
            block.timestamp.toUint96(),
            TaskStatus.Created,
            taskParams.executorOperatorSet.avs,
            taskParams.executorOperatorSet.id,
            taskParams.refundCollector,
            taskParams.avsFee,
            0, // TODO: Update with fee split % variable
            taskConfig,
            taskParams.payload,
            bytes("")
        );

        // TODO: Need a separate permissionless function to do the final transfer from this contract to AVS (or back to App)
        if (taskConfig.feeToken != IERC20(address(0)) && taskParams.avsFee > 0) {
            // TODO: Might need a separate variable for tracking balance transfer.
            taskConfig.feeToken.safeTransferFrom(msg.sender, address(this), taskParams.avsFee);
        }

        // Post-task submission checks:
        // 1. AVS can write to storage in their hook for validating task lifecycle
        // 2. AVS can design fee markets to validate their avsFee against.
        taskConfig.taskHook.handlePostTaskCreation(taskHash);

        emit TaskCreated(
            msg.sender,
            taskHash,
            taskParams.executorOperatorSet.avs,
            taskParams.executorOperatorSet.id,
            taskParams.refundCollector,
            taskParams.avsFee,
            block.timestamp + taskConfig.taskSLA,
            taskParams.payload
        );
        return taskHash;
    }

    /// @inheritdoc ITaskMailbox
    function cancelTask(
        bytes32 taskHash
    ) external {
        // TODO: Check if we even need this cancelTask function - Maybe have a flag with isCancelable in the AVS Task Config and further gate at the protocol level.
        Task storage task = tasks[taskHash];
        TaskStatus status = _getTaskStatus(task);
        require(status == TaskStatus.Created, InvalidTaskStatus(TaskStatus.Created, status));
        require(msg.sender == task.creator, InvalidTaskCreator());
        require(block.timestamp > task.creationTime, TimestampAtCreation());

        task.status = TaskStatus.Canceled;

        emit TaskCanceled(msg.sender, taskHash, task.avs, task.executorOperatorSetId);
    }

    /// @inheritdoc ITaskMailbox
    function submitResult(
        bytes32 taskHash,
        IBN254CertificateVerifierTypes.BN254Certificate memory cert,
        bytes memory result
    ) external nonReentrant {
        // TODO: Do we need a gasless version of this function?
        // TODO: require checks - Figure out what checks are needed
        Task storage task = tasks[taskHash];
        TaskStatus status = _getTaskStatus(task);
        require(status == TaskStatus.Created, InvalidTaskStatus(TaskStatus.Created, status));
        require(block.timestamp > task.creationTime, TimestampAtCreation());

        uint16[] memory totalStakeProportionThresholds = new uint16[](1);
        totalStakeProportionThresholds[0] = task.executorOperatorSetTaskConfig.stakeProportionThreshold;
        OperatorSet memory executorOperatorSet = OperatorSet(task.avs, task.executorOperatorSetId);

        address verifier = certificateVerifiers[task.executorOperatorSetTaskConfig.curveType];
        require(verifier != address(0), InvalidAddressZero());

        bool isCertificateValid = IBN254CertificateVerifier(verifier).verifyCertificateProportion(
            executorOperatorSet, cert, totalStakeProportionThresholds
        );

        require(isCertificateValid, CertificateVerificationFailed());

        task.status = TaskStatus.Verified;
        task.result = result;

        // TODO: Check what happens if we re-ennter the other state transition functions.
        // Task result submission checks:
        // 1. AVS can validate the task result, params and certificate.
        // 2. It can update hook storage for task lifecycle if needed.
        task.executorOperatorSetTaskConfig.taskHook.handleTaskResultSubmission(taskHash, cert);

        emit TaskVerified(msg.sender, taskHash, task.avs, task.executorOperatorSetId, task.result);
    }

    /**
     *
     *                         INTERNAL FUNCTIONS
     *
     */

    /**
     * @notice Gets the current status of a task
     * @param task The task to get the status for
     * @return The current status of the task, considering expiration
     */
    function _getTaskStatus(
        Task memory task
    ) internal view returns (TaskStatus) {
        if (
            task.status == TaskStatus.Created
                && block.timestamp > (task.creationTime + task.executorOperatorSetTaskConfig.taskSLA)
        ) {
            return TaskStatus.Expired;
        }
        return task.status;
    }

    /**
     * @notice Registers an executor operator set with the TaskMailbox
     * @param operatorSet The operator set to register
     * @param isRegistered Whether the operator set is registered
     */
    function _registerExecutorOperatorSet(OperatorSet memory operatorSet, bool isRegistered) internal {
        isExecutorOperatorSetRegistered[operatorSet.key()] = isRegistered;
        emit ExecutorOperatorSetRegistered(msg.sender, operatorSet.avs, operatorSet.id, isRegistered);
    }

    /**
     * @notice Sets a certificate verifier for a specific curve type
     * @param curveType The curve type for the verifier
     * @param certificateVerifier Address of the certificate verifier
     */
    function _setCertificateVerifier(IKeyRegistrarTypes.CurveType curveType, address certificateVerifier) internal {
        require(certificateVerifier != address(0), InvalidAddressZero());
        certificateVerifiers[curveType] = certificateVerifier;
        emit CertificateVerifierSet(curveType, certificateVerifier);
    }

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */

    /// @inheritdoc ITaskMailbox
    function getCertificateVerifier(
        IKeyRegistrarTypes.CurveType curveType
    ) external view returns (address) {
        return certificateVerifiers[curveType];
    }

    /// @inheritdoc ITaskMailbox
    function getExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet
    ) external view returns (ExecutorOperatorSetTaskConfig memory) {
        return executorOperatorSetTaskConfigs[operatorSet.key()];
    }

    /// @inheritdoc ITaskMailbox
    function getTaskInfo(
        bytes32 taskHash
    ) external view returns (Task memory) {
        Task memory task = tasks[taskHash];
        return Task(
            task.creator,
            task.creationTime,
            _getTaskStatus(task),
            task.avs,
            task.executorOperatorSetId,
            task.refundCollector,
            task.avsFee,
            task.feeSplit,
            task.executorOperatorSetTaskConfig,
            task.payload,
            task.result
        );
    }

    /// @inheritdoc ITaskMailbox
    function getTaskStatus(
        bytes32 taskHash
    ) external view returns (TaskStatus) {
        Task memory task = tasks[taskHash];
        return _getTaskStatus(task);
    }

    /// @inheritdoc ITaskMailbox
    function getTaskResult(
        bytes32 taskHash
    ) external view returns (bytes memory) {
        Task memory task = tasks[taskHash];
        TaskStatus status = _getTaskStatus(task);
        require(status == TaskStatus.Verified, InvalidTaskStatus(TaskStatus.Verified, status));
        return task.result;
    }
}
