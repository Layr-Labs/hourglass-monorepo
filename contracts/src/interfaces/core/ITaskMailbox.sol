// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IBN254CertificateVerifierTypes} from
    "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {IECDSACertificateVerifierTypes} from
    "@eigenlayer-contracts/src/contracts/interfaces/IECDSACertificateVerifier.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {IAVSTaskHook} from "../avs/l2/IAVSTaskHook.sol";

/**
 * @title ITaskMailboxTypes
 * @notice Interface defining the type structures used in the TaskMailbox
 */
interface ITaskMailboxTypes {
    /**
     * @notice Configuration for certificate verifiers
     * @param curveType The curve type for the verifier
     * @param verifier Address of the certificate verifier contract
     */
    struct CertificateVerifierConfig {
        IKeyRegistrarTypes.CurveType curveType;
        address verifier;
    }

    /**
     * @notice Configuration for the executor operator set
     * @param curveType The curve type used for signature verification
     * @param taskHook Address of the AVS task hook contract
     * @param feeToken ERC20 token used for task fees
     * @param feeCollector Address to receive AVS fees
     * @param taskSLA Time (in seconds) within which the task must be completed
     * @param stakeProportionThreshold Minimum proportion of executor operator set stake required to certify task execution (in basis points)
     * @param taskMetadata Additional metadata for task execution
     */
    struct ExecutorOperatorSetTaskConfig {
        // TODO: Pack storage efficiently.
        // TODO: We need to support proportional, nominal, none and custom verifications.
        IKeyRegistrarTypes.CurveType curveType;
        IAVSTaskHook taskHook;
        IERC20 feeToken;
        address feeCollector;
        uint96 taskSLA;
        uint16 stakeProportionThreshold;
        bytes taskMetadata;
    }

    /**
     * @notice Parameters for creating a new task
     * @param refundCollector Address to receive refunds if task is not completed
     * @param avsFee Fee paid to the AVS for processing the task
     * @param executorOperatorSet The operator set that will execute the task
     * @param payload Task payload
     */
    struct TaskParams {
        address refundCollector;
        uint96 avsFee;
        OperatorSet executorOperatorSet;
        bytes payload;
    }

    /**
     * @notice Status of a task in the system
     */
    enum TaskStatus {
        NONE, // 0 - Default value for uninitialized tasks
        CREATED, // 1 - Task has been created
        VERIFIED, // 2 - Task has been verified
        EXPIRED // 3 - Task has expired

    }

    /**
     * @notice Complete task information
     * @param creator Address that created the task
     * @param creationTime Block timestamp when task was created
     * @param status Current status of the task
     * @param avs Address of the AVS handling the task
     * @param executorOperatorSetId ID of the operator set executing the task
     * @param refundCollector Address to receive refunds
     * @param avsFee Fee paid to the AVS
     * @param feeSplit Percentage split of fees taken by the TaskMailbox
     * @param executorOperatorSetTaskConfig Configuration for executor operator set task execution
     * @param payload Task payload
     * @param result Task execution result data
     */
    struct Task {
        // TODO: Pack storage efficiently.
        address creator;
        uint96 creationTime;
        TaskStatus status;
        address avs;
        uint32 executorOperatorSetId;
        address refundCollector;
        uint96 avsFee;
        uint16 feeSplit;
        ExecutorOperatorSetTaskConfig executorOperatorSetTaskConfig;
        bytes payload;
        bytes result;
    }
}

/**
 * @title ITaskMailboxErrors
 * @notice Interface defining errors that can be thrown by the TaskMailbox
 */
interface ITaskMailboxErrors is ITaskMailboxTypes {
    /// @notice Thrown when a certificate verification fails
    error CertificateVerificationFailed();

    /// @notice Thrown when an executor operator set is not registered
    error ExecutorOperatorSetNotRegistered();

    /// @notice Thrown when an executor operator set task config is not set
    error ExecutorOperatorSetTaskConfigNotSet();

    /// @notice Thrown when an input address is zero
    error InvalidAddressZero();

    /// @notice Thrown when an invalid curve type is provided
    error InvalidCurveType();

    /// @notice Thrown when a task creator is invalid
    error InvalidTaskCreator();

    /// @notice Thrown when a task status is invalid
    /// @param expected The expected task status
    /// @param actual The actual task status
    error InvalidTaskStatus(TaskStatus expected, TaskStatus actual);

    /// @notice Thrown when a payload is empty
    error PayloadIsEmpty();

    /// @notice Thrown when a task SLA is zero
    error TaskSLAIsZero();

    /// @notice Thrown when a timestamp is at creation
    error TimestampAtCreation();

    /// @notice Thrown when an operator set owner is invalid
    error InvalidOperatorSetOwner();
}

/**
 * @title ITaskMailboxEvents
 * @notice Interface defining events emitted by the TaskMailbox
 */
interface ITaskMailboxEvents is ITaskMailboxTypes {
    /**
     * @notice Emitted when a certificate verifier is set
     * @param curveType The curve type for the verifier
     * @param certificateVerifier Address of the certificate verifier
     */
    event CertificateVerifierSet(IKeyRegistrarTypes.CurveType indexed curveType, address indexed certificateVerifier);

    /**
     * @notice Emitted when an executor operator set is registered
     * @param caller Address that called the registration function
     * @param avs Address of the AVS being registered
     * @param executorOperatorSetId ID of the executor operator set
     * @param isRegistered Whether the operator set is registered
     */
    event ExecutorOperatorSetRegistered(
        address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, bool isRegistered
    );

    /**
     * @notice Emitted when an executor operator set task configuration is set
     * @param caller Address that called the configuration function
     * @param avs Address of the AVS being configured
     * @param executorOperatorSetId ID of the executor operator set
     * @param config The task configuration for the executor operator set
     */
    event ExecutorOperatorSetTaskConfigSet(
        address indexed caller,
        address indexed avs,
        uint32 indexed executorOperatorSetId,
        ExecutorOperatorSetTaskConfig config
    );

    /**
     * @notice Emitted when a new task is created
     * @param creator Address that created the task
     * @param taskHash Unique identifier of the task
     * @param avs Address of the AVS handling the task
     * @param executorOperatorSetId ID of the executor operator set
     * @param refundCollector Address to receive refunds
     * @param avsFee Fee paid to the AVS
     * @param taskDeadline Timestamp by which the task must be completed
     * @param payload Task payload
     */
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

    /**
     * @notice Emitted when a task is verified
     * @param aggregator Address that submitted the verification
     * @param taskHash Unique identifier of the task
     * @param avs Address of the AVS handling the task
     * @param executorOperatorSetId ID of the executor operator set
     * @param result Task execution result data
     */
    event TaskVerified(
        address indexed aggregator,
        bytes32 indexed taskHash,
        address indexed avs,
        uint32 executorOperatorSetId,
        bytes result
    );
}

/**
 * @title ITaskMailbox
 * @author Layr Labs, Inc.
 * @notice Interface for the TaskMailbox contract.
 */
interface ITaskMailbox is ITaskMailboxErrors, ITaskMailboxEvents {
    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */

    /**
     * @notice Sets a certificate verifier for a specific curve type
     * @param curveType The curve type for the verifier
     * @param certificateVerifier Address of the certificate verifier
     */
    function setCertificateVerifier(IKeyRegistrarTypes.CurveType curveType, address certificateVerifier) external;

    /**
     * @notice Sets the task configuration for an executor operator set
     * @param operatorSet The operator set to configure
     * @param config Task configuration for the operator set
     */
    function setExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet,
        ExecutorOperatorSetTaskConfig memory config
    ) external;

    /**
     * @notice Registers an executor operator set with the TaskMailbox
     * @param operatorSet The operator set to register
     * @param isRegistered Whether the operator set is registered
     */
    function registerExecutorOperatorSet(OperatorSet memory operatorSet, bool isRegistered) external;

    /**
     * @notice Creates a new task
     * @param taskParams Parameters for the task
     * @return taskHash Unique identifier of the created task
     */
    function createTask(
        TaskParams memory taskParams
    ) external returns (bytes32 taskHash);

    /**
     * @notice Submits the result of a task execution
     * @param taskHash Unique identifier of the task
     * @param cert Certificate proving the validity of the result
     * @param result Task execution result data
     */
    function submitResult(bytes32 taskHash, bytes memory cert, bytes memory result) external;

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */

    /**
     * @notice Checks if an executor operator set is registered
     * @param operatorSetKey Key of the operator set to check
     * @return True if the executor operator set is registered, false otherwise
     */
    function isExecutorOperatorSetRegistered(
        bytes32 operatorSetKey
    ) external view returns (bool);

    /**
     * @notice Gets the certificate verifier for a specific curve type
     * @param curveType The curve type to get the verifier for
     * @return Address of the certificate verifier
     */
    function getCertificateVerifier(
        IKeyRegistrarTypes.CurveType curveType
    ) external view returns (address);

    /**
     * @notice Gets the task configuration for an executor operator set
     * @param operatorSet The operator set to get configuration for
     * @return Task configuration for the operator set
     */
    function getExecutorOperatorSetTaskConfig(
        OperatorSet memory operatorSet
    ) external view returns (ExecutorOperatorSetTaskConfig memory);

    /**
     * @notice Gets complete information about a task
     * @param taskHash Unique identifier of the task
     * @return Complete task information
     */
    function getTaskInfo(
        bytes32 taskHash
    ) external view returns (Task memory);

    /**
     * @notice Gets the current status of a task
     * @param taskHash Unique identifier of the task
     * @return Current status of the task
     */
    function getTaskStatus(
        bytes32 taskHash
    ) external view returns (TaskStatus);

    /**
     * @notice Gets the result of a verified task
     * @param taskHash Unique identifier of the task
     * @return Result data of the task
     */
    function getTaskResult(
        bytes32 taskHash
    ) external view returns (bytes memory);

    /**
     * @notice Gets the bytes of a BN254 certificate
     * @param cert The certificate to get the bytes of
     * @return The bytes of the certificate
     */
    function getBN254CertificateBytes(
        IBN254CertificateVerifierTypes.BN254Certificate memory cert
    ) external pure returns (bytes memory);

    /**
     * @notice Gets the bytes of a ECDSA certificate
     * @param cert The certificate to get the bytes of
     * @return The bytes of the certificate
     */
    function getECDSACertificateBytes(
        IECDSACertificateVerifierTypes.ECDSACertificate memory cert
    ) external pure returns (bytes memory);
}
