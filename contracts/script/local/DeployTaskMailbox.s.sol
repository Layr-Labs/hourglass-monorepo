// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {ProxyAdmin} from "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import {TransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import {ITransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

import {TaskMailbox} from "@eigenlayer-contracts/src/contracts/avs/task/TaskMailbox.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {ITaskMailboxTypes} from "@eigenlayer-contracts/src/contracts/interfaces/ITaskMailbox.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";

contract DeployTaskMailbox is Script {
    function setUp() public {}

    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);

    function run(
        address bn254CertVerifier,
        address ecdsaCertVerifier
    ) public {
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
        TaskMailbox taskMailboxImpl =
            new TaskMailbox(bn254CertVerifier, ecdsaCertVerifier, ALLOCATION_MANAGER.DEALLOCATION_DELAY() / 2, "0.0.1");
        console.log("TaskMailbox implementation deployed to:", address(taskMailboxImpl));

        // Deploy proxy with initialization
        // Initialize with no fee split (0) and deployer as fee split collector
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(taskMailboxImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(TaskMailbox.initialize.selector, deployer, 0, deployer)
        );
        console.log("TaskMailbox proxy deployed to:", address(proxy));

        // Transfer ProxyAdmin ownership to deployer (or a multisig in production)
        proxyAdmin.transferOwnership(deployer);

        vm.stopBroadcast();
    }
}
