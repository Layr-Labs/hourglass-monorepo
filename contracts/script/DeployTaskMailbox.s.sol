// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {TaskMailbox} from "src/core/TaskMailbox.sol";

contract DeployTaskMailbox is Script {
    function setUp() public {}

    function run() public {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        console.log("Deployer address:", vm.addr(deployerPrivateKey));

        vm.startBroadcast(deployerPrivateKey);
        TaskMailbox taskMailbox = new TaskMailbox();
        console.log("TaskMailbox deployed to:", address(taskMailbox));
        vm.stopBroadcast();
    }
}
