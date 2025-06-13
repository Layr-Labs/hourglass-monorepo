// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";

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
        // Load the private key from the environment variable
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");
        uint256 l1ChainId = uint256(vm.envUint("L1_CHAIN_ID"));

        // TODO(seanmcgary): update to use later
        // uint32 l2ChainId = uint32(vm.envUint("L2_CHAIN_ID"));
        // Holesky is 17000, but when we run anvil it becomes 31337, so we need to whitelist 31337 as valid
        uint256[] memory chainIds = new uint256[](1);
        chainIds[0] = l1ChainId;

        vm.startBroadcast(avsPrivateKey);
        address avs = vm.addr(avsPrivateKey);
        console.log("AVS address:", avs);

        // create reservations in the cross chain registry for each operator set
        for (uint32 i = 0; i < 2; i++) {
            OperatorSet memory operatorSet = OperatorSet({avs: avs, id: i});
            ICrossChainRegistryTypes.OperatorSetConfig memory config = ICrossChainRegistryTypes.OperatorSetConfig({
                owner: avs,
                maxStalenessPeriod: 604_800 // 1 week
            });

            CROSS_CHAIN_REGISTRY.createGenerationReservation(operatorSet, BN254_TABLE_CALCULATOR, config, chainIds);
            console.log("Generation reservation created for operator set", i);
        }

        vm.stopBroadcast();

        OperatorSet[] memory reservations = CROSS_CHAIN_REGISTRY.getActiveGenerationReservations();
        console.log("Number of reservations:", reservations.length);

        for (uint256 i = 0; i < reservations.length; i++) {
            console.log("Reservation", i, "- AVS:", reservations[i].avs);
            console.log("Reservation", i, "- ID:", reservations[i].id);
        }
    }
}
