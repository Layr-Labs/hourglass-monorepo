// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {ProxyAdmin} from "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import {TransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import {ITransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import { IAllocationManager, IAllocationManagerTypes } from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IAVSRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";
import {IStrategy} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";
import {IKeyRegistrar, IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {IPermissionController} from "@eigenlayer-contracts/src/contracts/interfaces/IPermissionController.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import { IReleaseManager } from "@eigenlayer-contracts/src/contracts/interfaces/IReleaseManager.sol";
import {ITaskAVSRegistrarBaseTypes} from "@eigenlayer-middleware/src/interfaces/ITaskAVSRegistrarBase.sol";
import {TaskAVSRegistrar} from "@project/l1-contracts/TaskAVSRegistrar.sol";
import { ICrossChainRegistry, ICrossChainRegistryTypes } from "@eigenlayer-contracts/src/contracts/interfaces/ICrossChainRegistry.sol";
import {IECDSATableCalculator} from "@eigenlayer-middleware/src/interfaces/IECDSATableCalculator.sol";

contract ManuallySetupAvsMainnet is Script {
    // Eigenlayer Core Contracts
    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);
    IKeyRegistrar constant KEY_REGISTRAR = IKeyRegistrar(0x54f4bC6bDEbe479173a2bbDc31dD7178408A57A4);
    IPermissionController constant PERMISSION_CONTROLLER = IPermissionController(0x25E5F8B1E7aDf44518d35D5B2271f114e081f0E5);
    IReleaseManager public RELEASE_MANAGER = IReleaseManager(0xeDA3CAd031c0cf367cF3f517Ee0DC98F9bA80C8F);
    ICrossChainRegistry public CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0x9376A5863F2193cdE13e1aB7c678F22554E2Ea2b);
    IECDSATableCalculator public ECDSA_TABLE_CALCULATOR = IECDSATableCalculator(0xA933CB4cbD0C4C208305917f56e0C3f51ad713Fa);

    IStrategy public STRATEGY_EIGEN = IStrategy(0x8E93249a6C37a32024756aaBd813E6139b17D1d5);


    function setUp() public {}

    function run(
        uint32 aggregatorOperatorSetId,
        uint32 executorOperatorSetId
    ) public {
        address taskAVSRegistrar = deployRegistrarContract(aggregatorOperatorSetId, executorOperatorSetId);

        createAvsInProtocol(taskAVSRegistrar);
        registerForMultichain();
    }

    function createAvsInProtocol(address taskAVSRegistrar) public {
        uint256 avsPrivateKey = vm.envUint("AVS_PRIVATE_KEY");
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
        strategies[0] = STRATEGY_EIGEN;
        strategies[1] = STRATEGY_EIGEN;
        IAllocationManagerTypes.CreateSetParams[] memory createOperatorSetParams =
                    new IAllocationManagerTypes.CreateSetParams[](2);

        IStrategy[] memory opsetZero = new IStrategy[](1);
        opsetZero[0] = STRATEGY_EIGEN;
        IStrategy[] memory opsetOne = new IStrategy[](1);
        opsetOne[0] = STRATEGY_EIGEN;

        createOperatorSetParams[0] = IAllocationManagerTypes.CreateSetParams({operatorSetId: 0, strategies: opsetZero});
        createOperatorSetParams[1] = IAllocationManagerTypes.CreateSetParams({operatorSetId: 1, strategies: opsetOne});

        ALLOCATION_MANAGER.createOperatorSets(avs, createOperatorSetParams);
        console.log("Operator sets created: ", ALLOCATION_MANAGER.getOperatorSetCount(avs));

        // Configure operator sets in the keyRegistrar
        // Use the AVS account address for the operator sets
        console.log("Configuring operator set 0 for ECDSA...");
        OperatorSet memory operatorSet0 = OperatorSet({ avs: avs, id: 0 });
        KEY_REGISTRAR.configureOperatorSet(operatorSet0, IKeyRegistrarTypes.CurveType.ECDSA);

        string memory opset0Uri = "http://operator-set-0.com";
        RELEASE_MANAGER.publishMetadataURI(operatorSet0, opset0Uri);

        // Continue with other operator sets...
        console.log("Configuring operator set 1 for ECDSA...");
        OperatorSet memory operatorSet1 = OperatorSet({avs: avs, id: 1});
        KEY_REGISTRAR.configureOperatorSet(operatorSet1, IKeyRegistrarTypes.CurveType.ECDSA);

        string memory opset1Uri = "http://operator-set-1.com";
        RELEASE_MANAGER.publishMetadataURI(operatorSet1, opset1Uri);

        vm.stopBroadcast();
    }

    function registerForMultichain() public {
        uint256 avsPrivateKey = vm.envUint("AVS_PRIVATE_KEY");
        address avs = vm.addr(avsPrivateKey);

        vm.startBroadcast(avsPrivateKey);

        OperatorSet memory operatorSet0 = OperatorSet({ avs: avs, id: 0 });
        OperatorSet memory operatorSet1 = OperatorSet({ avs: avs, id: 1 });

        ICrossChainRegistryTypes.OperatorSetConfig memory config = ICrossChainRegistryTypes.OperatorSetConfig({
            owner: avs,
            maxStalenessPeriod: 259_200 // 3 days
        });

        CROSS_CHAIN_REGISTRY.createGenerationReservation(operatorSet0, ECDSA_TABLE_CALCULATOR, config);
        CROSS_CHAIN_REGISTRY.createGenerationReservation(operatorSet1, ECDSA_TABLE_CALCULATOR, config);

        vm.stopBroadcast();
    }

    function deployRegistrarContract(uint32 aggregatorOperatorSetId, uint32 executorOperatorSetId) public returns (address) {
        uint256 avsPrivateKey = vm.envUint("AVS_PRIVATE_KEY");
        address avs = vm.addr(avsPrivateKey);

        vm.startBroadcast(avsPrivateKey);

        // Create initial config
        uint32[] memory executorOperatorSetIds = new uint32[](1);
        executorOperatorSetIds[0] = executorOperatorSetId;
        ITaskAVSRegistrarBaseTypes.AvsConfig memory initialConfig = ITaskAVSRegistrarBaseTypes.AvsConfig({
            aggregatorOperatorSetId: aggregatorOperatorSetId,
            executorOperatorSetIds: executorOperatorSetIds
        });

        // Deploy ProxyAdmin
        ProxyAdmin proxyAdmin = new ProxyAdmin();
        console.log("ProxyAdmin deployed to:", address(proxyAdmin));

        // Deploy implementation
        TaskAVSRegistrar taskAVSRegistrarImpl = new TaskAVSRegistrar(ALLOCATION_MANAGER, KEY_REGISTRAR, PERMISSION_CONTROLLER);
        console.log("TaskAVSRegistrar implementation deployed to:", address(taskAVSRegistrarImpl));

        // Deploy proxy with initialization
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(taskAVSRegistrarImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(TaskAVSRegistrar.initialize.selector, avs, avs, initialConfig)
        );
        console.log("TaskAVSRegistrar proxy deployed to:", address(proxy));

        // Transfer ProxyAdmin ownership to avs (or a multisig in production)
        proxyAdmin.transferOwnership(avs);

        vm.stopBroadcast();

        return address(proxy);
    }
}
