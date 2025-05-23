// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {ArtifactRegistry} from "../../src/core/ArtifactRegistry.sol";

contract DeployArtifactRegistry is Script {
    function setUp() public {}

    function run() public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // Deploy the ArtifactRegistry contract
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        ArtifactRegistry artifactRegistry = new ArtifactRegistry();
        console.log("ArtifactRegistry deployed to:", address(artifactRegistry));

        vm.stopBroadcast();
    }
}
