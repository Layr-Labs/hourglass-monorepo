// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

contract CheckL1State is Script {
    function run() public {
        console.log("\n========================================");
        console.log("L1 STATE CHECK");
        console.log("========================================\n");

        // Check CrossChainRegistry getSupportedChains
        address crossChainRegistry = 0x287381B1570d9048c4B4C7EC94d21dDb8Aa1352a;
        console.log("CrossChainRegistry:", crossChainRegistry);

        (bool success, bytes memory data) =
            crossChainRegistry.staticcall(abi.encodeWithSignature("getSupportedChains()"));

        if (success && data.length > 0) {
            (uint256[] memory chainIds, address[] memory updaters) = abi.decode(data, (uint256[], address[]));

            console.log("Supported chains count:", chainIds.length);
            for (uint256 i = 0; i < chainIds.length; i++) {
                console.log("  Chain ID:", chainIds[i]);
                console.log("  Table Updater:", updaters[i]);
            }
        } else {
            console.log("Failed to get supported chains");
        }

        // Check OperatorTableUpdater
        address operatorTableUpdater = 0xB02A15c6Bd0882b35e9936A9579f35FB26E11476;
        console.log("\nOperatorTableUpdater:", operatorTableUpdater);

        (success, data) = operatorTableUpdater.staticcall(abi.encodeWithSignature("getLatestReferenceTimestamp()"));

        if (success) {
            uint32 timestamp = abi.decode(data, (uint32));
            console.log("Latest Reference Timestamp:", timestamp);

            // Get global root for this timestamp
            (success, data) = operatorTableUpdater.staticcall(
                abi.encodeWithSignature("getGlobalTableRootByTimestamp(uint32)", timestamp)
            );

            if (success) {
                bytes32 root = abi.decode(data, (bytes32));
                console.log("Global Table Root:");
                console.logBytes32(root);
            }
        }

        // Check BN254 operator set config
        console.log("\nChecking BN254 operator set (AVS:", 0x8e14dB002737F89745bc98F987caeB18D0d47635);
        console.log("Operator Set ID: 1");

        // Encode the operator set struct
        bytes memory operatorSet = abi.encode(0x8e14dB002737F89745bc98F987caeB18D0d47635, uint32(1));

        (success, data) = crossChainRegistry.staticcall(
            abi.encodeWithSignature("getOperatorTableCalculator((address,uint32))", operatorSet)
        );

        if (success) {
            address tableCalc = abi.decode(data, (address));
            console.log("BN254 Table Calculator:", tableCalc);
        } else {
            console.log("Failed to get table calculator");
        }
    }
}
