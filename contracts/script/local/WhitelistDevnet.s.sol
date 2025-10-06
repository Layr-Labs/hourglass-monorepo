// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

import {
    ICrossChainRegistry,
    ICrossChainRegistryTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/ICrossChainRegistry.sol";
import {IBN254TableCalculator} from "@eigenlayer-middleware/src/interfaces/IBN254TableCalculator.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract WhitelistDevnet is Script {
    function setUp() public {}

    function run() public {
        // Get registry and chain configuration from environment
        address crossChainRegistryAddress = vm.envAddress("CROSS_CHAIN_REGISTRY");
        address tableUpdaterAddress = vm.envAddress("TABLE_UPDATER_ADDRESS");
        uint256 l1ChainId = uint256(vm.envUint("L1_CHAIN_ID"));
        uint256 l2ChainId = uint256(vm.envUint("L2_CHAIN_ID"));

        ICrossChainRegistry CROSS_CHAIN_REGISTRY = ICrossChainRegistry(crossChainRegistryAddress);

        // Get the owner of the CrossChainRegistry by casting to Ownable
        address owner = Ownable(address(CROSS_CHAIN_REGISTRY)).owner();
        console.log("CrossChainRegistry address:", crossChainRegistryAddress);
        console.log("CrossChainRegistry owner:", owner);
        console.log("Table updater address:", tableUpdaterAddress);

        // Whitelist anvil chain IDs
        uint256[] memory chainIds = new uint256[](2);
        chainIds[0] = l1ChainId;
        chainIds[1] = l2ChainId;

        address[] memory tableUpdaters = new address[](2);
        tableUpdaters[0] = tableUpdaterAddress;
        tableUpdaters[1] = tableUpdaterAddress;

        vm.startBroadcast();
        CROSS_CHAIN_REGISTRY.addChainIDsToWhitelist(chainIds, tableUpdaters);
        vm.stopBroadcast();

        console.log("Successfully whitelisted chains:", l1ChainId, l2ChainId);
    }
}
