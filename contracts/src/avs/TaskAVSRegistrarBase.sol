// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {AVSRegistrarWithSocket} from "@eigenlayer-middleware/src/middlewareV2/registrar/presets/AVSRegistrarWithSocket.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IKeyRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";

/**
 * @title TaskAVSRegistrarBase
 * @author Layr Labs, Inc.
 * @notice Abstract contract that extends AVSRegistrarWithSocket for task-based AVSs
 * @dev This is a minimal wrapper around AVSRegistrarWithSocket
 */
abstract contract TaskAVSRegistrarBase is AVSRegistrarWithSocket {
    /**
     * @notice Constructs the TaskAVSRegistrarBase contract
     * @param _avs The address of the AVS
     * @param _allocationManager The AllocationManager contract address
     * @param _keyRegistrar The KeyRegistrar contract address
     */
    constructor(
        address _avs,
        IAllocationManager _allocationManager,
        IKeyRegistrar _keyRegistrar
    ) AVSRegistrarWithSocket(_avs, _allocationManager, _keyRegistrar) {}
} 