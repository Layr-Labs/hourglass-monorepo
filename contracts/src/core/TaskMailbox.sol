// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";

import {IAVSTaskHook} from "../interfaces/avs/l2/IAVSTaskHook.sol";
import {IBN254CertificateVerifier} from "../interfaces/avs/l2/IBN254CertificateVerifier.sol";
import {TaskMailboxStorage} from "./TaskMailboxStorage.sol";

contract TaskMailbox is ReentrancyGuard, TaskMailboxStorage {
    // TODO: Decide if we want to make contract a transparent proxy with owner set up. And add Pausable and Ownable.

    using SafeERC20 for IERC20;
    using SafeCast for *;

    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */
    function registerAvs(address avs, bool isRegistered) external {
        // TODO: require checks - Figure out what checks are needed.
        // 1. AVS is valid
        // 2. Only AVS delegated address can (de)register.
        _registerAvs(avs, isRegistered);
    }

    function setAvsConfig(address avs, AvsConfig memory config) external {
        // TODO: require checks - Figure out what checks are needed.
        // 1. OperatorSets are valid
        // 2. Only AVS delegated address can set config.

        AvsConfig memory memAvsConfig = avsConfigs[avs];
        // Deregister all current executor operator sets.
        for (uint256 i = 0; i < memAvsConfig.executorOperatorSetIds.length; i++) {
            OperatorSet memory executorOperatorSet = OperatorSet(avs, memAvsConfig.executorOperatorSetIds[i]);
            isExecutorOperatorSetRegistered[executorOperatorSet.key()] = false;
        }

        // Register new executor operator sets.
        for (uint256 i = 0; i < config.executorOperatorSetIds.length; i++) {
            OperatorSet memory executorOperatorSet = OperatorSet(avs, config.executorOperatorSetIds[i]);
            require(config.aggregatorOperatorSetId != executorOperatorSet.id, InvalidAggregatorOperatorSetId());
            require(!isExecutorOperatorSetRegistered[executorOperatorSet.key()], DuplicateExecutorOperatorSetId());
            isExecutorOperatorSetRegistered[executorOperatorSet.key()] = true;
        }

        // If AVS is not registered, register it.
        if (!isAvsRegistered[avs]) {
            _registerAvs(avs, true);
        }

        avsConfigs[avs] = config;
        emit AvsConfigSet(msg.sender, avs, config.aggregatorOperatorSetId, config.executorOperatorSetIds);
    }

    function setExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet,
        ExecutorOperatorSetTaskConfig memory config
    ) external {
        // TODO: require checks - Figure out what checks are needed.
        // 1. OperatorSet is valid
        // 2. Only AVS delegated address can set config.

        // TODO: Do we need to make taskHook ERC165 compliant? and check for ERC165 interface support?
        // TODO: Double check if any other config checks are needed.
        require(isExecutorOperatorSetRegistered[operatorSet.key()], ExecutorOperatorSetNotRegistered());
        require(config.certificateVerifier != address(0), InvalidAddressZero());
        require(config.taskHook != IAVSTaskHook(address(0)), InvalidAddressZero());
        require(config.taskSLA > 0, TaskSLAIsZero());

        executorOperatorSetTaskConfigs[operatorSet.key()] = config;
        emit ExecutorOperatorSetTaskConfigSet(msg.sender, operatorSet.avs, operatorSet.id, config);
    }

    function createTask(
        TaskParams memory taskParams
    ) external nonReentrant returns (bytes32) {
        // TODO: require checks - Figure out what checks are needed
        // 1. OperatorSet is valid
        // TODO: Do we need a gasless version of this function?
        // TODO: `Created` status cannot be enum value 0 since that is the default value. Figure out how to handle this.

        require(isAvsRegistered[taskParams.executorOperatorSet.avs], AvsNotRegistered());
        require(
            isExecutorOperatorSetRegistered[taskParams.executorOperatorSet.key()], ExecutorOperatorSetNotRegistered()
        );
        require(taskParams.payload.length > 0, PayloadIsEmpty());

        ExecutorOperatorSetTaskConfig memory taskConfig =
            executorOperatorSetTaskConfigs[taskParams.executorOperatorSet.key()];
        require(
            taskConfig.certificateVerifier != address(0) && address(taskConfig.taskHook) != address(0)
                && taskConfig.taskSLA > 0,
            ExecutorOperatorSetTaskConfigNotSet()
        );

        // Pre-task submission checks: AVS can validate the caller, operator set and task payload
        taskConfig.taskHook.validatePreTaskCreation(msg.sender, taskParams.executorOperatorSet, taskParams.payload);

        bytes32 taskHash = keccak256(abi.encode(globalTaskCount, address(this), block.chainid, taskParams));
        globalTaskCount = globalTaskCount + 1;

        AvsConfig memory memAvsConfig = avsConfigs[taskParams.executorOperatorSet.avs];

        tasks[taskHash] = Task(
            msg.sender,
            block.timestamp.toUint96(),
            TaskStatus.Created,
            taskParams.executorOperatorSet.avs,
            taskParams.executorOperatorSet.id,
            memAvsConfig.aggregatorOperatorSetId,
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
        taskConfig.taskHook.validatePostTaskCreation(taskHash);

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

    function submitResult(
        bytes32 taskHash,
        IBN254CertificateVerifier.BN254Certificate memory cert,
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
        bool isCertificateValid = IBN254CertificateVerifier(task.executorOperatorSetTaskConfig.certificateVerifier)
            .verifyCertificateProportion(cert, totalStakeProportionThresholds);

        require(isCertificateValid, CertificateVerificationFailed());

        task.status = TaskStatus.Verified;
        task.result = result;

        // TODO: Check what happens if we re-ennter the other state transition functions.
        // Task result submission checks:
        // 1. AVS can validate the task result, params and certificate.
        // 2. It can update hook storage for task lifecycle if needed.
        task.executorOperatorSetTaskConfig.taskHook.validateTaskResultSubmission(taskHash, cert);

        emit TaskVerified(msg.sender, taskHash, task.avs, task.executorOperatorSetId, task.result);
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
                && block.timestamp > (task.creationTime + task.executorOperatorSetTaskConfig.taskSLA)
        ) {
            return TaskStatus.Expired;
        }
        return task.status;
    }

    function _registerAvs(address avs, bool isRegistered) internal {
        isAvsRegistered[avs] = isRegistered;
        emit AvsRegistered(msg.sender, avs, isRegistered);
    }

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */
    function getAvsConfig(
        address avs
    ) external view returns (AvsConfig memory) {
        return avsConfigs[avs];
    }

    function getExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet
    ) external view returns (ExecutorOperatorSetTaskConfig memory) {
        return executorOperatorSetTaskConfigs[operatorSet.key()];
    }

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
            task.aggregatorOperatorSetId,
            task.refundCollector,
            task.avsFee,
            task.feeSplit,
            task.executorOperatorSetTaskConfig,
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
