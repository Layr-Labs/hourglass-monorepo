// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {
    ICrossChainRegistry,
    ICrossChainRegistryTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/ICrossChainRegistry.sol";
import {IBN254TableCalculator} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254TableCalculator.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract SetupAVSMultichain is Script {
    ICrossChainRegistry public CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0x0022d2014901F2AFBF5610dDFcd26afe2a65Ca6F);
    IBN254TableCalculator public BN254_TABLE_CALCULATOR =
        IBN254TableCalculator(0x033af59c1b030Cc6eEE07B150FD97668497dc74b);

    function setUp() public {}

    function run() public {
        address ownerAddr = address(0xDA29BB71669f46F2a779b4b62f03644A84eE3479);
        uint256 l1ChainId = uint256(vm.envUint("L1_CHAIN_ID"));

        // TODO(seanmcgary): update to use later
        // uint32 l2ChainId = uint32(vm.envUint("L2_CHAIN_ID"));
        // Holesky is 17000, but when we run anvil it becomes 31337, so we need to whitelist 31337 as valid
        vm.startBroadcast();
        uint256[] memory chainIds = new uint256[](1);
        chainIds[0] = l1ChainId;

        address[] memory tableUpdaters = new address[](1);
        tableUpdaters[0] = address(0xd7230B89E5E2ed1FD068F0FF9198D7960243f12a);

        CROSS_CHAIN_REGISTRY.addChainIDsToWhitelist(chainIds, tableUpdaters);

        vm.stopBroadcast();
    }
}
