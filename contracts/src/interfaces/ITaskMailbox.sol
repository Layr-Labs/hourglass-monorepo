// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {
    OperatorSet,
    OperatorSetLib
} from "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IERC20} from "@eigenlayer-middleware/lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";

import {IAVSTaskHook} from "src/interfaces/IAVSTaskHook.sol";
import {IBN254CertificateVerifier} from "src/interfaces/IBN254CertificateVerifier.sol";

interface ITaskMailboxTypes {
    // TODO: Pack Storage efficiently.
    // TODO: We need to support proportional, nominal, none and custom verifications.
    // TODO: We also need to support BN254, ECDSA, BLS and custom curves.
    struct OperatorSetTaskConfig {
        address certificateVerifier;
        IAVSTaskHook taskHook;
        address aggregator;
        IERC20 feeToken;
        address feeCollector;
        uint96 taskSLA;
        uint16 stakeProportionThreshold;
        bytes taskMetadata;
    }

    struct TaskParams {
        address refundCollector;
        uint96 avsFee;
        OperatorSet operatorSet;
        bytes payload;
    }

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
        OperatorSet operatorSet;
        address refundCollector;
        uint96 avsFee;
        uint16 feeSplit;
        OperatorSetTaskConfig operatorSetTaskConfig;
        bytes payload;
        bytes result;
    }
}

interface ITaskMailboxErrors is ITaskMailboxTypes {
    /// @dev Thrown when a certificate verification fails
    error CertificateVerificationFailed();
    /// @dev Thrown when an input address is zero
    error InvalidAddressZero();
    /// @dev Thrown when an aggregator is invalid
    error InvalidTaskAggregator();
    /// @dev Thrown when a task creator is invalid
    error InvalidTaskCreator();
    /// @dev Thrown when a task status is invalid
    error InvalidTaskStatus(TaskStatus expected, TaskStatus actual);
    /// @dev Thrown when an operator set is not registered to the task mailbox
    error OperatorSetNotRegistered();
    /// @dev Thrown when an operator set task config is not set
    error OperatorSetTaskConfigNotSet();
    /// @dev Thrown when a payload is empty
    error PayloadIsEmpty();
    /// @dev Thrown when a task SLA is zero
    error TaskSLAIsZero();
    /// @dev Thrown when a timestamp is at creation
    error TimestampAtCreation();
}

interface ITaskMailboxEvents is ITaskMailboxTypes {
    event OperatorSetRegistered(
        address indexed caller, address indexed avs, uint32 indexed operatorSetId, bool isRegistered
    );

    event OperatorSetTaskConfigSet(
        address indexed caller, address indexed avs, uint32 indexed operatorSetId, OperatorSetTaskConfig config
    );

    event TaskCreated(
        address indexed creator,
        bytes32 indexed taskHash,
        address indexed avs,
        uint32 operatorSetId,
        address refundCollector,
        uint96 avsFee,
        uint256 taskDeadline,
        bytes payload
    );

    event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId);

    event TaskVerified(
        address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, bytes result
    );
}

interface ITaskMailbox is ITaskMailboxErrors, ITaskMailboxEvents {
    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */
    function registerOperatorSet(OperatorSet memory operatorSet, bool isRegistered) external;

    function setOperatorSetTaskConfig(OperatorSet memory operatorSet, OperatorSetTaskConfig memory config) external;

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
    function getOperatorSetTaskConfig(
        OperatorSet memory operatorSet
    ) external view returns (OperatorSetTaskConfig memory);

    function getTaskInfo(
        bytes32 taskHash
    ) external view returns (Task memory);

    function getTaskStatus(
        bytes32 taskHash
    ) external view returns (TaskStatus);
}
