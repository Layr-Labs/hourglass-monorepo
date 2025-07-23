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
    ICrossChainRegistry public CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0x287381B1570d9048c4B4C7EC94d21dDb8Aa1352a);
    IBN254TableCalculator public BN254_TABLE_CALCULATOR =
        IBN254TableCalculator(0xc2c0bc13571aC5115709C332dc7AE666606b08E8);

    IECDSATableCalculator public ECDSA_TABLE_CALCULATOR =
        IECDSATableCalculator(0x5612Fd146C2d40f1269E0e73945A534ec706dCDc);

    function setUp() public {}

    function run() public {
        // Load the private key from the environment variable
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");

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

            // aggregator is bn254, executor is ecdsa
            if (i == 0) {
                CROSS_CHAIN_REGISTRY.createGenerationReservation(operatorSet, BN254_TABLE_CALCULATOR, config);
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
