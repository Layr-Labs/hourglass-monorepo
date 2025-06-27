// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {TaskMailbox} from "../../src/core/TaskMailbox.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {ITaskMailboxTypes} from "../../src/interfaces/core/ITaskMailbox.sol";

contract DeployTaskMailbox is Script {
    // Eigenlayer Core Contracts
    address public BN254_CERTIFICATE_VERIFIER = 0x824604a31b580Aec16D8Dd7ae9A27661Dc65cBA3;
    address public ECDSA_CERTIFICATE_VERIFIER = 0x95A49cB0aED0e8f299223Da3A8A335440f5F00E7;

    function setUp() public {}

    function run() public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // Deploy the TaskMailbox contract
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        ITaskMailboxTypes.CertificateVerifierConfig[] memory certificateVerifiers =
            new ITaskMailboxTypes.CertificateVerifierConfig[](2);
        certificateVerifiers[0] = ITaskMailboxTypes.CertificateVerifierConfig({
            curveType: IKeyRegistrarTypes.CurveType.BN254,
            verifier: BN254_CERTIFICATE_VERIFIER
        });
        certificateVerifiers[1] = ITaskMailboxTypes.CertificateVerifierConfig({
            curveType: IKeyRegistrarTypes.CurveType.ECDSA,
            verifier: ECDSA_CERTIFICATE_VERIFIER
        });

        TaskMailbox taskMailbox = new TaskMailbox(deployer, certificateVerifiers);
        console.log("TaskMailbox deployed to:", address(taskMailbox));

        vm.stopBroadcast();
    }
}
