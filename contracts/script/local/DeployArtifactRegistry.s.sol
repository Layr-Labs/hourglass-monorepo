// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {AVSArtifactRegistry} from "../../src/core/AVSArtifactRegistry.sol";

contract DeployAVSArtifactRegistry is Script {
    function setUp() public {}

    function run() public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // Deploy the AVSArtifactRegistry contract
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        AVSArtifactRegistry avsArtifactRegistry = new AVSArtifactRegistry();
        console.log("AVSArtifactRegistry deployed to:", address(avsArtifactRegistry));

        vm.stopBroadcast();
    }
}
