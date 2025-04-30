// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAVSRegistrar} from "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";

interface ITaskAVSRegistrarTypes {}

interface ITaskAVSRegistrarErrors is ITaskAVSRegistrarTypes {
    /// @notice Thrown when the provided AVS address does not match the expected one.
    error InvalidAVS();
    /// @notice Thrown when the caller is not the AllocationManager
    error OnlyAllocationManager();
}

interface ITaskAVSRegistrarEvents is ITaskAVSRegistrarTypes {}

interface ITaskAVSRegistrar is ITaskAVSRegistrarErrors, ITaskAVSRegistrarEvents, IAVSRegistrar {}
