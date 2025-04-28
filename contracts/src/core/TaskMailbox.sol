// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {
    OperatorSet,
    OperatorSetLib
} from "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {ReentrancyGuard} from "@eigenlayer-middleware/lib/openzeppelin-contracts/contracts/security/ReentrancyGuard.sol";
import {IERC20} from "@eigenlayer-middleware/lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@eigenlayer-middleware/lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";
import {SafeCast} from "@eigenlayer-middleware/lib/openzeppelin-contracts/contracts/utils/math/SafeCast.sol";

import {IAVSTaskHook} from "src/interfaces/IAVSTaskHook.sol";
import {IBN254CertificateVerifier} from "src/interfaces/IBN254CertificateVerifier.sol";
import {TaskMailBoxStorage} from "src/core/TaskMailBoxStorage.sol";

contract TaskMailbox is ReentrancyGuard, TaskMailBoxStorage {
    // TODO: Decide if we want to make contract a transparent proxy with owner set up. And add Pausable.

    using SafeERC20 for IERC20;
    using SafeCast for *;

    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */
    function registerOperatorSet(OperatorSet memory operatorSet, bool isRegistered) external {
        // TODO: require checks - Figure out what checks are needed.
        // 1. OperatorSet is valid
        // 2. Only AVS delegated address can (de)register.
        _registerOperatorSet(operatorSet, isRegistered);
    }

    function setOperatorSetTaskConfig(OperatorSet memory operatorSet, OperatorSetTaskConfig memory config) external {
        // TODO: require checks - Figure out what checks are needed.
        // 1. OperatorSet is valid
        // 2. Only AVS delegated address can set config.

        // TODO: Do we need to make taskHook ERC165 compliant? and check for ERC165 interface support?
        // TODO: Double check if any other config checks are needed.
        require(config.certificateVerifier != address(0), InvalidAddressZero());
        require(config.taskHook != IAVSTaskHook(address(0)), InvalidAddressZero());
        require(config.aggregator != address(0), InvalidAddressZero());
        require(config.taskSLA > 0, TaskSLAIsZero());

        // If operator set is not registered, register it.
        if (!isOperatorSetRegistered[operatorSet.key()]) {
            _registerOperatorSet(operatorSet, true);
        }

        operatorSetTaskConfig[operatorSet.key()] = config;
        emit OperatorSetTaskConfigSet(msg.sender, operatorSet.avs, operatorSet.id, config);
    }

    function createTask(
        TaskParams memory taskParams
    ) external nonReentrant returns (bytes32) {
        // TODO: require checks - Figure out what checks are needed
        // 1. OperatorSet is valid
        // TODO: Do we need a gasless version of this function?

        require(isOperatorSetRegistered[taskParams.operatorSet.key()], OperatorSetNotRegistered());
        require(taskParams.payload.length > 0, PayloadIsEmpty());

        OperatorSetTaskConfig memory config = operatorSetTaskConfig[taskParams.operatorSet.key()];
        require(
            config.certificateVerifier != address(0) && address(config.taskHook) != address(0)
                && config.aggregator != address(0) && config.taskSLA != 0,
            OperatorSetTaskConfigNotSet()
        );

        // Pre-task submission checks: AVS can validate the caller, operator set and task payload
        config.taskHook.validatePreTaskCreation(msg.sender, taskParams.operatorSet, taskParams.payload);

        bytes32 taskHash = keccak256(abi.encode(globalTaskCount, address(this), block.chainid, taskParams));
        globalTaskCount = globalTaskCount + 1;

        tasks[taskHash] = Task(
            msg.sender,
            block.timestamp.toUint96(),
            TaskStatus.Created,
            taskParams.operatorSet,
            taskParams.refundCollector,
            taskParams.avsFee,
            0, // TODO: Update with fee split %
            config,
            taskParams.payload,
            bytes("")
        );

        // TODO: Need a separate permissionless function to do the final transfer from this contract to AVS (or back to App)
        if (config.feeToken != IERC20(address(0)) && taskParams.avsFee > 0) {
            // TODO: Might need a separate variable for tracking balance transfer.
            config.feeToken.safeTransferFrom(msg.sender, address(this), taskParams.avsFee);
        }

        // Post-task submission checks:
        // 1. AVS can write to storage in their hook for validating task lifecycle
        // 2. AVS can design fee markets to validate their avsFee against.
        config.taskHook.validatePostTaskCreation(taskHash);

        emit TaskCreated(
            msg.sender,
            taskHash,
            taskParams.operatorSet.avs,
            taskParams.operatorSet.id,
            taskParams.refundCollector,
            taskParams.avsFee,
            block.timestamp + config.taskSLA,
            taskParams.payload
        );
        return taskHash;
    }

    function cancelTask(
        bytes32 taskHash
    ) external {
        // TODO: Check if we even need this cancelTask function - Maybe have a flag with isCancelable in the AVS Task Config.
        Task storage task = tasks[taskHash];
        TaskStatus status = _getTaskStatus(task);
        require(status == TaskStatus.Created, InvalidTaskStatus(TaskStatus.Created, status));
        require(msg.sender == task.creator, InvalidTaskCreator());
        require(block.timestamp > task.creationTime, TimestampAtCreation());

        task.status = TaskStatus.Canceled;

        emit TaskCanceled(msg.sender, taskHash, task.operatorSet.avs, task.operatorSet.id);
    }

    function submitResult(
        bytes32 taskHash,
        IBN254CertificateVerifier.BN254Certificate memory cert,
        bytes memory result
    ) external {
        // TODO: require checks - Figure out what checks are needed
        Task storage task = tasks[taskHash];
        TaskStatus status = _getTaskStatus(task);
        require(status == TaskStatus.Created, InvalidTaskStatus(TaskStatus.Created, status));
        require(msg.sender == task.operatorSetTaskConfig.aggregator, InvalidTaskAggregator());
        require(block.timestamp > task.creationTime, TimestampAtCreation());

        uint16[] memory totalStakeProportionThresholds = new uint16[](1);
        totalStakeProportionThresholds[0] = task.operatorSetTaskConfig.stakeProportionThreshold;
        bool isCertificateValid = IBN254CertificateVerifier(task.operatorSetTaskConfig.certificateVerifier)
            .verifyCertificateProportion(cert, totalStakeProportionThresholds);

        require(isCertificateValid, CertificateVerificationFailed());

        task.status = TaskStatus.Verified;
        task.result = result;

        // TODO: Check what happens if we re-ennter the other state transition functions.
        // Task result submission checks:
        // 1. AVS can validate the task result, params and certificate.
        // 2. It can update hook storage for task lifecycle if needed.
        task.operatorSetTaskConfig.taskHook.validateTaskResultSubmission(taskHash, cert);

        emit TaskVerified(msg.sender, taskHash, task.operatorSet.avs, task.operatorSet.id, task.result);
    }

    /**
     *
     *                         INTERNAL FUNCTIONS
     *
     */
    function _getTaskStatus(
        Task memory task
    ) internal view returns (TaskStatus) {
        if (
            task.status == TaskStatus.Created
                && block.timestamp > (task.creationTime + task.operatorSetTaskConfig.taskSLA)
        ) {
            return TaskStatus.Expired;
        }
        return task.status;
    }

    function _registerOperatorSet(OperatorSet memory operatorSet, bool isRegistered) internal {
        isOperatorSetRegistered[operatorSet.key()] = isRegistered;
        emit OperatorSetRegistered(msg.sender, operatorSet.avs, operatorSet.id, isRegistered);
    }

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */
    function getOperatorSetTaskConfig(
        OperatorSet memory operatorSet
    ) external view returns (OperatorSetTaskConfig memory) {
        return operatorSetTaskConfig[operatorSet.key()];
    }

    function getTaskInfo(
        bytes32 taskHash
    ) external view returns (Task memory) {
        Task memory task = tasks[taskHash];
        return Task(
            task.creator,
            task.creationTime,
            _getTaskStatus(task),
            task.operatorSet,
            task.refundCollector,
            task.avsFee,
            0, // TODO: Update with fee split %
            task.operatorSetTaskConfig,
            task.payload,
            task.result
        );
    }

    function getTaskStatus(
        bytes32 taskHash
    ) external view returns (TaskStatus) {
        Task memory task = tasks[taskHash];
        return _getTaskStatus(task);
    }

    function getTaskResult(
        bytes32 taskHash
    ) external view returns (bytes memory) {
        Task memory task = tasks[taskHash];
        TaskStatus status = _getTaskStatus(task);
        require(status == TaskStatus.Verified, InvalidTaskStatus(TaskStatus.Verified, status));
        return task.result;
    }
}
