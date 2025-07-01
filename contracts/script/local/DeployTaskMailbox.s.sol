// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {TaskMailbox} from "../../src/core/TaskMailbox.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {ITaskMailboxTypes} from "../../src/interfaces/core/ITaskMailbox.sol";

contract DeployTaskMailbox is Script {
    function setUp() public {}

    function run(address bn254CertVerifier, address ecdsaCertVerifier) public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // Deploy the TaskMailbox contract
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        TaskMailbox taskMailbox = new TaskMailbox(deployer, bn254CertVerifier, ecdsaCertVerifier);
        console.log("TaskMailbox deployed to:", address(taskMailbox));

        vm.stopBroadcast();
    }
}
