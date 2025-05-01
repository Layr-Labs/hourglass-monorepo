// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAVSRegistrar} from "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";
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
}

interface ITaskAVSRegistrarErrors is ITaskAVSRegistrarTypes {
    /// @notice Thrown when the provided AVS address does not match the expected one.
    error InvalidAVS();
    /// @notice Thrown when the caller is not the AllocationManager
    error OnlyAllocationManager();
}

interface ITaskAVSRegistrarEvents is ITaskAVSRegistrarTypes {}

interface ITaskAVSRegistrar is ITaskAVSRegistrarErrors, ITaskAVSRegistrarEvents, IAVSRegistrar {}
