// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {TaskAVSRegistrarStorage} from "src/avs/l1-contracts/TaskAVSRegistrarStorage.sol";
import {IAllocationManager} from
    "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";

contract TaskAVSRegistrar is TaskAVSRegistrarStorage {
    /// @notice Modifier to ensure the caller is the AllocationManager
    modifier onlyAllocationManager() {
        require(msg.sender == address(ALLOCATION_MANAGER), OnlyAllocationManager());
        _;
    }

    constructor(address avs, IAllocationManager allocationManager) TaskAVSRegistrarStorage(avs, allocationManager) {}

    function registerOperator(
        address operator,
        address avs,
        uint32[] calldata operatorSetIds,
        bytes calldata data
    ) external onlyAllocationManager {
        require(supportsAVS(avs), InvalidAVS());
        // TODO: Implement
    }

    function deregisterOperator(
        address operator,
        address avs,
        uint32[] calldata operatorSetIds
    ) external onlyAllocationManager {
        require(supportsAVS(avs), InvalidAVS());
        // TODO: Implement
    }

    function supportsAVS(
        address avs
    ) public view returns (bool) {
        return avs == AVS;
    }
}
