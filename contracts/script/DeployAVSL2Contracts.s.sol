// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";

import {AVSTaskHook} from "../src/avs/l2-contracts/AVSTaskHook.sol";
import {BN254CertificateVerifier} from "../src/avs/l2-contracts/BN254CertificateVerifier.sol";

contract DeployAVSL2Contracts is Script {
    function setUp() public {}

    function run() public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // Deploy the AVSTaskHook and CertificateVerifier contracts
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        AVSTaskHook avsTaskHook = new AVSTaskHook();
        console.log("AVSTaskHook deployed to:", address(avsTaskHook));

        BN254CertificateVerifier bn254CertificateVerifier = new BN254CertificateVerifier();
        console.log("BN254CertificateVerifier deployed to:", address(bn254CertificateVerifier));

        vm.stopBroadcast();
    }
}
