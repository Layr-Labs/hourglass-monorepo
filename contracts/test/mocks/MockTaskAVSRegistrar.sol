// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";

import {TaskAVSRegistrarBase} from "../../src/avs/TaskAVSRegistrarBase.sol";

contract MockTaskAVSRegistrar is TaskAVSRegistrarBase {
    constructor(address avs, IAllocationManager allocationManager) TaskAVSRegistrarBase(avs, allocationManager) {}
}
