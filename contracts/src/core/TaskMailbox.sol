// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {
    IBN254CertificateVerifier,
    IBN254CertificateVerifierTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {
    IECDSACertificateVerifier,
    IECDSACertificateVerifierTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IECDSACertificateVerifier.sol";
import {IBaseCertificateVerifier} from "@eigenlayer-contracts/src/contracts/interfaces/IBaseCertificateVerifier.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin-upgrades/contracts/security/ReentrancyGuardUpgradeable.sol";
import {OwnableUpgradeable} from "@openzeppelin-upgrades/contracts/access/OwnableUpgradeable.sol";
import {Initializable} from "@openzeppelin-upgrades/contracts/proxy/utils/Initializable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";

import {IAVSTaskHook} from "../interfaces/avs/l2/IAVSTaskHook.sol";
import {ITaskMailbox} from "../interfaces/core/ITaskMailbox.sol";
import {TaskMailboxStorage} from "./TaskMailboxStorage.sol";
import {SemVerMixin} from "@eigenlayer-contracts/src/contracts/mixins/SemVerMixin.sol";

/**
 * @title TaskMailbox
 * @author Layr Labs, Inc.
 * @notice Contract for managing the lifecycle of tasks that are executed by operator sets of task-based AVSs.
 */
contract TaskMailbox is
    Initializable,
    OwnableUpgradeable,
    ReentrancyGuardUpgradeable,
    TaskMailboxStorage,
    SemVerMixin
{
    using SafeERC20 for IERC20;
    using SafeCast for *;

    /**
     * @notice Constructor for TaskMailbox
     * @param _bn254CertificateVerifier Address of the BN254 certificate verifier
     * @param _ecdsaCertificateVerifier Address of the ECDSA certificate verifier
     * @param _version The semantic version of the contract
     */
    constructor(
        address _bn254CertificateVerifier,
        address _ecdsaCertificateVerifier,
        string memory _version
    ) TaskMailboxStorage(_bn254CertificateVerifier, _ecdsaCertificateVerifier) SemVerMixin(_version) {
        _disableInitializers();
    }

    /**
     * @notice Initializer for TaskMailbox
     * @param _owner The owner of the contract
     */
    function initialize(
        address _owner
    ) external initializer {
        __Ownable_init();
        __ReentrancyGuard_init();
        _transferOwnership(_owner);
    }

    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */

    /// @inheritdoc ITaskMailbox
    function setExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet,
        ExecutorOperatorSetTaskConfig memory config
    ) external {
        address certificateVerifier = _getCertificateVerifier(config.curveType);
        require(
            IBaseCertificateVerifier(certificateVerifier).getOperatorSetOwner(operatorSet) == msg.sender,
            InvalidOperatorSetOwner()
        );

        // TODO: Double check if any other config checks are needed.
        require(config.curveType != IKeyRegistrarTypes.CurveType.NONE, InvalidCurveType());
        require(config.taskHook != IAVSTaskHook(address(0)), InvalidAddressZero());
        require(config.taskSLA > 0, TaskSLAIsZero());

        executorOperatorSetTaskConfigs[operatorSet.key()] = config;
        emit ExecutorOperatorSetTaskConfigSet(msg.sender, operatorSet.avs, operatorSet.id, config);

        // If executor operator set is not registered, register it.
        if (!isExecutorOperatorSetRegistered[operatorSet.key()]) {
            _registerExecutorOperatorSet(operatorSet, true);
        }
    }

    /// @inheritdoc ITaskMailbox
    function registerExecutorOperatorSet(OperatorSet memory operatorSet, bool isRegistered) external {
        ExecutorOperatorSetTaskConfig memory taskConfig = executorOperatorSetTaskConfigs[operatorSet.key()];

        require(
            taskConfig.curveType != IKeyRegistrarTypes.CurveType.NONE && address(taskConfig.taskHook) != address(0)
                && taskConfig.taskSLA > 0,
            ExecutorOperatorSetTaskConfigNotSet()
        );
        address certificateVerifier = _getCertificateVerifier(taskConfig.curveType);
        require(
            IBaseCertificateVerifier(certificateVerifier).getOperatorSetOwner(operatorSet) == msg.sender,
            InvalidOperatorSetOwner()
        );

        _registerExecutorOperatorSet(operatorSet, isRegistered);
    }

    /// @inheritdoc ITaskMailbox
    function createTask(
        TaskParams memory taskParams
    ) external nonReentrant returns (bytes32) {
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

        // Pre-task submission checks:
        // 1. AVS can validate the caller and task params.
        // 2. AVS can design fee markets to validate their avsFee against.
        taskConfig.taskHook.validatePreTaskCreation(msg.sender, taskParams);

        bytes32 taskHash = keccak256(abi.encode(_globalTaskCount, address(this), block.chainid, taskParams));
        _globalTaskCount = _globalTaskCount + 1;

        _tasks[taskHash] = Task(
            msg.sender,
            block.timestamp.toUint96(),
            TaskStatus.CREATED,
            taskParams.executorOperatorSet.avs,
            taskParams.executorOperatorSet.id,
            taskParams.refundCollector,
            taskParams.avsFee,
            0, // TODO: Update with fee split % variable
            taskConfig,
            taskParams.payload,
            bytes(""),
            bytes("")
        );

        // TODO: Need a separate permissionless function to do the final transfer from this contract to AVS (or back to App)
        if (taskConfig.feeToken != IERC20(address(0)) && taskParams.avsFee > 0) {
            // TODO: Might need a separate variable for tracking balance transfer.
            taskConfig.feeToken.safeTransferFrom(msg.sender, address(this), taskParams.avsFee);
        }

        // Post-task submission checks: AVS can write to storage in their hook for validating task lifecycle
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
    function submitResult(bytes32 taskHash, bytes memory cert, bytes memory result) external nonReentrant {
        // TODO: Handle case of anyone submitting a result with empty signature in the certificate.

        Task storage task = _tasks[taskHash];
        TaskStatus status = _getTaskStatus(task);
        require(status == TaskStatus.CREATED, InvalidTaskStatus(TaskStatus.CREATED, status));
        require(block.timestamp > task.creationTime, TimestampAtCreation());

        // Pre-task result submission checks: AVS can validate the caller, task result, params and certificate.
        task.executorOperatorSetTaskConfig.taskHook.validatePreTaskResultSubmission(msg.sender, taskHash, cert, result);

        uint16[] memory totalStakeProportionThresholds = new uint16[](1);
        totalStakeProportionThresholds[0] = task.executorOperatorSetTaskConfig.stakeProportionThreshold;

        OperatorSet memory executorOperatorSet = OperatorSet(task.avs, task.executorOperatorSetId);
        bool isCertificateValid;
        if (task.executorOperatorSetTaskConfig.curveType == IKeyRegistrarTypes.CurveType.BN254) {
            // BN254 Certificate verification
            IBN254CertificateVerifierTypes.BN254Certificate memory bn254Cert =
                abi.decode(cert, (IBN254CertificateVerifierTypes.BN254Certificate));
            isCertificateValid = IBN254CertificateVerifier(BN254_CERTIFICATE_VERIFIER).verifyCertificateProportion(
                executorOperatorSet, bn254Cert, totalStakeProportionThresholds
            );
        } else if (task.executorOperatorSetTaskConfig.curveType == IKeyRegistrarTypes.CurveType.ECDSA) {
            // ECDSA Certificate verification
            IECDSACertificateVerifierTypes.ECDSACertificate memory ecdsaCert =
                abi.decode(cert, (IECDSACertificateVerifierTypes.ECDSACertificate));
            isCertificateValid = IECDSACertificateVerifier(ECDSA_CERTIFICATE_VERIFIER).verifyCertificateProportion(
                executorOperatorSet, ecdsaCert, totalStakeProportionThresholds
            );
        } else {
            revert InvalidCurveType();
        }
        require(isCertificateValid, CertificateVerificationFailed());

        task.status = TaskStatus.VERIFIED;
        task.executorCert = cert;
        task.result = result;

        // Task result submission checks: AVS can update hook storage for task lifecycle if needed.
        task.executorOperatorSetTaskConfig.taskHook.handlePostTaskResultSubmission(taskHash);

        emit TaskVerified(msg.sender, taskHash, task.avs, task.executorOperatorSetId, task.executorCert, task.result);
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
            task.status == TaskStatus.CREATED
                && block.timestamp > (task.creationTime + task.executorOperatorSetTaskConfig.taskSLA)
        ) {
            return TaskStatus.EXPIRED;
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
     * @notice Gets the certificate verifier for a specific curve type
     * @param curveType The curve type for the verifier
     * @return The address of the certificate verifier
     */
    function _getCertificateVerifier(
        IKeyRegistrarTypes.CurveType curveType
    ) internal view returns (address) {
        if (curveType == IKeyRegistrarTypes.CurveType.BN254) {
            return BN254_CERTIFICATE_VERIFIER;
        } else if (curveType == IKeyRegistrarTypes.CurveType.ECDSA) {
            return ECDSA_CERTIFICATE_VERIFIER;
        } else {
            revert InvalidCurveType();
        }
    }

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */

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
        Task memory task = _tasks[taskHash];
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
            task.executorCert,
            task.result
        );
    }

    /// @inheritdoc ITaskMailbox
    function getTaskStatus(
        bytes32 taskHash
    ) external view returns (TaskStatus) {
        Task memory task = _tasks[taskHash];
        return _getTaskStatus(task);
    }

    /// @inheritdoc ITaskMailbox
    function getTaskResult(
        bytes32 taskHash
    ) external view returns (bytes memory) {
        Task memory task = _tasks[taskHash];
        TaskStatus status = _getTaskStatus(task);
        require(status == TaskStatus.VERIFIED, InvalidTaskStatus(TaskStatus.VERIFIED, status));
        return task.result;
    }

    /// @inheritdoc ITaskMailbox
    function getBN254CertificateBytes(
        IBN254CertificateVerifierTypes.BN254Certificate memory cert
    ) external pure returns (bytes memory) {
        return abi.encode(cert);
    }

    /// @inheritdoc ITaskMailbox
    function getECDSACertificateBytes(
        IECDSACertificateVerifierTypes.ECDSACertificate memory cert
    ) external pure returns (bytes memory) {
        return abi.encode(cert);
    }
}
