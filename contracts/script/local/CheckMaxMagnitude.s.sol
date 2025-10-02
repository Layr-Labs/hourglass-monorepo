// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {IStrategy} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IDelegationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IDelegationManager.sol";

contract CheckMaxMagnitude is Script {
    IAllocationManager public constant ALLOCATION_MANAGER =
        IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);
    IDelegationManager public constant DELEGATION_MANAGER =
        IDelegationManager(0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A);

    IStrategy public constant STRATEGY_STETH = IStrategy(0x93c4b944D05dfe6df7645A86cd2206016c51564D);

    function setUp() public {}

    function run(
        address operatorAddr
    ) public view {
        console.log("=== Checking Max Magnitude for Operator ===");
        console.log("Operator:", operatorAddr);
        console.log("");

        // Check max magnitude from AllocationManager
        uint64 maxMagnitude = ALLOCATION_MANAGER.getMaxMagnitude(operatorAddr, STRATEGY_STETH);
        console.log("AllocationManager.getMaxMagnitude:", maxMagnitude);

        // Check allocation delay
        (bool isSet, uint32 delay) = ALLOCATION_MANAGER.getAllocationDelay(operatorAddr);
        console.log("Allocation Delay - isSet:", isSet, "delay:", delay);

        // Check if operator is registered
        bool isOperator = DELEGATION_MANAGER.isOperator(operatorAddr);
        console.log("Is registered EigenLayer operator:", isOperator);

        console.log("");
        console.log("=== Analysis ===");
        if (maxMagnitude == 0) {
            console.log("ERROR: maxMagnitude is 0!");
            console.log("  This means operator cannot allocate any magnitude");
            console.log("  Root cause: Either no staker delegated, or delegation not finalized");
            console.log("  Any ModifyAllocations call will fail with InsufficientMagnitude()");
        } else {
            console.log("OK: maxMagnitude =", maxMagnitude);
            console.log("  Operator can allocate up to", maxMagnitude, "magnitude");
        }

        if (!isSet) {
            console.log("ERROR: Allocation delay not initialized");
            console.log("  ModifyAllocations will revert with UninitializedAllocationDelay()");
        }

        if (maxMagnitude > 0 && isSet && isOperator) {
            console.log("");
            console.log("SUCCESS: Operator is ready for ModifyAllocations!");
        }
    }
}
