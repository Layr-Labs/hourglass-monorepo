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
import {IKeyRegistrar, IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract RegisterOperator is Script {
    // Eigenlayer Core Contracts (Sepolia Testnet)
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x42583067658071247ec8CE0A516A58f682002d07);
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0xD4A7E1Bd8015057293f0D0A557088c286942e84b);
    IKeyRegistrar public KEY_REGISTRAR = IKeyRegistrar(0xA4dB30D08d8bbcA00D40600bee9F029984dB162a);

    function setUp() public {}

    function run(
        bytes32 operatorPrivateKey,
        bytes32 systemPrivateKey,
        uint32 allocationDelay,
        string memory metadataURI,
        address avsAddr,
        uint32 operatorSetId,
        string memory socket
    ) public {
        // Derive addresses from private keys
        address operator = vm.addr(uint256(operatorPrivateKey));
        address systemKey = vm.addr(uint256(systemPrivateKey));

        vm.startBroadcast(uint256(operatorPrivateKey));
        console.log("Operator address:", operator);
        console.log("System key address:", systemKey);

        // 1. Register the operator with DelegationManager
        DELEGATION_MANAGER.registerAsOperator(address(0), allocationDelay, metadataURI);
        console.log("Operator registered:", operator, DELEGATION_MANAGER.isOperator(operator));

        // 2. Register ECDSA system key with KeyRegistrar (required before registering for operator set)
        // Switch to system key for key registration
        OperatorSet memory operatorSet = OperatorSet({avs: avsAddr, id: operatorSetId});

        // Get the message hash for ECDSA key registration
        // The system key signs a message that registers itself for the operator
        bytes32 messageHash = KEY_REGISTRAR.getECDSAKeyRegistrationMessageHash(operator, operatorSet, systemKey);

        // Sign the message hash with the system private key
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(uint256(systemPrivateKey), messageHash);
        bytes memory signature = abi.encodePacked(r, s, v);

        // Register the system ECDSA key for the operator
        bytes memory pubKey = abi.encodePacked(systemKey);
        KEY_REGISTRAR.registerKey(operator, operatorSet, pubKey, signature);
        console.log("System ECDSA key registered for operator:", operator, "key:", systemKey);

        // 3. Register for operator set
        uint32[] memory operatorSetIds = new uint32[](1);
        operatorSetIds[0] = operatorSetId;
        ALLOCATION_MANAGER.registerForOperatorSets(
            operator,
            IAllocationManagerTypes.RegisterParams({
                avs: avsAddr, operatorSetIds: operatorSetIds, data: abi.encode(socket)
            })
        );
        console.log(
            "Operator registered to operator set:",
            avsAddr,
            operatorSetId,
            ALLOCATION_MANAGER.isMemberOfOperatorSet(operator, OperatorSet(avsAddr, operatorSetId))
        );

        vm.stopBroadcast();
    }
}
