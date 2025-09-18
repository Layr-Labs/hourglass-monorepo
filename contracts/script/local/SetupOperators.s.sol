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

        uint256 executor2PrivateKey = vm.envUint("EXECUTOR2_PRIVATE_KEY");
        address executor2Addr = vm.addr(executor2PrivateKey);

        uint256 executor3PrivateKey = vm.envUint("EXECUTOR3_PRIVATE_KEY");
        address executor3Addr = vm.addr(executor3PrivateKey);

        uint256 executor4PrivateKey = vm.envUint("EXECUTOR4_PRIVATE_KEY");
        address executor4Addr = vm.addr(executor4PrivateKey);

        address zeroAddress = address(0);

        // Register aggregator
        vm.startBroadcast(aggregatorPrivateKey);
        DELEGATION_MANAGER.registerAsOperator(zeroAddress, 1, "");
        console.log("Aggregator registered as operator:", aggregatorAddr);
        vm.stopBroadcast();

        // Register executor 1
        vm.startBroadcast(executorPrivateKey);
        DELEGATION_MANAGER.registerAsOperator(zeroAddress, 1, "");
        console.log("Executor 1 registered as operator:", executorAddr);
        vm.stopBroadcast();

        // Register executor 2
        vm.startBroadcast(executor2PrivateKey);
        DELEGATION_MANAGER.registerAsOperator(zeroAddress, 1, "");
        console.log("Executor 2 registered as operator:", executor2Addr);
        vm.stopBroadcast();

        // Register executor 3
        vm.startBroadcast(executor3PrivateKey);
        DELEGATION_MANAGER.registerAsOperator(zeroAddress, 1, "");
        console.log("Executor 3 registered as operator:", executor3Addr);
        vm.stopBroadcast();

        // Register executor 4
        vm.startBroadcast(executor4PrivateKey);
        DELEGATION_MANAGER.registerAsOperator(zeroAddress, 1, "");
        console.log("Executor 4 registered as operator:", executor4Addr);
        vm.stopBroadcast();

        // Fast forward past the allocation delay
        uint256 currentTimestamp = block.timestamp;
        console.log("Current timestamp:", currentTimestamp);
        vm.warp(currentTimestamp + 10);
        console.log("Warped to timestamp:", block.timestamp);

        // Verify all operators are registered
        bool isOperator = DELEGATION_MANAGER.isOperator(aggregatorAddr);
        console.log("Check, is aggregator operator:", isOperator);
        isOperator = DELEGATION_MANAGER.isOperator(executorAddr);
        console.log("Check, is executor 1 operator:", isOperator);
        isOperator = DELEGATION_MANAGER.isOperator(executor2Addr);
        console.log("Check, is executor 2 operator:", isOperator);
        isOperator = DELEGATION_MANAGER.isOperator(executor3Addr);
        console.log("Check, is executor 3 operator:", isOperator);
        isOperator = DELEGATION_MANAGER.isOperator(executor4Addr);
        console.log("Check, is executor 4 operator:", isOperator);
    }
}
