// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAVSRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";
import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";

/**
 * @title ITaskAVSRegistrarTypes
 * @notice Interface defining data types used in the TaskAVSRegistrar
 */
interface ITaskAVSRegistrarTypes {
    /**
     * @notice Parameters required when registering a new BLS public key.
     * @dev Contains the registration signature and both G1/G2 public key components.
     * @param pubkeyRegistrationSignature Registration message signed by operator's private key to prove ownership.
     * @param pubkeyG1 The operator's public key in G1 group format.
     * @param pubkeyG2 The operator's public key in G2 group format, must correspond to the same private key as pubkeyG1.
     */
    struct PubkeyRegistrationParams {
        BN254.G1Point pubkeyRegistrationSignature;
        BN254.G1Point pubkeyG1;
        BN254.G2Point pubkeyG2;
    }

    /**
     * @notice Parameters required when registering a new operator.
     * @dev Contains the operator's socket (url:port) and BLS public key.
     * @param socket The operator's socket address.
     * @param pubkeyRegistrationParams Parameters required when registering a new BLS public key.
     */
    struct OperatorRegistrationParams {
        string socket;
        PubkeyRegistrationParams pubkeyRegistrationParams;
    }

    /**
     * @notice Information about a BLS public key.
     * @param pubkeyG1 The operator's public key in G1 group format.
     * @param pubkeyG2 The operator's public key in G2 group format, must correspond to the same private key as pubkeyG1.
     * @param pubkeyHash The unique identifier of the operator's BLS public key.
     */
    struct PubkeyInfo {
        BN254.G1Point pubkeyG1;
        BN254.G2Point pubkeyG2;
        bytes32 pubkeyHash;
    }

    /**
     * @notice Information about a BLS public key and its corresponding socket.
     * @param pubkeyInfo The information about the BLS public key.
     * @param socket The socket address of the operator.
     */
    struct PubkeyInfoAndSocket {
        PubkeyInfo pubkeyInfo;
        string socket;
    }
}

/**
 * @title ITaskAVSRegistrarErrors
 * @notice Interface defining errors that can be thrown by the TaskAVSRegistrar
 */
interface ITaskAVSRegistrarErrors is ITaskAVSRegistrarTypes {
    /// @notice Thrown when the provided AVS address does not match the expected one.
    error InvalidAVS();
    /// @notice Thrown when the caller is not the AllocationManager
    error OnlyAllocationManager();
    /// @notice Thrown when the operator is already registered.
    error OperatorAlreadyRegistered();
    /// @notice Thrown when the BLS public key is already registered.
    error BLSPubkeyAlreadyRegistered();
    /// @notice Thrown when the provided BLS signature is invalid.
    error InvalidBLSSignatureOrPrivateKey();
    /// @notice Thrown when the operator is not registered.
    error OperatorNotRegistered();
    /// @notice Thrown when the provided pubkey hash is zero.
    error ZeroPubKey();
}

/**
 * @title ITaskAVSRegistrarEvents
 * @notice Interface defining events emitted by the TaskAVSRegistrar
 */
interface ITaskAVSRegistrarEvents is ITaskAVSRegistrarTypes {
    /**
     * @notice Emitted when a new BLS public key is registered.
     * @param operator The address of the operator registering the pubkey.
     * @param pubkeyHash The hash of the registered public key.
     * @param pubkeyG1 The registered G1 public key.
     * @param pubkeyG2 The registered G2 public key.
     */
    event NewPubkeyRegistration(
        address indexed operator, bytes32 indexed pubkeyHash, BN254.G1Point pubkeyG1, BN254.G2Point pubkeyG2
    );

    /**
     * @notice Emitted when an operator's socket address is updated.
     * @param operator The address of the operator.
     * @param pubkeyHash The hash of the operator's public key.
     * @param socket The new socket address.
     */
    event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket);

    /**
     * @notice Emitted when the APK for an operatorSet is updated.
     * @param operator The address of the operator causing the update.
     * @param pubkeyHash The hash of the operator's public key.
     * @param operatorSetId The ID of the operatorSet whose APK was updated.
     * @param apk The new aggregate public key.
     */
    event OperatorSetApkUpdated(
        address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, BN254.G1Point apk
    );
}

/**
 * @title ITaskAVSRegistrar
 * @notice Interface for the TaskAVSRegistrar contract that handles operator registration and BLS pubkey management
 */
interface ITaskAVSRegistrar is ITaskAVSRegistrarErrors, ITaskAVSRegistrarEvents, IAVSRegistrar {
    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */

    /**
     * @notice Updates an operator's socket address
     * @param socket The new socket address to set for the operator
     * @dev Only registered operators can update their socket
     */
    function updateOperatorSocket(
        string memory socket
    ) external;

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */

    /**
     * @notice Gets the aggregate public key for an operator set
     * @param operatorSetId The ID of the operator set
     * @return The aggregate public key in G1 format
     */
    // TODO: Update operatorSetId to uint32
    function getApk(
        uint8 operatorSetId
    ) external view returns (BN254.G1Point memory);

    /**
     * @notice Gets the registered public key information for an operator
     * @param operator The address of the operator
     * @return Public key information including G1, G2, and hash
     */
    function getRegisteredPubkeyInfo(
        address operator
    ) external view returns (PubkeyInfo memory);

    /**
     * @notice Gets the registered G1 public key and hash for an operator
     * @param operator The address of the operator
     * @return The operator's G1 public key and its hash
     * @dev This function is kept for backwards compatibility
     */
    function getRegisteredPubkey(
        address operator
    ) external view returns (BN254.G1Point memory, bytes32);

    /**
     * @notice Gets the G2 public key for an operator
     * @param operator The address of the operator
     * @return The operator's G2 public key
     */
    function getOperatorPubkeyG2(
        address operator
    ) external view returns (BN254.G2Point memory);

    /**
     * @notice Gets the operator address associated with a public key hash
     * @param pubkeyHash The hash of the public key
     * @return The address of the operator who registered the public key
     */
    function getOperatorFromPubkeyHash(
        bytes32 pubkeyHash
    ) external view returns (address);

    /**
     * @notice Gets the public key hash for an operator
     * @param operator The address of the operator
     * @return The hash of the operator's public key
     */
    function getOperatorPubkeyHash(
        address operator
    ) external view returns (bytes32);

    /**
     * @notice Returns the message hash that an operator must sign to register their BLS public key
     * @param operator The address of the operator registering their BLS public key
     * @return The message hash in G1 point format
     */
    function pubkeyRegistrationMessageHash(
        address operator
    ) external view returns (BN254.G1Point memory);

    /**
     * @notice Calculates the message hash that an operator must sign to register their BLS public key
     * @param operator The address of the operator registering their BLS public key
     * @return The raw message hash bytes
     */
    function calculatePubkeyRegistrationMessageHash(
        address operator
    ) external view returns (bytes32);

    /**
     * @notice Gets the socket address for an operator by public key hash
     * @param pubkeyHash The hash of the operator's public key
     * @return The socket address of the operator
     */
    function getOperatorSocketByPubkeyHash(
        bytes32 pubkeyHash
    ) external view returns (string memory);

    /**
     * @notice Gets the socket address for an operator by address
     * @param operator The address of the operator
     * @return The socket address of the operator
     */
    function getOperatorSocketByOperator(
        address operator
    ) external view returns (string memory);

    /**
     * @notice Gets public key information and socket addresses for multiple operators
     * @param operators Array of operator addresses
     * @return Array of public key information and socket addresses
     */
    function getBatchOperatorPubkeyInfoAndSocket(
        address[] calldata operators
    ) external view returns (PubkeyInfoAndSocket[] memory);

    /**
     * @notice Packs operator registration parameters into a single bytes object
     * @param socket The operator's socket address
     * @param pubkeyRegistrationParams The operator's public key registration parameters
     * @return Packed bytes representing the registration payload
     */
    function packRegisterPayload(
        string memory socket,
        PubkeyRegistrationParams memory pubkeyRegistrationParams
    ) external pure returns (bytes memory);
}
