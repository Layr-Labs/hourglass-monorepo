// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

import {ITaskMailbox, ITaskMailboxTypes} from "../../src/interfaces/core/ITaskMailbox.sol";

contract CreateTask is Script {
    function setUp() public {}

    function run(address taskMailbox, address avs) public {
        // Load the private key from the environment variable
        uint256 appPrivateKey = vm.envUint("PRIVATE_KEY_APP");
        address app = vm.addr(appPrivateKey);

        vm.startBroadcast(appPrivateKey);
        console.log("App address:", app);

        // Call createTask
        OperatorSet memory executorOperatorSet = OperatorSet({avs: avs, id: 1});
        ITaskMailboxTypes.TaskParams memory taskParams = ITaskMailboxTypes.TaskParams({
            refundCollector: address(0),
            avsFee: 0,
            executorOperatorSet: executorOperatorSet,
            payload: bytes("Hello World")
        });
        bytes32 taskHash = ITaskMailbox(taskMailbox).createTask(taskParams);
        console.log("Created task with hash:");
        console.logBytes32(taskHash);
        ITaskMailboxTypes.Task memory task = ITaskMailbox(taskMailbox).getTaskInfo(taskHash);
        console.log("Task status:", uint8(task.status));
        console.log("Task payload:", string(task.payload));

        vm.stopBroadcast();
    }
}
