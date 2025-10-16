// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {ProxyAdmin} from "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import {TransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import {ITransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IKeyRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {IPermissionController} from "@eigenlayer-contracts/src/contracts/interfaces/IPermissionController.sol";

import {MockTaskAVSRegistrar} from "@eigenlayer-middleware/test/mocks/MockTaskAVSRegistrar.sol";
import {ITaskAVSRegistrarBaseTypes} from "@eigenlayer-middleware/src/interfaces/ITaskAVSRegistrarBase.sol";

contract DeployAVSL1Contracts is Script {
    // Eigenlayer Core Contracts (Mainnet)
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);
    IKeyRegistrar public KEY_REGISTRAR = IKeyRegistrar(0x54f4bC6bDEbe479173a2bbDc31dD7178408A57A4);
    IPermissionController public PERMISSION_CONTROLLER =
        IPermissionController(0x25E5F8B1E7aDf44518d35D5B2271f114e081f0E5);

    function setUp() public {}

    function run(
        address avs
    ) public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        // 1. Deploy the TaskAVSRegistrar middleware contract
        vm.startBroadcast(deployerPrivateKey);
        console.log("Deployer address:", deployer);

        // Create initial config
        uint32[] memory executorOperatorSetIds = new uint32[](1);
        executorOperatorSetIds[0] = 1;
        ITaskAVSRegistrarBaseTypes.AvsConfig memory initialConfig = ITaskAVSRegistrarBaseTypes.AvsConfig({
            aggregatorOperatorSetId: 0, executorOperatorSetIds: executorOperatorSetIds
        });

        // Deploy ProxyAdmin
        ProxyAdmin proxyAdmin = new ProxyAdmin();
        console.log("ProxyAdmin deployed to:", address(proxyAdmin));

        // Deploy implementation
        MockTaskAVSRegistrar taskAVSRegistrarImpl =
            new MockTaskAVSRegistrar(ALLOCATION_MANAGER, KEY_REGISTRAR, PERMISSION_CONTROLLER);
        console.log("TaskAVSRegistrar implementation deployed to:", address(taskAVSRegistrarImpl));

        // Deploy proxy with initialization
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(taskAVSRegistrarImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(MockTaskAVSRegistrar.initialize.selector, avs, deployer, initialConfig)
        );
        console.log("TaskAVSRegistrar proxy deployed to:", address(proxy));

        // Transfer ProxyAdmin ownership to avs (or a multisig in production)
        proxyAdmin.transferOwnership(avs);

        vm.stopBroadcast();
    }
}
