// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {IAVSTaskHook} from "../avs/l2/IAVSTaskHook.sol";
import {IBN254CertificateVerifier} from "../avs/l2/IBN254CertificateVerifier.sol";

interface ITaskMailboxTypes {
    // TODO: Pack Storage efficiently.
    struct AvsConfig {
        uint32 aggregatorOperatorSetId; // TODO: Add avs address too: Any AVS can be an aggregator.
        uint32[] executorOperatorSetIds;
    }

    // TODO: Pack Storage efficiently.
    // TODO: We need to support proportional, nominal, none and custom verifications.
    // TODO: We also need to support BN254, ECDSA, BLS and custom curves.
    struct ExecutorOperatorSetTaskConfig {
        address certificateVerifier;
        IAVSTaskHook taskHook;
        IERC20 feeToken;
        address feeCollector;
        uint96 taskSLA;
        uint16 stakeProportionThreshold;
        bytes taskMetadata;
    }

    struct TaskParams {
        address refundCollector;
        uint96 avsFee;
        OperatorSet executorOperatorSet;
        bytes payload;
    }

    // TODO: `Created` status cannot be enum value 0 since that is the default value. Figure out how to handle this.
    enum TaskStatus {
        Created,
        Canceled,
        Verified,
        Expired
    }

    // TODO: Pack Storage efficiently.
    struct Task {
        address creator;
        uint96 creationTime;
        TaskStatus status;
        address avs;
        uint32 executorOperatorSetId;
        uint32 aggregatorOperatorSetId;
        address refundCollector;
        uint96 avsFee;
        uint16 feeSplit;
        ExecutorOperatorSetTaskConfig executorOperatorSetTaskConfig;
        bytes payload;
        bytes result;
    }
}

interface ITaskMailboxErrors is ITaskMailboxTypes {
    /// @dev Thrown when an AVS is not registered
    error AvsNotRegistered();
    /// @dev Thrown when a certificate verification fails
    error CertificateVerificationFailed();
    /// @dev Thrown when an executor operator set id is already in the set
    error DuplicateExecutorOperatorSetId();
    /// @dev Thrown when an executor operator set is not registered
    error ExecutorOperatorSetNotRegistered();
    /// @dev Thrown when an executor operator set task config is not set
    error ExecutorOperatorSetTaskConfigNotSet();
    /// @dev Thrown when an input address is zero
    error InvalidAddressZero();
    /// @dev Thrown when an aggregator operator set id is also an executor operator set id
    error InvalidAggregatorOperatorSetId();
    /// @dev Thrown when a task creator is invalid
    error InvalidTaskCreator();
    /// @dev Thrown when a task status is invalid
    error InvalidTaskStatus(TaskStatus expected, TaskStatus actual);
    /// @dev Thrown when a payload is empty
    error PayloadIsEmpty();
    /// @dev Thrown when a task SLA is zero
    error TaskSLAIsZero();
    /// @dev Thrown when a timestamp is at creation
    error TimestampAtCreation();
}

interface ITaskMailboxEvents is ITaskMailboxTypes {
    event AvsRegistered(address indexed caller, address indexed avs, bool isRegistered);

    event AvsConfigSet(
        address indexed caller, address indexed avs, uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds
    );

    event ExecutorOperatorSetTaskConfigSet(
        address indexed caller,
        address indexed avs,
        uint32 indexed executorOperatorSetId,
        ExecutorOperatorSetTaskConfig config
    );

    event TaskCreated(
        address indexed creator,
        bytes32 indexed taskHash,
        address indexed avs,
        uint32 executorOperatorSetId,
        address refundCollector,
        uint96 avsFee,
        uint256 taskDeadline,
        bytes payload
    );

    event TaskCanceled(
        address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId
    );

    event TaskVerified(
        address indexed aggregator,
        bytes32 indexed taskHash,
        address indexed avs,
        uint32 executorOperatorSetId,
        bytes result
    );
}

interface ITaskMailbox is ITaskMailboxErrors, ITaskMailboxEvents {
    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */
    function registerAvs(address avs, bool isRegistered) external;

    function setAvsConfig(address avs, AvsConfig memory config) external;

    function setExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet,
        ExecutorOperatorSetTaskConfig memory config
    ) external;

    function createTask(
        TaskParams memory taskParams
    ) external returns (bytes32 taskHash);

    function cancelTask(
        bytes32 taskHash
    ) external;

    function submitResult(
        bytes32 taskHash,
        IBN254CertificateVerifier.BN254Certificate memory cert,
        bytes memory result
    ) external;

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */
    function isAvsRegistered(
        address avs
    ) external view returns (bool);

    function isExecutorOperatorSetRegistered(
        bytes32 operatorSetKey
    ) external view returns (bool);

    function getAvsConfig(
        address avs
    ) external view returns (AvsConfig memory);

    function getExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet
    ) external view returns (ExecutorOperatorSetTaskConfig memory);

    function getTaskInfo(
        bytes32 taskHash
    ) external view returns (Task memory);

    function getTaskStatus(
        bytes32 taskHash
    ) external view returns (TaskStatus);

    function getTaskResult(
        bytes32 taskHash
    ) external view returns (bytes memory);
}
