// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {ProxyAdmin} from "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import {TransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import {ITransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

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

        // Deploy ProxyAdmin
        ProxyAdmin proxyAdmin = new ProxyAdmin();
        console.log("ProxyAdmin deployed to:", address(proxyAdmin));

        // Deploy implementation
        TaskMailbox taskMailboxImpl = new TaskMailbox(bn254CertVerifier, ecdsaCertVerifier, "1.0.0");
        console.log("TaskMailbox implementation deployed to:", address(taskMailboxImpl));

        // Deploy proxy with initialization
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(taskMailboxImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(TaskMailbox.initialize.selector, deployer)
        );
        console.log("TaskMailbox proxy deployed to:", address(proxy));

        // Transfer ProxyAdmin ownership to deployer (or a multisig in production)
        proxyAdmin.transferOwnership(deployer);

        vm.stopBroadcast();
    }
}
