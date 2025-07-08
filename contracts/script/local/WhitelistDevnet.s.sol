// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {
    ICrossChainRegistry,
    ICrossChainRegistryTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/ICrossChainRegistry.sol";
import {IBN254TableCalculator} from "@eigenlayer-middleware/src/interfaces/IBN254TableCalculator.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract SetupAVSMultichain is Script {
    ICrossChainRegistry public CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0xe850D8A178777b483D37fD492a476e3E6004C816);

    function setUp() public {}

    function run() public {
        address ownerAddr = address(0xb094Ba769b4976Dc37fC689A76675f31bc4923b0);
        uint256 l1ChainId = uint256(vm.envUint("L1_CHAIN_ID"));
        uint256 l2ChainId = uint256(vm.envUint("L2_CHAIN_ID"));

        // Holesky is 17000, but when we run anvil it becomes 31337, so we need to whitelist 31337 as valid
        vm.startBroadcast();
        uint256[] memory chainIds = new uint256[](2);
        chainIds[0] = l1ChainId;
        chainIds[1] = l2ChainId;

        address[] memory tableUpdaters = new address[](2);
        // preprod holesky
        tableUpdaters[0] = address(0xE12C4cebd680a917271145eDbFB091B1BdEFD74D);
        // base sepolia
        tableUpdaters[1] = address(0xE12C4cebd680a917271145eDbFB091B1BdEFD74D);

        CROSS_CHAIN_REGISTRY.addChainIDsToWhitelist(chainIds, tableUpdaters);

        vm.stopBroadcast();
    }
}
