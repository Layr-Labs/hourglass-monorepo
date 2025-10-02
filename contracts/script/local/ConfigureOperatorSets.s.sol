// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import {IKeyRegistrar, IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract ConfigureOperatorSets is Script {
    // Mainnet KeyRegistrar address
    IKeyRegistrar constant KEY_REGISTRAR = IKeyRegistrar(0x54f4bC6bDEbe479173a2bbDc31dD7178408A57A4);

    function run(
        address avsAddress
    ) external {
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");

        console.log("AVS Account:", avsAddress);
        console.log("Key Registrar:", address(KEY_REGISTRAR));

        vm.startBroadcast(avsPrivateKey);

        // Use the AVS account address for the operator sets
        OperatorSet memory operatorSet0 = OperatorSet({
            avs: avsAddress, // Use the AVS account address
            id: 0
        });

        console.log("Configuring operator set 0 for ECDSA...");
        KEY_REGISTRAR.configureOperatorSet(operatorSet0, IKeyRegistrarTypes.CurveType.ECDSA);

        // Continue with other operator sets...
        OperatorSet memory operatorSet1 = OperatorSet({avs: avsAddress, id: 1});

        console.log("Configuring operator set 1 for BN254...");
        KEY_REGISTRAR.configureOperatorSet(operatorSet1, IKeyRegistrarTypes.CurveType.BN254);

        vm.stopBroadcast();

        console.log("Successfully configured operator sets");
    }
}
