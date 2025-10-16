// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {IBN254CertificateVerifier} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract CheckL2State is Script {
    function run() public view {
        console.log("\n========================================");
        console.log("L2 STATE CHECK (BASE MAINNET FORK)");
        console.log("========================================\n");

        // Check CrossChainRegistry getSupportedChains
        address crossChainRegistry = 0x9376A5863F2193cdE13e1aB7c678F22554E2Ea2b;
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
        address operatorTableUpdater = 0x5557E1fE3068A1e823cE5Dcd052c6C352E2617B5;
        console.log("\nOperatorTableUpdater:", operatorTableUpdater);

        (success, data) = operatorTableUpdater.staticcall(abi.encodeWithSignature("getLatestReferenceTimestamp()"));

        uint32 timestamp = 0;
        if (success) {
            timestamp = abi.decode(data, (uint32));
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
        } else {
            console.log("No OperatorTableUpdater found or not initialized");
        }

        // Check BN254CertificateVerifier - THIS IS THE CRITICAL CHECK
        console.log("\n========================================");
        console.log("BN254 CERTIFICATE VERIFIER (CRITICAL)");
        console.log("========================================\n");

        address bn254Verifier = 0xff58A373c18268F483C1F5cA03Cf885c0C43373a;
        console.log("BN254CertificateVerifier:", bn254Verifier);

        // Check operator set info for BN254 operator set
        OperatorSet memory bn254OpSet = OperatorSet({avs: 0x8e14dB002737F89745bc98F987caeB18D0d47635, id: 1});

        console.log("Checking BN254 Operator Set:");
        console.log("  AVS:", bn254OpSet.avs);
        console.log("  Operator Set ID:", bn254OpSet.id);

        // Use a reasonable timestamp if we don't have one
        if (timestamp == 0) {
            timestamp = 1_735_272_624; // You may need to adjust this
            console.log("  Using hardcoded timestamp:", timestamp);
        } else {
            console.log("  Reference Timestamp:", timestamp);
        }

        IBN254CertificateVerifier verifier = IBN254CertificateVerifier(bn254Verifier);

        try verifier.getOperatorSetInfo(
            bn254OpSet, timestamp
        ) returns (IBN254CertificateVerifier.BN254OperatorSetInfo memory info) {
            console.log("\nOperator Set Info Retrieved:");
            console.log("  Number of operators:", info.numOperators);

            if (info.aggregatePubkey.X != 0 || info.aggregatePubkey.Y != 0) {
                console.log("  Aggregate pubkey X:", info.aggregatePubkey.X);
                console.log("  Aggregate pubkey Y:", info.aggregatePubkey.Y);
            } else {
                console.log("  Aggregate pubkey: NOT SET");
            }

            // THIS IS THE CRITICAL CHECK
            console.log("\n  >>> OPERATOR INFO TREE ROOT <<<");
            console.logBytes32(info.operatorInfoTreeRoot);

            if (info.operatorInfoTreeRoot == bytes32(0)) {
                console.log("  >>> ERROR: OPERATOR INFO TREE ROOT IS EMPTY!");
                console.log("  >>> This is why BN254 certificate verification fails!");
                console.log("  >>> The L2 transport needs to set this value!");
            } else {
                console.log("  >>> SUCCESS: Operator info tree root is present");
            }

            if (info.totalWeights.length > 0) {
                console.log("\n  Total weights[0]:", info.totalWeights[0]);
            }
        } catch Error(string memory reason) {
            console.log("  >>> ERROR: Failed to get operator set info:", reason);
        } catch {
            console.log("  >>> ERROR: Failed to get operator set info (unknown error)");
        }

        console.log("\n========================================");
        console.log("DIAGNOSIS");
        console.log("========================================\n");

        console.log("If the operator info tree root is empty (0x0000...),");
        console.log("then the L2 transport process is NOT properly updating");
        console.log("the BN254CertificateVerifier with operator-specific data.");
        console.log("");
        console.log("The multiOperatorTransport needs to:");
        console.log("1. Calculate the operator info tree root for BN254 operators");
        console.log("2. Call updateOperatorTable on the L2 BN254CertificateVerifier");
        console.log("3. Pass the operator info tree root along with other data");
    }
}
