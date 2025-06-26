// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {TaskMailbox} from "../../src/core/TaskMailbox.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {ITaskMailboxTypes} from "../../src/interfaces/core/ITaskMailbox.sol";

contract DeployTaskMailbox is Script {
    // Eigenlayer Core Contracts
    address public BN254_CERTIFICATE_VERIFIER = 0xf462d03A82C1F3496B0DFe27E978318eD1720E1f;
    address public ECDSA_CERTIFICATE_VERIFIER = 0xF9BDd6af3Fb02659101cbb776DC7cb4879c93d8A;

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
