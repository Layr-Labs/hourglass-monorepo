// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {IDelegationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IDelegationManager.sol";

contract SetupOperators is Script {
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0x75dfE5B44C2E530568001400D3f704bC8AE350CC);

    function setUp() public {}

    function run() public {
        uint256 aggregatorPrivateKey = vm.envUint("AGGREGATOR_PRIVATE_KEY");
        address aggregatorAddr = vm.addr(aggregatorPrivateKey);

        uint256 executorPrivateKey = vm.envUint("EXECUTOR_PRIVATE_KEY");
        address executorAddr = vm.addr(executorPrivateKey);

        vm.startBroadcast(aggregatorPrivateKey);
        DELEGATION_MANAGER.registerAsOperator(aggregatorAddr, 1, "");
        console.log("Aggregator registered as operator:", aggregatorAddr);
        vm.stopBroadcast();

        vm.startBroadcast(executorPrivateKey);
        DELEGATION_MANAGER.registerAsOperator(executorAddr, 1, "");
        console.log("Executor registered as operator:", executorAddr);
        vm.stopBroadcast();

        // Fast forward past the allocation delay
        uint256 currentTimestamp = block.timestamp;
        console.log("Current timestamp:", currentTimestamp);
        vm.warp(currentTimestamp + 10);
        console.log("Warped to timestamp:", block.timestamp);

        bool isOperator = DELEGATION_MANAGER.isOperator(aggregatorAddr);
        console.log("Check, is aggregator operator:", isOperator);
        isOperator = DELEGATION_MANAGER.isOperator(executorAddr);
        console.log("Check, is executor operator:", isOperator);
    }
}
