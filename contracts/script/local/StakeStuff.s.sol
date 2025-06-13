// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {Test} from "forge-std/Test.sol";

import {
    ICrossChainRegistry,
    ICrossChainRegistryTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/ICrossChainRegistry.sol";
import {IBN254TableCalculator} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254TableCalculator.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IStrategy} from "@eigenlayer-contracts/src/contracts/interfaces/IStrategy.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
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
    ICrossChainRegistry public CROSS_CHAIN_REGISTRY = ICrossChainRegistry(0x0022d2014901F2AFBF5610dDFcd26afe2a65Ca6F);
    IBN254TableCalculator public BN254_TABLE_CALCULATOR =
        IBN254TableCalculator(0x033af59c1b030Cc6eEE07B150FD97668497dc74b);

    IAllocationManager public ALLOCATION_MANAGER = IAllocationManager(0xFdD5749e11977D60850E06bF5B13221Ad95eb6B4);
    IDelegationManager public DELEGATION_MANAGER = IDelegationManager(0x75dfE5B44C2E530568001400D3f704bC8AE350CC);
    IStrategyManager public STRATEGY_MANAGER = IStrategyManager(0xF9fbF2e35D8803273E214c99BF15174139f4E67a);

    IStrategy public STRATEGY_WETH = IStrategy(0xD523267698C81a372191136e477fdebFa33D9FB4);
    IStrategy public STRATEGY_STETH = IStrategy(0x5C8b55722f421556a2AAfb7A3EA63d4c3e514312);

    function setUp() public {}

    function run(address aggregatorAddr, address executorAddr) public {
        uint256 aggStakerPrivateKey = vm.envUint("AGG_STAKER_PRIVATE_KEY");
        address aggStakerAddr = vm.addr(aggStakerPrivateKey);

        uint256 execStakerPrivateKey = vm.envUint("EXEC_STAKER_PRIVATE_KEY");
        address execStakerAddr = vm.addr(execStakerPrivateKey);

        // ... operator checks ...

        IERC20 wethToken = STRATEGY_WETH.underlyingToken();
        IERC20 stethToken = STRATEGY_STETH.underlyingToken();

        // For WETH - deposit ETH to get WETH
        vm.deal(aggStakerAddr, 100_000e18);
        vm.startBroadcast(aggStakerPrivateKey);
        IWETH(address(wethToken)).deposit{value: 20e18}();
        // Approve while still broadcasting
        wethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        vm.stopBroadcast();

        // For stETH - submit ETH to get stETH
        vm.deal(execStakerAddr, 100_000e18);
        vm.startBroadcast(execStakerPrivateKey);
        IStETH(address(stethToken)).submit{value: 20e18}(address(0));
        // Approve while still broadcasting
        stethToken.approve(address(STRATEGY_MANAGER), type(uint256).max);
        vm.stopBroadcast();

        // Check balances
        uint256 balance = IERC20(wethToken).balanceOf(aggStakerAddr);
        console.log("WETH balance for aggregator staker:", balance);
        balance = IERC20(stethToken).balanceOf(execStakerAddr);
        console.log("STETH balance for executor staker:", balance);

        uint256 depositAmount = uint256(1 ether);

        // Aggregator staker operations
        console.log("Depositing weth into strategy");
        vm.startBroadcast(aggStakerPrivateKey);
        // No need to approve again since we did it above
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_WETH, wethToken, depositAmount);
        vm.stopBroadcast();

        // Check deposit
        uint256 depositedAmount = STRATEGY_MANAGER.stakerDepositShares(aggStakerAddr, STRATEGY_WETH);
        console.log("Staker deposit shares in STRATEGY_WETH:", depositedAmount);

        // Executor staker operations
        console.log("Depositing steth into strategy");
        vm.startBroadcast(execStakerPrivateKey);
        // No need to approve again since we did it above
        STRATEGY_MANAGER.depositIntoStrategy(STRATEGY_STETH, stethToken, depositAmount);
        vm.stopBroadcast();

        // Check deposit
        depositedAmount = STRATEGY_MANAGER.stakerDepositShares(execStakerAddr, STRATEGY_STETH);
        console.log("Staker deposit shares in STRATEGY_STETH:", depositedAmount);
    }
}
