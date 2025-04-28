// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {TaskMailbox} from "src/core/TaskMailbox.sol";

contract DeployTaskMailbox is Script {
    function setUp() public {}

    function run() public {
        // Get the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");

        // Start broadcasting transactions
        vm.startBroadcast(deployerPrivateKey);

        // Deploy the TaskMailbox contract
        TaskMailbox taskMailbox = new TaskMailbox();

        // Log the contract address
        console.log("TaskMailbox deployed to:", address(taskMailbox));

        // Stop broadcasting transactions
        vm.stopBroadcast();
    }
}
