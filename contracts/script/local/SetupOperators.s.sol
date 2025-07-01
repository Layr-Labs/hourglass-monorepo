// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {IDelegationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IDelegationManager.sol";

contract SetupOperators is Script {
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0xD4A7E1Bd8015057293f0D0A557088c286942e84b);

    function setUp() public {}

    function run() public {
        uint256 aggregatorPrivateKey = vm.envUint("AGGREGATOR_PRIVATE_KEY");
        address aggregatorAddr = vm.addr(aggregatorPrivateKey);

        uint256 executorPrivateKey = vm.envUint("EXECUTOR_PRIVATE_KEY");
        address executorAddr = vm.addr(executorPrivateKey);

        address zeroAddress = address(0);

        vm.startBroadcast(aggregatorPrivateKey);
        DELEGATION_MANAGER.registerAsOperator(zeroAddress, 1, "");
        console.log("Aggregator registered as operator:", aggregatorAddr);
        vm.stopBroadcast();

        vm.startBroadcast(executorPrivateKey);
        DELEGATION_MANAGER.registerAsOperator(zeroAddress, 1, "");
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
