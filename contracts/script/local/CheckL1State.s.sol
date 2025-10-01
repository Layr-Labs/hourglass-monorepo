// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {IBN254TableCalculator} from "@eigenlayer-middleware/src/interfaces/IBN254TableCalculator.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IOperatorTableCalculator, IOperatorTableCalculatorTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IOperatorTableCalculator.sol";
import {IBN254CertificateVerifier} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {ICrossChainRegistry} from "@eigenlayer-contracts/src/contracts/interfaces/ICrossChainRegistry.sol";
import {IOperatorTableUpdater} from "@eigenlayer-contracts/src/contracts/interfaces/IOperatorTableUpdater.sol";
import {IStrategyManager} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategyManager.sol";
import {IDelegationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IDelegationManager.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IStrategy} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";

contract CheckL1State is Script {
    // Mainnet contract addresses
    ICrossChainRegistry constant CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0x9376A5863F2193cdE13e1aB7c678F22554E2Ea2b);
    IOperatorTableUpdater constant OPERATOR_TABLE_UPDATER = IOperatorTableUpdater(0x5557E1fE3068A1e823cE5Dcd052c6C352E2617B5);
    IStrategyManager constant STRATEGY_MANAGER = IStrategyManager(0x858646372CC42E1A627fcE94aa7A7033e7CF075A);
    IDelegationManager constant DELEGATION_MANAGER = IDelegationManager(0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A);
    IAllocationManager constant ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);

    IStrategy constant STETH_STRATEGY = IStrategy(0x93c4b944D05dfe6df7645A86cd2206016c51564D);
    IStrategy constant WETH_STRATEGY = IStrategy(0x0Fe4F44beE93503346A3Ac9EE5A26b130a5796d6);

    // BN254 Table Calculator address
    IBN254TableCalculator constant BN254_TABLE_CALCULATOR = IBN254TableCalculator(0x55F4b21681977F412B318eCB204cB933bD1dF57c);

    // Test addresses (from chain config)
    address constant AVS_ADDRESS = 0x8e14dB002737F89745bc98F987caeB18D0d47635;
    address constant AGG_STAKER = 0x17f894Dfb1918a21aeE5c401f35DeB65ED150CCD;
    address constant AGG_OPERATOR = 0x77bF74ED0d350bcfD198445E4F379C8fc88029e4;
    address constant EXEC_OPERATOR_1 = 0xD64ba8C9fA929E3A030a52a0Ed73f13224c8c37C;
    address constant EXEC_OPERATOR_2 = 0x32436B945658d8Aa9d26A3089E194D764e65d718;
    address constant EXEC_OPERATOR_3 = 0x08d063f823Ce335d98d60bd0BEA3C6Dff1BB5A69;
    address constant EXEC_OPERATOR_4 = 0xf3972Db6721571A5dd7963ec3D68F551839ac817;
    address constant EXEC_STAKER_1 = 0x625eF3cac014322840D373ED1E11dAc244E26A40;
    address constant EXEC_STAKER_2 = 0x4dBEBAcaFc37817D824b8C182752d33a943F5B2D;
    address constant EXEC_STAKER_3 = 0x009D994d30C96335de62035B0A34aD7220192406;
    address constant EXEC_STAKER_4 = 0x0B4e7eAA62CD95b3f9720b86E79d3234fB0A1B81;

    function run() public view {
        console.log("\n========================================");
        console.log("L1 STATE CHECK");
        console.log("========================================\n");

        checkCrossChainRegistry();
        checkOperatorTableUpdater();
        checkOperatorSets();
        checkStakers();
        checkOperators();
        checkBN254TableCalculator();
        checkBN254CertificateVerifier();
    }

    function checkCrossChainRegistry() internal view {
        console.log("=== CrossChainRegistry ===");
        console.log("Address:", address(CROSS_CHAIN_REGISTRY));

        try CROSS_CHAIN_REGISTRY.getSupportedChains() returns (
            uint256[] memory chainIds,
            address[] memory updaters
        ) {
            console.log("Supported chains count:", chainIds.length);
            for (uint256 i = 0; i < chainIds.length; i++) {
                console.log("  Chain ID:", chainIds[i], "Updater:", updaters[i]);
            }
        } catch {
            console.log("ERROR: Failed to get supported chains");
        }
        console.log("");
    }

    function checkOperatorTableUpdater() internal view {
        console.log("=== OperatorTableUpdater ===");
        console.log("Address:", address(OPERATOR_TABLE_UPDATER));

        try OPERATOR_TABLE_UPDATER.getLatestReferenceTimestamp() returns (uint32 timestamp) {
            console.log("Latest Reference Timestamp:", timestamp);

            try OPERATOR_TABLE_UPDATER.getGlobalTableRootByTimestamp(timestamp) returns (bytes32 root) {
                console.log("Global Table Root:");
                console.logBytes32(root);
            } catch {
                console.log("ERROR: Failed to get global table root");
            }
        } catch {
            console.log("ERROR: Failed to get timestamp");
        }
        console.log("");
    }

    function checkOperatorSets() internal view {
        console.log("=== Operator Sets Configuration ===");
        console.log("AVS Address:", AVS_ADDRESS);

        // Check operator set 0 (aggregator)
        console.log("\n--- Operator Set 0 (Aggregator) ---");
        checkOperatorSetConfig(AVS_ADDRESS, 0);

        // Check operator set 1 (executors)
        console.log("\n--- Operator Set 1 (Executors) ---");
        checkOperatorSetConfig(AVS_ADDRESS, 1);

        console.log("");
    }

    function checkOperatorSetConfig(address avs, uint32 operatorSetId) internal view {
        OperatorSet memory operatorSet = OperatorSet({
            avs: avs,
            id: operatorSetId
        });

        // Check if table calculator is configured
        try CROSS_CHAIN_REGISTRY.getOperatorTableCalculator(operatorSet) returns (IOperatorTableCalculator tableCalc) {
            console.log("  Table Calculator:", address(tableCalc));
            if (address(tableCalc) == address(0)) {
                console.log("  WARNING: No table calculator configured!");
            }
        } catch {
            console.log("  ERROR: Failed to get table calculator");
        }

        // Check if operators are registered in this set
        checkOperatorInSet(AGG_OPERATOR, avs, operatorSetId, "Aggregator");
        checkOperatorInSet(EXEC_OPERATOR_1, avs, operatorSetId, "Executor 1");
        checkOperatorInSet(EXEC_OPERATOR_2, avs, operatorSetId, "Executor 2");
        checkOperatorInSet(EXEC_OPERATOR_3, avs, operatorSetId, "Executor 3");
        checkOperatorInSet(EXEC_OPERATOR_4, avs, operatorSetId, "Executor 4");
    }

    function checkOperatorInSet(address operator, address avs, uint32 operatorSetId, string memory label) internal view {
        OperatorSet memory operatorSet = OperatorSet({
            avs: avs,
            id: operatorSetId
        });

        console.log("  ", label, ":", operator);
        try ALLOCATION_MANAGER.isMemberOfOperatorSet(operator, operatorSet) returns (bool isRegistered) {
            if (isRegistered) {
                console.log("    Status: REGISTERED");
            } else {
                console.log("    Status: NOT REGISTERED");
            }
        } catch {
            console.log("    Status: ERROR checking registration");
        }
    }

    function checkStakers() internal view {
        console.log("=== Staker States ===");

        console.log("\n--- Aggregator Staker ---");
        checkStakerState(AGG_STAKER, WETH_STRATEGY);

        console.log("\n--- Executor 1 Staker ---");
        checkStakerState(EXEC_STAKER_1, STETH_STRATEGY);

        console.log("\n--- Executor 2 Staker ---");
        checkStakerState(EXEC_STAKER_2, STETH_STRATEGY);

        console.log("\n--- Executor 3 Staker ---");
        checkStakerState(EXEC_STAKER_3, STETH_STRATEGY);

        console.log("\n--- Executor 4 Staker ---");
        checkStakerState(EXEC_STAKER_4, STETH_STRATEGY);

        console.log("");
    }

    function checkStakerState(address staker, IStrategy strategy) internal view {
        console.log("Address:", staker);

        // Check deposited shares
        try STRATEGY_MANAGER.stakerDepositShares(staker, strategy) returns (uint256 shares) {
            console.log("Deposited shares:", shares);
            if (shares == 0) {
                console.log("WARNING: No deposited shares!");
            }
        } catch {
            console.log("ERROR: Failed to get deposited shares");
        }

        // Check delegation
        try DELEGATION_MANAGER.delegatedTo(staker) returns (address delegatedTo) {
            console.log("Delegated to:", delegatedTo);
            if (delegatedTo == address(0)) {
                console.log("WARNING: Not delegated!");
            }
        } catch {
            console.log("ERROR: Failed to get delegation");
        }
    }

    function checkOperators() internal view {
        console.log("=== Operator States ===");

        console.log("\n--- Aggregator Operator ---");
        checkOperatorState(AGG_OPERATOR, AVS_ADDRESS, 0);

        console.log("\n--- Executor 1 Operator ---");
        checkOperatorState(EXEC_OPERATOR_1, AVS_ADDRESS, 1);

        console.log("\n--- Executor 2 Operator ---");
        checkOperatorState(EXEC_OPERATOR_2, AVS_ADDRESS, 1);

        console.log("\n--- Executor 3 Operator ---");
        checkOperatorState(EXEC_OPERATOR_3, AVS_ADDRESS, 1);

        console.log("\n--- Executor 4 Operator ---");
        checkOperatorState(EXEC_OPERATOR_4, AVS_ADDRESS, 1);

        console.log("");
    }

    function checkOperatorState(address operator, address avs, uint32 operatorSetId) internal view {
        console.log("Address:", operator);

        OperatorSet memory operatorSet = OperatorSet({
            avs: avs,
            id: operatorSetId
        });

        // Check if operator is registered
        try DELEGATION_MANAGER.isOperator(operator) returns (bool isOperator) {
            console.log("Is EigenLayer operator:", isOperator);
        } catch {
            console.log("ERROR: Failed to check operator registration");
        }

        // Check allocated magnitude for STETH strategy
        try ALLOCATION_MANAGER.getAllocation(operator, operatorSet, STETH_STRATEGY) returns (IAllocationManager.Allocation memory allocation) {
            console.log("STETH allocation - current magnitude:", allocation.currentMagnitude);
            console.log("STETH allocation - pending diff:", uint256(uint128(allocation.pendingDiff)));
            console.log("STETH allocation - effect block:", allocation.effectBlock);
        } catch {
            console.log("ERROR: Failed to get STETH allocation");
        }

        // Check allocated magnitude for WETH strategy
        try ALLOCATION_MANAGER.getAllocation(operator, operatorSet, WETH_STRATEGY) returns (IAllocationManager.Allocation memory allocation) {
            console.log("WETH allocation - current magnitude:", allocation.currentMagnitude);
            console.log("WETH allocation - pending diff:", uint256(uint128(allocation.pendingDiff)));
            console.log("WETH allocation - effect block:", allocation.effectBlock);
        } catch {
            console.log("ERROR: Failed to get WETH allocation");
        }

        // Check encumbered magnitude for STETH
        try ALLOCATION_MANAGER.getEncumberedMagnitude(operator, STETH_STRATEGY) returns (uint64 encumbered) {
            console.log("STETH encumbered magnitude:", encumbered);
        } catch {
            console.log("ERROR: Failed to get STETH encumbered magnitude");
        }

        // Check encumbered magnitude for WETH
        try ALLOCATION_MANAGER.getEncumberedMagnitude(operator, WETH_STRATEGY) returns (uint64 encumbered) {
            console.log("WETH encumbered magnitude:", encumbered);
        } catch {
            console.log("ERROR: Failed to get WETH encumbered magnitude");
        }
    }

    function checkBN254TableCalculator() internal view {
        console.log("=== BN254 Table Calculator ===");
        console.log("Address:", address(BN254_TABLE_CALCULATOR));

        // Check for executor operator set (opset 1)
        OperatorSet memory executorOpSet = OperatorSet({
            avs: AVS_ADDRESS,
            id: 1
        });

        console.log("\n--- Executor Operator Set (ID: 1) ---");

        // Try to get operator infos
        try IBN254TableCalculator(BN254_TABLE_CALCULATOR).getOperatorInfos(executorOpSet) returns (
            IOperatorTableCalculatorTypes.BN254OperatorInfo[] memory operatorInfos
        ) {
            console.log("Number of operators with BN254 keys:", operatorInfos.length);

            if (operatorInfos.length == 0) {
                console.log("WARNING: No operators found in BN254TableCalculator!");
            } else {
                for (uint256 i = 0; i < operatorInfos.length; i++) {
                    console.log("\n  Operator", i);
                    console.log("    Pubkey X:", operatorInfos[i].pubkey.X);
                    console.log("    Pubkey Y:", operatorInfos[i].pubkey.Y);
                    console.log("    Number of weight types:", operatorInfos[i].weights.length);
                    for (uint256 j = 0; j < operatorInfos[i].weights.length; j++) {
                        console.log("      Weight[", j, "]:", operatorInfos[i].weights[j]);
                        if (operatorInfos[i].weights[j] == 0) {
                            console.log("        WARNING: Zero weight!");
                        }
                    }
                }
            }
        } catch {
            console.log("ERROR: Failed to get operator infos from BN254TableCalculator");
        }

        // Try to calculate the full operator table
        console.log("\n--- Calculated Operator Table ---");
        try IBN254TableCalculator(BN254_TABLE_CALCULATOR).calculateOperatorTable(executorOpSet) returns (
            IOperatorTableCalculatorTypes.BN254OperatorSetInfo memory opSetInfo
        ) {
            console.log("Operator Info Tree Root:");
            console.logBytes32(opSetInfo.operatorInfoTreeRoot);
            console.log("Number of operators:", opSetInfo.numOperators);
            console.log("Aggregate Pubkey X:", opSetInfo.aggregatePubkey.X);
            console.log("Aggregate Pubkey Y:", opSetInfo.aggregatePubkey.Y);
            console.log("Number of weight types:", opSetInfo.totalWeights.length);
            for (uint256 i = 0; i < opSetInfo.totalWeights.length; i++) {
                console.log("  Total Weight[", i, "]:", opSetInfo.totalWeights[i]);
                if (opSetInfo.totalWeights[i] == 0) {
                    console.log("    WARNING: Zero total weight!");
                }
            }
        } catch {
            console.log("ERROR: Failed to calculate operator table");
        }

        console.log("");
    }

    function checkBN254CertificateVerifier() internal view {
        console.log("=== BN254 Certificate Verifier ===");

        // Try to find the certificate verifier address from the CrossChainRegistry
        // The verifier is typically configured per operator set
        OperatorSet memory executorOpSet = OperatorSet({
            avs: AVS_ADDRESS,
            id: 1
        });

        // Try to get the verifier address from known deployment addresses
        // The BN254CertificateVerifier is typically deployed at a fixed address on mainnet
        // For now, we'll try the most common mainnet address
        address[] memory potentialVerifiers = new address[](2);
        potentialVerifiers[0] = 0x3F55654b2b2b86bB11bE2f72657f9C33bf88120A; // Common mainnet address
        address verifierAddress = address(0);

        // Try each potential address to see if it responds correctly
        for (uint256 i = 0; i < potentialVerifiers.length && verifierAddress == address(0); i++) {
            if (potentialVerifiers[i] == address(0)) continue;

            try IBN254CertificateVerifier(potentialVerifiers[i]).getOperatorSetInfo(
                executorOpSet,
                0 // timestamp 0 as a probe
            ) returns (IOperatorTableCalculatorTypes.BN254OperatorSetInfo memory) {
                verifierAddress = potentialVerifiers[i];
            } catch {
                // This address doesn't work, try next
            }
        }

        if (verifierAddress == address(0)) {
            console.log("WARNING: Could not find BN254CertificateVerifier address");
            console.log("The verifier may not be deployed yet or may be at a different address");
            console.log("");
            return;
        }

        console.log("Address:", verifierAddress);

        // Get the latest reference timestamp from the OperatorTableUpdater
        uint32 latestTimestamp;
        try OPERATOR_TABLE_UPDATER.getLatestReferenceTimestamp() returns (uint32 timestamp) {
            latestTimestamp = timestamp;
            console.log("Querying at reference timestamp:", latestTimestamp);
        } catch {
            console.log("ERROR: Failed to get latest reference timestamp");
            console.log("");
            return;
        }

        // Try to get operator set info for the executor operator set
        try IBN254CertificateVerifier(verifierAddress).getOperatorSetInfo(
            executorOpSet,
            latestTimestamp
        ) returns (IOperatorTableCalculatorTypes.BN254OperatorSetInfo memory opSetInfo) {
            console.log("\n--- Stored Operator Set Info ---");
            console.log("Operator Info Tree Root:");
            console.logBytes32(opSetInfo.operatorInfoTreeRoot);
            console.log("Number of operators:", opSetInfo.numOperators);
            console.log("Aggregate Pubkey X:", opSetInfo.aggregatePubkey.X);
            console.log("Aggregate Pubkey Y:", opSetInfo.aggregatePubkey.Y);
            console.log("Number of weight types:", opSetInfo.totalWeights.length);

            if (opSetInfo.numOperators == 0) {
                console.log("WARNING: No operators in certificate verifier!");
            }

            for (uint256 i = 0; i < opSetInfo.totalWeights.length; i++) {
                console.log("  Total Weight[", i, "]:", opSetInfo.totalWeights[i]);
                if (opSetInfo.totalWeights[i] == 0) {
                    console.log("    WARNING: Zero total weight in verifier!");
                }
            }
        } catch {
            console.log("ERROR: No operator set info found for this timestamp");
            console.log("The operator table may not have been transported yet");
        }

        console.log("");
    }
}
