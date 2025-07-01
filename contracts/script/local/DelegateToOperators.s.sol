// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {Test} from "forge-std/Test.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IStrategy} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";
import {
    IAllocationManager,
    IAllocationManagerTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IDelegationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IDelegationManager.sol";
import {IStrategyManager} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategyManager.sol";
import {ISignatureUtilsMixinTypes} from "@eigenlayer-contracts/src/contracts/interfaces/ISignatureUtilsMixin.sol";

contract DelegateToOperators is Script {
    // Constants
    IAllocationManager public constant ALLOCATION_MANAGER =
        IAllocationManager(0x42583067658071247ec8CE0A516A58f682002d07);
    IDelegationManager public constant DELEGATION_MANAGER =
        IDelegationManager(0xD4A7E1Bd8015057293f0D0A557088c286942e84b);

    IStrategy public constant STRATEGY_WETH = IStrategy(0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc);
    IStrategy public constant STRATEGY_STETH = IStrategy(0x8b29d91e67b013e855EaFe0ad704aC4Ab086a574);

    uint64 public constant MAGNITUDE_TO_SET = 1e18;

    // State variables to reduce stack depth
    uint256 private aggStakerPrivateKey;
    uint256 private execStakerPrivateKey;
    uint256 private aggregatorPrivateKey;
    uint256 private executorPrivateKey;

    address private aggStakerAddr;
    address private execStakerAddr;
    address private aggregatorAddr;
    address private executorAddr;

    function setUp() public {}

    function run(
        address avsAddr
    ) public {
        // Load all private keys and addresses
        _loadCredentials();

        // Delegate and allocate for aggregator
        _setupAggregator(avsAddr);

        // Delegate and allocate for executor
        _setupExecutor(avsAddr);
    }

    function _loadCredentials() private {
        aggStakerPrivateKey = vm.envUint("AGG_STAKER_PRIVATE_KEY");
        aggStakerAddr = vm.addr(aggStakerPrivateKey);

        execStakerPrivateKey = vm.envUint("EXEC_STAKER_PRIVATE_KEY");
        execStakerAddr = vm.addr(execStakerPrivateKey);

        aggregatorPrivateKey = vm.envUint("AGGREGATOR_PRIVATE_KEY");
        aggregatorAddr = vm.addr(aggregatorPrivateKey);

        executorPrivateKey = vm.envUint("EXECUTOR_PRIVATE_KEY");
        executorAddr = vm.addr(executorPrivateKey);
    }

    function _setupAggregator(
        address avsAddr
    ) private {
        // Delegate to aggregator
        ISignatureUtilsMixinTypes.SignatureWithExpiry memory emptySignature;

        vm.startBroadcast(aggStakerPrivateKey);
        DELEGATION_MANAGER.delegateTo(aggregatorAddr, emptySignature, bytes32(0));
        vm.stopBroadcast();

        // Modify aggregator allocations
        vm.startBroadcast(aggregatorPrivateKey);
        _allocateToOperatorSet(
            aggregatorAddr,
            avsAddr,
            0, // Operator set ID for aggregator
            STRATEGY_WETH
        );
        vm.stopBroadcast();
    }

    function _setupExecutor(
        address avsAddr
    ) private {
        // Delegate to executor
        ISignatureUtilsMixinTypes.SignatureWithExpiry memory emptySignature;

        vm.startBroadcast(execStakerPrivateKey);
        DELEGATION_MANAGER.delegateTo(executorAddr, emptySignature, bytes32(0));
        vm.stopBroadcast();

        // Modify executor allocations
        vm.startBroadcast(executorPrivateKey);
        _allocateToOperatorSet(
            executorAddr,
            avsAddr,
            1, // Operator set ID for executor
            STRATEGY_STETH
        );
        vm.stopBroadcast();
    }

    function _allocateToOperatorSet(
        address operator,
        address avsAddr,
        uint32 operatorSetId,
        IStrategy strategy
    ) private {
        // Create strategies array
        IStrategy[] memory strategies = new IStrategy[](1);
        strategies[0] = strategy;

        // Create magnitudes array
        uint64[] memory magnitudes = new uint64[](1);
        magnitudes[0] = MAGNITUDE_TO_SET;

        // Create operator set
        OperatorSet memory operatorSet = OperatorSet({avs: avsAddr, id: operatorSetId});

        // Create allocation params
        IAllocationManagerTypes.AllocateParams[] memory allocations = new IAllocationManagerTypes.AllocateParams[](1);
        allocations[0] = IAllocationManagerTypes.AllocateParams({
            operatorSet: operatorSet,
            strategies: strategies,
            newMagnitudes: magnitudes
        });

        // Modify allocations
        ALLOCATION_MANAGER.modifyAllocations(operator, allocations);
    }
}
