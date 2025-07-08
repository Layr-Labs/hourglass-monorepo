// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {
    IAllocationManager,
    IAllocationManagerTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IDelegationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IDelegationManager.sol";
import {IAVSRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";
import {IStrategy} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";
import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract RegisterOperator is Script {
    // Eigenlayer Core Contracts
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x42583067658071247ec8CE0A516A58f682002d07);
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0xD4A7E1Bd8015057293f0D0A557088c286942e84b);

    function setUp() public {}

    function run(
        bytes32 operatorPrivateKey,
        uint32 allocationDelay,
        string memory metadataURI,
        address avs,
        uint32 operatorSetId,
        string memory socket
    ) public {
        // Load the private key from the environment variable
        address operator = vm.addr(uint256(operatorPrivateKey));

        vm.startBroadcast(uint256(operatorPrivateKey));
        console.log("Operator address:", operator);

        // 1. Register the operator
        // set the
        DELEGATION_MANAGER.registerAsOperator(address(0), allocationDelay, metadataURI);
        console.log("Operator registered:", operator, DELEGATION_MANAGER.isOperator(operator));

        // 2. Register for operator set
        uint32[] memory operatorSetIds = new uint32[](1);
        operatorSetIds[0] = operatorSetId;
        ALLOCATION_MANAGER.registerForOperatorSets(
            operator,
            IAllocationManagerTypes.RegisterParams({
                avs: avs,
                operatorSetIds: operatorSetIds,
                data: abi.encode(socket) // AVSRegistrarWithSocket expects just the socket string
            })
        );
        console.log(
            "Operator registered to operator set:",
            avs,
            operatorSetId,
            ALLOCATION_MANAGER.isMemberOfOperatorSet(operator, OperatorSet(avs, operatorSetId))
        );

        vm.stopBroadcast();
    }
}
