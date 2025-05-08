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

contract RegisterOperator is Script {
    // Eigenlayer Core Contracts
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A);

    function setUp() public {}

    function run(uint32 allocatonDelay, string memory metadataURI, address avs, uint32 operatorSetId, bytes memory data) public {
        // Load the private key from the environment variable
        uint256 operatorPrivateKey = vm.envUint("PRIVATE_KEY_OPERATOR");
        address operator = vm.addr(operatorPrivateKey);

        vm.startBroadcast(operatorPrivateKey);
        console.log("Operator address:", operator);

        // 1. Register the operator
        DELEGATION_MANAGER.registerAsOperator(operator, allocatonDelay, metadataURI);
        bool isOperator = DELEGATION_MANAGER.isOperator(operator);
        console.log("Operator registered:", operator, isOperator);

        // 2. Register to operator set
        uint32[] memory operatorSetIds = new uint32[](1);
        operatorSetIds[0] = operatorSetId;
        ALLOCATION_MANAGER.registerForOperatorSets(operator, IAllocationManagerTypes.RegisterParams({
            avs: avs,
            operatorSetIds: operatorSetIds,
            data: data
        }));
        console.log("Operator registered to operator set:", operatorSetId);

        vm.stopBroadcast();
    }
}