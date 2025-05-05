// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";

import {MockAVSTaskHook} from "../../test/mocks/MockAVSTaskHook.sol";
import {MockBN254CertificateVerifier} from "../../test/mocks/MockBN254CertificateVerifier.sol";

contract DeployAVSL2Contracts is Script {
    function setUp() public {}

    function run() public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // Deploy the AVSTaskHook and CertificateVerifier contracts
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        MockAVSTaskHook avsTaskHook = new MockAVSTaskHook();
        console.log("AVSTaskHook deployed to:", address(avsTaskHook));

        MockBN254CertificateVerifier bn254CertificateVerifier = new MockBN254CertificateVerifier();
        console.log("BN254CertificateVerifier deployed to:", address(bn254CertificateVerifier));

        vm.stopBroadcast();
    }
}
