// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAVSRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";
import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";

interface ITaskAVSRegistrarTypes {
    /// @notice Parameters required when registering a new BLS public key.
    /// @dev Contains the registration signature and both G1/G2 public key components.
    /// @param pubkeyRegistrationSignature Registration message signed by operator's private key to prove ownership.
    /// @param pubkeyG1 The operator's public key in G1 group format.
    /// @param pubkeyG2 The operator's public key in G2 group format, must correspond to the same private key as pubkeyG1.
    struct PubkeyRegistrationParams {
        BN254.G1Point pubkeyRegistrationSignature;
        BN254.G1Point pubkeyG1;
        BN254.G2Point pubkeyG2;
    }

    /// @notice Parameters required when registering a new operator.
    /// @dev Contains the operator's socket (url:port) and BLS public key.
    /// @param socket The operator's socket.
    /// @param pubkeyRegistrationParams Parameters required when registering a new BLS public key.
    struct OperatorRegistrationParams {
        string socket;
        PubkeyRegistrationParams pubkeyRegistrationParams;
    }

    /// @notice Information about a BLS public key.
    /// @param pubkeyG1 The operator's public key in G1 group format.
    /// @param pubkeyG2 The operator's public key in G2 group format, must correspond to the same private key as pubkeyG1.
    /// @param pubkeyHash The unique identifier of the operator's BLS public key.
    struct PubkeyInfo {
        BN254.G1Point pubkeyG1;
        BN254.G2Point pubkeyG2;
        bytes32 pubkeyHash;
    }

    /// @notice Information about a BLS public key and its corresponding socket.
    /// @param pubkeyInfo The information about the BLS public key.
    /// @param socket The socket address of the operator.
    struct PubkeyInfoAndSocket {
        PubkeyInfo pubkeyInfo;
        string socket;
    }
}

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

interface ITaskAVSRegistrarEvents is ITaskAVSRegistrarTypes {
    /// @notice Emitted when a new BLS public key is registered.
    event NewPubkeyRegistration(
        address indexed operator, bytes32 indexed pubkeyHash, BN254.G1Point pubkeyG1, BN254.G2Point pubkeyG2
    );

    /// @notice Emitted when an operator's socket address is updated.
    event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket);

    /// @notice Emitted when the APK for an operatorSet is updated.
    event OperatorSetApkUpdated(
        address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, BN254.G1Point apk
    );
}

interface ITaskAVSRegistrar is ITaskAVSRegistrarErrors, ITaskAVSRegistrarEvents, IAVSRegistrar {
    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */
    function updateOperatorSocket(
        string memory socket
    ) external;

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */

    // TODO: Update operatorSetId to uint32
    function getApk(
        uint8 operatorSetId
    ) external view returns (BN254.G1Point memory);

    function getRegisteredPubkeyInfo(
        address operator
    ) external view returns (PubkeyInfo memory);

    function getRegisteredPubkey(
        address operator
    ) external view returns (BN254.G1Point memory, bytes32);

    function getOperatorPubkeyG2(
        address operator
    ) external view returns (BN254.G2Point memory);

    function getOperatorFromPubkeyHash(
        bytes32 pubkeyHash
    ) external view returns (address);

    function getOperatorPubkeyHash(
        address operator
    ) external view returns (bytes32);

    function pubkeyRegistrationMessageHash(
        address operator
    ) external view returns (BN254.G1Point memory);

    function calculatePubkeyRegistrationMessageHash(
        address operator
    ) external view returns (bytes32);

    function getOperatorSocketByPubkeyHash(
        bytes32 pubkeyHash
    ) external view returns (string memory);

    function getOperatorSocketByOperator(
        address operator
    ) external view returns (string memory);

    function getBatchOperatorPubkeyInfoAndSocket(
        address[] calldata operators
    ) external view returns (PubkeyInfoAndSocket[] memory);

    function packRegisterPayload(
        string memory socket,
        PubkeyRegistrationParams memory pubkeyRegistrationParams
    ) external pure returns (bytes memory);
}
