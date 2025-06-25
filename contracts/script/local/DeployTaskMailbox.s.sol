// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {TaskMailbox} from "../../src/core/TaskMailbox.sol";
import {ITaskMailboxTypes} from "../../src/interfaces/core/ITaskMailbox.sol";

contract DeployTaskMailbox is Script {
    function setUp() public {}

    function run() public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // Deploy the TaskMailbox contract
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        // For now, deploy with empty certificate verifiers array
        // The owner can set them later using setCertificateVerifier
        ITaskMailboxTypes.CertificateVerifierConfig[] memory certificateVerifiers =
            new ITaskMailboxTypes.CertificateVerifierConfig[](0);

        TaskMailbox taskMailbox = new TaskMailbox(deployer, certificateVerifiers);
        console.log("TaskMailbox deployed to:", address(taskMailbox));

        vm.stopBroadcast();
    }
}
