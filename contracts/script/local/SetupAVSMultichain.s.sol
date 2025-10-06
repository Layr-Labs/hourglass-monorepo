// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";

import {
    ICrossChainRegistry,
    ICrossChainRegistryTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/ICrossChainRegistry.sol";
import {IBN254TableCalculator} from "@eigenlayer-middleware/src/interfaces/IBN254TableCalculator.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IECDSATableCalculator} from "@eigenlayer-middleware/src/interfaces/IECDSATableCalculator.sol";
import {BLSApkRegistry} from "@eigenlayer-middleware/src/BLSApkRegistry.sol";

contract SetupAVSMultichain is Script {
    function setUp() public {}

    function run() public {
        // Load addresses from environment variables
        address crossChainRegistryAddress = vm.envAddress("CROSS_CHAIN_REGISTRY");
        address bn254TableCalculatorAddress = vm.envAddress("BN254_TABLE_CALCULATOR");
        address ecdsaTableCalculatorAddress = vm.envAddress("ECDSA_TABLE_CALCULATOR");
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");

        ICrossChainRegistry CROSS_CHAIN_REGISTRY = ICrossChainRegistry(crossChainRegistryAddress);
        IBN254TableCalculator BN254_TABLE_CALCULATOR = IBN254TableCalculator(bn254TableCalculatorAddress);
        IECDSATableCalculator ECDSA_TABLE_CALCULATOR = IECDSATableCalculator(ecdsaTableCalculatorAddress);

        vm.startBroadcast(avsPrivateKey);
        address avs = vm.addr(avsPrivateKey);

        console.log("AVS address:", avs);
        console.log("CrossChainRegistry:", crossChainRegistryAddress);
        console.log("BN254 Table Calculator:", bn254TableCalculatorAddress);
        console.log("ECDSA Table Calculator:", ecdsaTableCalculatorAddress);

        // create reservations in the cross chain registry for each operator set
        for (uint32 i = 0; i < 2; i++) {
            OperatorSet memory operatorSet = OperatorSet({avs: avs, id: i});
            ICrossChainRegistryTypes.OperatorSetConfig memory config = ICrossChainRegistryTypes.OperatorSetConfig({
                owner: avs,
                maxStalenessPeriod: 604_800 // 1 week
            });

            if (i == 0) {
                CROSS_CHAIN_REGISTRY.createGenerationReservation(operatorSet, ECDSA_TABLE_CALCULATOR, config);
            } else {
                CROSS_CHAIN_REGISTRY.createGenerationReservation(operatorSet, ECDSA_TABLE_CALCULATOR, config);
            }

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
