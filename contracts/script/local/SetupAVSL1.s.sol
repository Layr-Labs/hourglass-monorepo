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
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x42583067658071247ec8CE0A516A58f682002d07);

    // Eigenlayer Strategies
    IStrategy public STRATEGY_WETH = IStrategy(0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc);
    IStrategy public STRATEGY_STETH = IStrategy(0x8b29d91e67b013e855EaFe0ad704aC4Ab086a574);

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
        strategies[0] = STRATEGY_WETH;
        strategies[1] = STRATEGY_STETH;
        IAllocationManagerTypes.CreateSetParams[] memory createOperatorSetParams =
            new IAllocationManagerTypes.CreateSetParams[](2);

        IStrategy[] memory opsetZero = new IStrategy[](1);
        opsetZero[0] = STRATEGY_WETH;
        IStrategy[] memory opsetOne = new IStrategy[](1);
        opsetOne[0] = STRATEGY_STETH;

        createOperatorSetParams[0] = IAllocationManagerTypes.CreateSetParams({operatorSetId: 0, strategies: opsetZero});
        createOperatorSetParams[1] = IAllocationManagerTypes.CreateSetParams({operatorSetId: 1, strategies: opsetOne});

        ALLOCATION_MANAGER.createOperatorSets(avs, createOperatorSetParams);
        console.log("Operator sets created: ", ALLOCATION_MANAGER.getOperatorSetCount(avs));

        vm.stopBroadcast();
    }
}
