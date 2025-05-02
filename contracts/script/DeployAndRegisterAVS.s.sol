// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {
    IAllocationManager,
    IAllocationManagerTypes
} from "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IAVSRegistrar} from "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";
import {IPermissionController} from
    "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IPermissionController.sol";
import {IStrategy} from "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";

import {TaskAVSRegistrar} from "src/avs/l1-contracts/TaskAVSRegistrar.sol";

contract DeployAndRegisterAVS is Script {
    // Eigenlayer Core Contracts
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);
    IPermissionController public PERMISSION_CONTROLLER =
        IPermissionController(0x25E5F8B1E7aDf44518d35D5B2271f114e081f0E5);

    // Eigenlayer Strategies
    IStrategy public STRATEGY_EIGEN = IStrategy(0xaCB55C530Acdb2849e6d4f36992Cd8c9D50ED8F7);
    IStrategy public STRATEGY_STETH = IStrategy(0x93c4b944D05dfe6df7645A86cd2206016c51564D);

    function setUp() public {}

    function run(
        string memory metadataURI
    ) public {
        // Load the private key from the environment variable
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY_DEPLOYER");
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");
        address avs = vm.addr(avsPrivateKey);

        // 1. Deploy the TaskAVSRegistrar middleware contract
        vm.startBroadcast(deployerPrivateKey);
        TaskAVSRegistrar taskAVSRegistrar = new TaskAVSRegistrar(avs, ALLOCATION_MANAGER);
        console.log("TaskAVSRegistrar deployed to:", address(taskAVSRegistrar));
        vm.stopBroadcast();

        // 2. Register the AVS
        vm.startBroadcast(avsPrivateKey);
        ALLOCATION_MANAGER.updateAVSMetadataURI(avs, metadataURI);
        ALLOCATION_MANAGER.setAVSRegistrar(avs, IAVSRegistrar(taskAVSRegistrar));

        // 3. Create the operator sets
        IStrategy[] memory strategies = new IStrategy[](2);
        strategies[0] = STRATEGY_EIGEN;
        strategies[1] = STRATEGY_STETH;
        IAllocationManagerTypes.CreateSetParams[] memory createOperatorSetParams =
            new IAllocationManagerTypes.CreateSetParams[](2);
        createOperatorSetParams[0] = IAllocationManagerTypes.CreateSetParams({operatorSetId: 0, strategies: strategies});
        createOperatorSetParams[1] = IAllocationManagerTypes.CreateSetParams({operatorSetId: 1, strategies: strategies});
        ALLOCATION_MANAGER.createOperatorSets(avs, createOperatorSetParams);

        vm.stopBroadcast();
    }
}
