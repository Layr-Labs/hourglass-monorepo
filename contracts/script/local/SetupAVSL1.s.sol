// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {
    IAllocationManager,
    IAllocationManagerTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IAVSRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";
import {IStrategy} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";

contract SetupAVSL1 is Script {
    // Eigenlayer Core Contracts
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);

    // Eigenlayer Strategies
    IStrategy public STRATEGY_EIGEN = IStrategy(0xaCB55C530Acdb2849e6d4f36992Cd8c9D50ED8F7);
    IStrategy public STRATEGY_STETH = IStrategy(0x93c4b944D05dfe6df7645A86cd2206016c51564D);

    function setUp() public {}

    function run(
        address taskAVSRegistrar
    ) public {
        // Load the private key from the environment variable
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");
        address avs = vm.addr(avsPrivateKey);

        vm.startBroadcast(avsPrivateKey);
        console.log("AVS address:", avs);

        // 1. Update the AVS metadata URI
        ALLOCATION_MANAGER.updateAVSMetadataURI(avs, "Test AVS");
        console.log("AVS metadata URI updated: Test AVS");

        // 2. Set the AVS Registrar
        ALLOCATION_MANAGER.setAVSRegistrar(avs, IAVSRegistrar(taskAVSRegistrar));
        console.log("AVS Registrar set:", address(ALLOCATION_MANAGER.getAVSRegistrar(avs)));

        // 3. Create the operator sets
        IStrategy[] memory strategies = new IStrategy[](2);
        strategies[0] = STRATEGY_EIGEN;
        strategies[1] = STRATEGY_STETH;
        IAllocationManagerTypes.CreateSetParams[] memory createOperatorSetParams =
            new IAllocationManagerTypes.CreateSetParams[](2);
        createOperatorSetParams[0] = IAllocationManagerTypes.CreateSetParams({operatorSetId: 0, strategies: strategies});
        createOperatorSetParams[1] = IAllocationManagerTypes.CreateSetParams({operatorSetId: 1, strategies: strategies});
        ALLOCATION_MANAGER.createOperatorSets(avs, createOperatorSetParams);
        console.log("Operator sets created: ", ALLOCATION_MANAGER.getOperatorSetCount(avs));

        vm.stopBroadcast();
    }
}
