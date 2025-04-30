// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {ITaskAVSRegistrar} from "src/interfaces/ITaskAVSRegistrar.sol";
import {IAllocationManager} from
    "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";

abstract contract TaskAVSRegistrarStorage is ITaskAVSRegistrar {
    /// @notice The avs address for this AVS (used for UAM integration in EigenLayer)
    /// @dev NOTE: Updating this value will break existing OperatorSets and UAM integration.
    /// This value should only be set once.
    address public immutable AVS;

    /// @notice The AllocationManager that tracks OperatorSets and Slashing in EigenLayer
    IAllocationManager public immutable ALLOCATION_MANAGER;

    constructor(address avs, IAllocationManager allocationManager) {
        AVS = avs;
        ALLOCATION_MANAGER = allocationManager;
    }
}
