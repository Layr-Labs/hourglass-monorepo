// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {Test} from "forge-std/Test.sol";

import {
    ICrossChainRegistry,
    ICrossChainRegistryTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/ICrossChainRegistry.sol";
import {IBN254TableCalculator} from "@eigenlayer-middleware/src/interfaces/IBN254TableCalculator.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IStrategy} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";
import {
    IAllocationManager,
    IAllocationManagerTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IDelegationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IDelegationManager.sol";
import {IStrategyManager} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategyManager.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ISignatureUtilsMixinTypes} from "@eigenlayer-contracts/src/contracts/interfaces/ISignatureUtilsMixin.sol";

interface IWETH {
    function deposit() external payable;
}

interface IStETH {
    function submit(
        address _referral
    ) external payable returns (uint256);
}

/**
 * @title StakeWithStrategies
 * @notice Parameterized script for staking tokens across different networks (Mainnet, Sepolia, etc.)
 * @dev Follows the same pattern as StakeStuff.s.sol but accepts contract addresses as parameters
 */
contract StakeWithStrategies is Script {
    // Contract addresses to be passed as parameters
    IAllocationManager public allocationManager;
    IDelegationManager public delegationManager;
    IStrategyManager public strategyManager;

    // Strategies to be passed as parameters
    IStrategy public strategyWeth;
    IStrategy public strategySteth;

    uint64 public magnitudeToSet = 1e18;

    /**
     * @notice Main entry point with parameterized addresses
     * @param _allocationManager Address of the AllocationManager contract
     * @param _delegationManager Address of the DelegationManager contract
     * @param _strategyManager Address of the StrategyManager contract
     * @param _strategyWeth Address of the WETH Strategy contract
     * @param _strategySteth Address of the stETH Strategy contract
     */
    function run(
        address _allocationManager,
        address _delegationManager,
        address _strategyManager,
        address _strategyWeth,
        address _strategySteth
    ) public {
        // Initialize contract references
        allocationManager = IAllocationManager(_allocationManager);
        delegationManager = IDelegationManager(_delegationManager);
        strategyManager = IStrategyManager(_strategyManager);
        strategyWeth = IStrategy(_strategyWeth);
        strategySteth = IStrategy(_strategySteth);

        // Load staker private keys from environment
        uint256 aggStakerPrivateKey = vm.envUint("AGG_STAKER_PRIVATE_KEY");
        address aggStakerAddr = vm.addr(aggStakerPrivateKey);

        uint256 execStakerPrivateKey = vm.envUint("EXEC_STAKER_PRIVATE_KEY");
        address execStakerAddr = vm.addr(execStakerPrivateKey);

        uint256 execStaker2PrivateKey = vm.envUint("EXEC_STAKER2_PRIVATE_KEY");
        address execStaker2Addr = vm.addr(execStaker2PrivateKey);

        uint256 execStaker3PrivateKey = vm.envUint("EXEC_STAKER3_PRIVATE_KEY");
        address execStaker3Addr = vm.addr(execStaker3PrivateKey);

        uint256 execStaker4PrivateKey = vm.envUint("EXEC_STAKER4_PRIVATE_KEY");
        address execStaker4Addr = vm.addr(execStaker4PrivateKey);

        IERC20 wethToken = strategyWeth.underlyingToken();
        IERC20 stethToken = strategySteth.underlyingToken();

        // Stake WETH for aggregator
        console.log("Staking WETH for aggregator staker:", aggStakerAddr);
        vm.deal(aggStakerAddr, 100_000e18);
        vm.startBroadcast(aggStakerPrivateKey);
        IWETH(address(wethToken)).deposit{value: 20e18}();
        wethToken.approve(address(strategyManager), type(uint256).max);
        strategyManager.depositIntoStrategy(strategyWeth, wethToken, 10 ether);
        vm.stopBroadcast();

        uint256 balance = IERC20(wethToken).balanceOf(aggStakerAddr);
        console.log("WETH balance for aggregator staker:", balance);
        uint256 depositedAmount = strategyManager.stakerDepositShares(aggStakerAddr, strategyWeth);
        console.log("Aggregator staker deposit shares in STRATEGY_WETH:", depositedAmount);

        // Stake stETH for executor 1
        console.log("Staking stETH for executor 1 staker:", execStakerAddr);
        vm.deal(execStakerAddr, 100_000e18);
        vm.startBroadcast(execStakerPrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(strategyManager), type(uint256).max);
        strategyManager.depositIntoStrategy(strategySteth, stethToken, 10 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStakerAddr);
        console.log("stETH balance for executor 1 staker:", balance);
        depositedAmount = strategyManager.stakerDepositShares(execStakerAddr, strategySteth);
        console.log("Executor 1 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        // Stake stETH for executor 2
        console.log("Staking stETH for executor 2 staker:", execStaker2Addr);
        vm.deal(execStaker2Addr, 100_000e18);
        vm.startBroadcast(execStaker2PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(strategyManager), type(uint256).max);
        strategyManager.depositIntoStrategy(strategySteth, stethToken, 10 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker2Addr);
        console.log("stETH balance for executor 2 staker:", balance);
        depositedAmount = strategyManager.stakerDepositShares(execStaker2Addr, strategySteth);
        console.log("Executor 2 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        // Stake stETH for executor 3
        console.log("Staking stETH for executor 3 staker:", execStaker3Addr);
        vm.deal(execStaker3Addr, 100_000e18);
        vm.startBroadcast(execStaker3PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(strategyManager), type(uint256).max);
        strategyManager.depositIntoStrategy(strategySteth, stethToken, 10 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker3Addr);
        console.log("stETH balance for executor 3 staker:", balance);
        depositedAmount = strategyManager.stakerDepositShares(execStaker3Addr, strategySteth);
        console.log("Executor 3 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        // Stake stETH for executor 4
        console.log("Staking stETH for executor 4 staker:", execStaker4Addr);
        vm.deal(execStaker4Addr, 100_000e18);
        vm.startBroadcast(execStaker4PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(strategyManager), type(uint256).max);
        strategyManager.depositIntoStrategy(strategySteth, stethToken, 10 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker4Addr);
        console.log("stETH balance for executor 4 staker:", balance);
        depositedAmount = strategyManager.stakerDepositShares(execStaker4Addr, strategySteth);
        console.log("Executor 4 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        console.log("All staking operations completed successfully!");
        console.log("Stake weights summary:");
        console.log("- Aggregator: 10 ETH (WETH)");
        console.log("- Executor 1: 10 ETH (stETH)");
        console.log("- Executor 2: 10 ETH (stETH)");
        console.log("- Executor 3: 10 ETH (stETH)");
        console.log("- Executor 4: 10 ETH (stETH)");
    }
}