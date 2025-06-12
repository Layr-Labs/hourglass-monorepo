// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {AVSRegistrarWithSocket} from
    "@eigenlayer-middleware/src/middlewareV2/registrar/presets/AVSRegistrarWithSocket.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IKeyRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";

/**
 * @title TaskAVSRegistrarBaseStorage
 * @author Layr Labs, Inc.
 * @notice Storage contract for TaskAVSRegistrarBase that extends AVSRegistrarWithSocket
 * @dev This contract extends AVSRegistrarWithSocket which already includes AVSRegistrar and SocketRegistry storage
 */
abstract contract TaskAVSRegistrarBaseStorage is AVSRegistrarWithSocket {
    /**
     * @notice Constructor for TaskAVSRegistrarBaseStorage
     * @param _avs The address of the AVS
     * @param _allocationManager The AllocationManager contract address
     * @param _keyRegistrar The KeyRegistrar contract address
     */
    constructor(
        address _avs,
        IAllocationManager _allocationManager,
        IKeyRegistrar _keyRegistrar
    ) AVSRegistrarWithSocket(_avs, _allocationManager, _keyRegistrar) {}

    /**
     * @dev This empty reserved space is put in place to allow future versions to add new
     * variables without shifting down storage in the inheritance chain.
     * See https://docs.openzeppelin.com/contracts/4.x/upgradeable#storage_gaps
     */
    uint256[50] private __gap;
}
