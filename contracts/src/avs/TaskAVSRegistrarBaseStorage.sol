// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";

import {ITaskAVSRegistrar} from "../interfaces/avs/l1/ITaskAVSRegistrar.sol";

/**
 * @title TaskAVSRegistrarBaseStorage
 * @notice Storage contract for the TaskAVSRegistrar system
 * @dev Contains all state variables used by the TaskAVSRegistrar contract
 */
abstract contract TaskAVSRegistrarBaseStorage is ITaskAVSRegistrar {
    /// @notice The avs address for this AVS (used for UAM integration in EigenLayer)
    /// @dev NOTE: Updating this value will break existing OperatorSets and UAM integration.
    /// This value should only be set once.
    address public immutable AVS;

    /// @notice Returns the hash of the zero pubkey aka BN254.G1Point(0,0)
    /// @dev Used to validate that pubkeys are not zero
    bytes32 internal constant ZERO_PK_HASH = hex"ad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5";

    /// @notice The EIP-712 typehash used for registering BLS public keys
    bytes32 public constant PUBKEY_REGISTRATION_TYPEHASH = keccak256("BN254PubkeyRegistration(address operator)");

    /// @notice The AllocationManager that tracks OperatorSets and Slashing in EigenLayer
    IAllocationManager public immutable ALLOCATION_MANAGER;

    /// @notice Mapping from operator address to pubkey hash
    mapping(address operator => bytes32 pubkeyHash) public operatorToPubkeyHash;

    /// @notice Mapping from pubkey hash to operator address
    mapping(bytes32 pubkeyHash => address operator) public pubkeyHashToOperator;

    /// @notice Mapping from operator address to G1 pubkey
    mapping(address operator => BN254.G1Point pubkeyG1) public operatorToPubkey;

    /// @notice Mapping from operator address to G2 pubkey
    mapping(address operator => BN254.G2Point) internal operatorToPubkeyG2;

    /// @notice Mapping from pubkey hash to operator socket
    mapping(bytes32 pubkeyHash => string socket) public pubkeyHashToSocket;

    /// @notice Mapping from operator address to operator socket
    mapping(address operator => string socket) public operatorToSocket;

    /// @notice Current Aggregate Public Key (APK) of each OperatorSet
    mapping(uint32 operatorSetId => BN254.G1Point apk) public currentApk;

    /**
     * @notice Constructor for TaskAVSRegistrarBaseStorage
     * @param avs The address of the AVS
     * @param allocationManager The AllocationManager contract address
     */
    constructor(address avs, IAllocationManager allocationManager) {
        AVS = avs;
        ALLOCATION_MANAGER = allocationManager;
    }
}
