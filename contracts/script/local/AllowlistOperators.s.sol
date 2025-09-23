// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IAllowlist} from "@eigenlayer-middleware/src/interfaces/IAllowlist.sol";

contract AllowlistOperators is Script {
    function setUp() public {}

    function run(
        address taskAVSRegistrar
    ) public {
        // Load the private keys from environment variables
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        address deployer = vm.addr(deployerPrivateKey);

        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");
        address avs = vm.addr(avsPrivateKey);

        uint256 aggregatorPrivateKey = vm.envUint("AGGREGATOR_PRIVATE_KEY");
        address aggregatorAddr = vm.addr(aggregatorPrivateKey);

        console.log("Deployer address (owner):", deployer);
        console.log("AVS address:", avs);
        console.log("TaskAVSRegistrar address:", taskAVSRegistrar);
        console.log("Aggregator address to allowlist:", aggregatorAddr);

        // Create the operator set for the aggregator (ID 0)
        OperatorSet memory aggregatorOperatorSet = OperatorSet({avs: avs, id: 0});

        // Add the aggregator operator to the allowlist using deployer key (who is the owner)
        vm.startBroadcast(deployerPrivateKey);

        IAllowlist(taskAVSRegistrar).addOperatorToAllowlist(aggregatorOperatorSet, aggregatorAddr);

        console.log("Added aggregator operator to allowlist for operator set 0");

        // Verify the operator is allowlisted
        bool isAllowed = IAllowlist(taskAVSRegistrar).isOperatorAllowed(aggregatorOperatorSet, aggregatorAddr);

        console.log("Aggregator allowlist status:", isAllowed);
        require(isAllowed, "Aggregator operator not properly allowlisted");

        vm.stopBroadcast();
    }
}
