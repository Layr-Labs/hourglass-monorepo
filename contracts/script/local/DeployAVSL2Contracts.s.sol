// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";

import {MockAVSTaskHook} from "../../test/mocks/MockAVSTaskHook.sol";

contract DeployAVSL2Contracts is Script {
    function setUp() public {}

    function run() public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // Deploy the AVSTaskHook contract
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        MockAVSTaskHook avsTaskHook = new MockAVSTaskHook();
        console.log("AVSTaskHook deployed to:", address(avsTaskHook));

        vm.stopBroadcast();
    }
}
