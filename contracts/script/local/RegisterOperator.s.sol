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

import {ITaskAVSRegistrar, ITaskAVSRegistrarTypes} from "../../src/interfaces/avs/l1/ITaskAVSRegistrar.sol";

contract RegisterOperator is Script {
    // Eigenlayer Core Contracts
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A);

    function setUp() public {}

    function run(
        bytes32 operatorPrivateKey,
        uint32 allocationDelay,
        string memory metadataURI,
        address avs,
        uint32 operatorSetId,
        string memory socket,
        bytes memory pubkeyRegistrationParams
    ) public {
        // Load the private key from the environment variable
        address operator = vm.addr(uint256(operatorPrivateKey));

        vm.startBroadcast(uint256(operatorPrivateKey));
        console.log("Operator address:", operator);

        // 1. Register the operator
        DELEGATION_MANAGER.registerAsOperator(operator, allocationDelay, metadataURI);
        console.log("Operator registered:", operator, DELEGATION_MANAGER.isOperator(operator));

        // 2. Register for operator set
        uint32[] memory operatorSetIds = new uint32[](1);
        operatorSetIds[0] = operatorSetId;
        ALLOCATION_MANAGER.registerForOperatorSets(
            operator,
            IAllocationManagerTypes.RegisterParams({
                avs: avs,
                operatorSetIds: operatorSetIds,
                data: abi.encode(
                    ITaskAVSRegistrarTypes.OperatorRegistrationParams({
                        socket: socket,
                        pubkeyRegistrationParams: abi.decode(
                            pubkeyRegistrationParams, (ITaskAVSRegistrarTypes.PubkeyRegistrationParams)
                        )
                    })
                )
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
