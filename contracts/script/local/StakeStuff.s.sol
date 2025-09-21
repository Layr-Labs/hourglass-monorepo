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
    ICrossChainRegistry public CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0x287381B1570d9048c4B4C7EC94d21dDb8Aa1352a);

    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0x42583067658071247ec8CE0A516A58f682002d07);
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0xD4A7E1Bd8015057293f0D0A557088c286942e84b);
    IStrategyManager public STRATEGY_MANAGER = IStrategyManager(0x2E3D6c0744b10eb0A4e6F679F71554a39Ec47a5D);

    IStrategy public STRATEGY_WETH = IStrategy(0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc);
    IStrategy public STRATEGY_STETH = IStrategy(0x8b29d91e67b013e855EaFe0ad704aC4Ab086a574);

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

        // Stake for aggregator (WETH) - 20 ETH
        vm.deal(aggStakerAddr, 100_000e18);
        vm.startBroadcast(aggStakerPrivateKey);
        IWETH(address(wethToken)).deposit{value: 20e18}();
        wethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_WETH, wethToken, 1 ether);
        vm.stopBroadcast();

        uint256 balance = IERC20(wethToken).balanceOf(aggStakerAddr);
        console.log("WETH balance for aggregator staker:", balance);
        uint256 depositedAmount = STRATEGY_MANAGER.stakerDepositShares(aggStakerAddr, STRATEGY_WETH);
        console.log("Aggregator staker deposit shares in STRATEGY_WETH:", depositedAmount);

        // Stake for all executors using stETH (since operator set 1 only supports STETH)
        // Executor 1: 2 ETH
        vm.deal(execStakerAddr, 100_000e18);
        vm.startBroadcast(execStakerPrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, 2 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStakerAddr);
        console.log("stETH balance for executor 1 staker:", balance);
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStakerAddr, STRATEGY_STETH);
        console.log("Executor 1 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        // Stake for executor 2 (stETH) - 1.5 ETH
        vm.deal(execStaker2Addr, 100_000e18);
        vm.startBroadcast(execStaker2PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, 1.5 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker2Addr);
        console.log("stETH balance for executor 2 staker:", balance);
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStaker2Addr, STRATEGY_STETH);
        console.log("Executor 2 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        // Stake for executor 3 (stETH) - 1 ETH
        vm.deal(execStaker3Addr, 100_000e18);
        vm.startBroadcast(execStaker3PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, 1 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker3Addr);
        console.log("stETH balance for executor 3 staker:", balance);
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStaker3Addr, STRATEGY_STETH);
        console.log("Executor 3 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        // Stake for executor 4 (stETH) - 0.5 ETH
        vm.deal(execStaker4Addr, 100_000e18);
        vm.startBroadcast(execStaker4PrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, 0.5 ether);
        vm.stopBroadcast();

        balance = IERC20(stethToken).balanceOf(execStaker4Addr);
        console.log("stETH balance for executor 4 staker:", balance);
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStaker4Addr, STRATEGY_STETH);
        console.log("Executor 4 staker deposit shares in STRATEGY_STETH:", depositedAmount);

        console.log("All staking operations completed successfully!");
        console.log("Stake weights summary:");
        console.log("- Aggregator: 1 ETH (WETH)");
        console.log("- Executor 1: 2 ETH (stETH)");
        console.log("- Executor 2: 1.5 ETH (stETH)");
        console.log("- Executor 3: 1 ETH (stETH)");
        console.log("- Executor 4: 0.5 ETH (stETH)");
    }
}