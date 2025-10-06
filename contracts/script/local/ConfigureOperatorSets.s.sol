// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import {IKeyRegistrar, IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract ConfigureOperatorSets is Script {
    function setUp() public {}

    function run() external {
        // Get configuration from environment variables
        address keyRegistrarAddress = vm.envAddress("KEY_REGISTRAR_ADDRESS");
        address avsAddress = vm.envAddress("AVS_ADDRESS");
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");

        IKeyRegistrar KEY_REGISTRAR = IKeyRegistrar(keyRegistrarAddress);

        console.log("Configuring operator sets for AVS");
        console.log("  AVS Address:", avsAddress);
        console.log("  Key Registrar:", keyRegistrarAddress);

        vm.startBroadcast(avsPrivateKey);

        // Configure aggregator operator set (ID 0) for ECDSA
        OperatorSet memory operatorSet0 = OperatorSet({avs: avsAddress, id: 0});

        console.log("Configuring operator set 0 (aggregator) for ECDSA...");
        KEY_REGISTRAR.configureOperatorSet(operatorSet0, IKeyRegistrarTypes.CurveType.ECDSA);

        // Configure executor operator set (ID 1) for ECDSA
        OperatorSet memory operatorSet1 = OperatorSet({avs: avsAddress, id: 1});

        console.log("Configuring operator set 1 (executor) for ECDSA...");
        KEY_REGISTRAR.configureOperatorSet(operatorSet1, IKeyRegistrarTypes.CurveType.ECDSA);

        vm.stopBroadcast();

        console.log("Successfully configured operator sets");
    }
}
