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
    // Mainnet CrossChainRegistry
    ICrossChainRegistry public CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0x9376A5863F2193cdE13e1aB7c678F22554E2Ea2b);

    function setUp() public {}

    function run() public {
        uint256 l1ChainId = uint256(vm.envUint("L1_CHAIN_ID"));
        uint256 l2ChainId = uint256(vm.envUint("L2_CHAIN_ID"));

        // Get the owner of the CrossChainRegistry by casting to Ownable
        address owner = Ownable(address(CROSS_CHAIN_REGISTRY)).owner();
        console.log("CrossChainRegistry owner:", owner);

        // Mainnet anvil - whitelist anvil chain IDs (31337 for Ethereum, 31338 for Base)
        uint256[] memory chainIds = new uint256[](2);
        chainIds[0] = l1ChainId;
        chainIds[1] = l2ChainId;

        address[] memory tableUpdaters = new address[](2);
        // Mainnet OperatorTableUpdater (same for both Ethereum and Base)
        tableUpdaters[0] = address(0x5557E1fE3068A1e823cE5Dcd052c6C352E2617B5);
        tableUpdaters[1] = address(0x5557E1fE3068A1e823cE5Dcd052c6C352E2617B5);

        // Impersonate the owner to make the call
        vm.prank(owner);
        CROSS_CHAIN_REGISTRY.addChainIDsToWhitelist(chainIds, tableUpdaters);
    }
}
