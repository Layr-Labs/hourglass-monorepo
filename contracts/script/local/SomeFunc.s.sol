// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import "forge-std/Test.sol";
import {IKeyRegistrar, IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

// Totally a scratch pad for validating ones sanity...
contract SomeFunc is Script {
    function run() external {
        uint256 avsPrivateKey = 0x90a7b1bcc84977a8b008fea51da40ad7e58b844095b13518f575ded17a4c67e4;

        // Mainnet KeyRegistrar address
        IKeyRegistrar KEY_REGISTRAR = IKeyRegistrar(0x54f4bC6bDEbe479173a2bbDc31dD7178408A57A4);

        address someAddr = 0x6B58f6762689DF33fe8fa3FC40Fb5a3089D3a8cc;
        bytes memory packedAddr = abi.encodePacked(someAddr);
        console.log("Packed bytes:", vm.toString(packedAddr));

        address operatorAddr = 0x6B58f6762689DF33fe8fa3FC40Fb5a3089D3a8cc;
        address avsAddr = 0xCE2Ac75bE2E0951F1F7B288c7a6A9BfB6c331DC4;

        bytes32 hash = KEY_REGISTRAR.getECDSAKeyRegistrationMessageHash(
            operatorAddr, OperatorSet({avs: avsAddr, id: 0}), operatorAddr
        );
        console.log("Hash:", vm.toString(hash));

        uint256 pk = 0x3dd7c381f27775d9945f0fcf5bb914484c4d01681824603c71dd762259f43214;
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(pk, hash);

        console.log("Signature:");
        console.log("v:   ", v);
        console.log("r:   ", vm.toString(r));
        console.log("s:   ", vm.toString(s));

        bytes memory packedSig = abi.encodePacked(r, s, v);
        console.log("Packed signature:", vm.toString(packedSig));
        // 0x94f6cc73a89b1304088f1348d150abc37483375c3a1ede44f7758aac3b8adb3d23e1caede86581b00003b327392d6031ccefe08e4f79913713a40b8cddb987c91b

        console.log("curve type", uint8(IKeyRegistrarTypes.CurveType.ECDSA));

        vm.startBroadcast(avsPrivateKey);
        KEY_REGISTRAR.configureOperatorSet(OperatorSet({avs: avsAddr, id: 0}), IKeyRegistrarTypes.CurveType.ECDSA);
        vm.stopBroadcast();
        vm.roll(block.number + 5);

        bytes memory pubKey = abi.encodePacked(operatorAddr);
        vm.startBroadcast(pk);
        KEY_REGISTRAR.registerKey(operatorAddr, OperatorSet({avs: avsAddr, id: 0}), pubKey, packedSig);

        vm.stopBroadcast();
    }
}
