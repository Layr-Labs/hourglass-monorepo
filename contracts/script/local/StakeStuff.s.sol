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

contract StakeStuff is Script {
    // Eigenlayer Core Contracts (Mainnet)
    ICrossChainRegistry public CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0x9376A5863F2193cdE13e1aB7c678F22554E2Ea2b);

    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39);
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A);
    IStrategyManager public STRATEGY_MANAGER = IStrategyManager(0x858646372CC42E1A627fcE94aa7A7033e7CF075A);

    // Eigenlayer Strategies (Mainnet)
    IStrategy public STRATEGY_WETH = IStrategy(0x0Fe4F44beE93503346A3Ac9EE5A26b130a5796d6);
    IStrategy public STRATEGY_STETH = IStrategy(0x93c4b944D05dfe6df7645A86cd2206016c51564D);

    uint64 public magnitudeToSet = 1e18;

    function setUp() public {}

    function run() public {
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

        IERC20 wethToken = STRATEGY_WETH.underlyingToken();
        IERC20 stethToken = STRATEGY_STETH.underlyingToken();

        vm.deal(aggStakerAddr, 100_000e18);
        vm.startBroadcast(aggStakerPrivateKey);
        IWETH(address(wethToken)).deposit{value: 20e18}();
        wethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_WETH, wethToken, 10 ether);
        vm.stopBroadcast();

        uint256 balance = IERC20(wethToken).balanceOf(aggStakerAddr);
        console.log("WETH balance for aggregator staker:", balance);
        uint256 depositedAmount = STRATEGY_MANAGER.stakerDepositShares(aggStakerAddr, STRATEGY_WETH);
        console.log("Aggregator staker deposit shares in STRATEGY_WETH:", depositedAmount);

        vm.deal(execStakerAddr, 100_000e18);
        vm.startBroadcast(execStakerPrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, 10 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStakerAddr);
        console.log("stETH balance for executor 1 staker:", balance);
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStakerAddr, STRATEGY_STETH);
        console.log("Executor 1 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        vm.deal(execStaker2Addr, 100_000e18);
        vm.startBroadcast(execStaker2PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, 10 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker2Addr);
        console.log("stETH balance for executor 2 staker:", balance);
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStaker2Addr, STRATEGY_STETH);
        console.log("Executor 2 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        vm.deal(execStaker3Addr, 100_000e18);
        vm.startBroadcast(execStaker3PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, 10 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker3Addr);
        console.log("stETH balance for executor 3 staker:", balance);
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStaker3Addr, STRATEGY_STETH);
        console.log("Executor 3 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        vm.deal(execStaker4Addr, 100_000e18);
        vm.startBroadcast(execStaker4PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, 10 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker4Addr);
        console.log("stETH balance for executor 4 staker:", balance);
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStaker4Addr, STRATEGY_STETH);
        console.log("Executor 4 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        console.log("All staking operations completed successfully!");
        console.log("Stake weights summary:");
        console.log("- Aggregator: 10 ETH (WETH)");
        console.log("- Executor 1: 20 ETH (stETH)");
        console.log("- Executor 2: 10 ETH (stETH)");
        console.log("- Executor 3: 10 ETH (stETH)");
        console.log("- Executor 4: 10 ETH (stETH)");
    }
}
