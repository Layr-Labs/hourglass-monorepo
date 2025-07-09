// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";

import {ITaskMailbox, ITaskMailboxTypes} from "../../src/interfaces/core/ITaskMailbox.sol";
import {IAVSTaskHook} from "../../src/interfaces/avs/l2/IAVSTaskHook.sol";

contract SetupAVSTaskMailboxConfig is Script {
    function setUp() public {}

    function run(address taskMailbox, address taskHook) public {
        // Load the private key from the environment variable
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");
        address avs = vm.addr(avsPrivateKey);

        vm.startBroadcast(avsPrivateKey);
        console.log("AVS address:", avs);

        // Set the Executor Operator Set Task Config
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory executorOperatorSetTaskConfig = ITaskMailboxTypes
            .ExecutorOperatorSetTaskConfig({
            curveType: IKeyRegistrarTypes.CurveType.BN254,
            taskHook: IAVSTaskHook(taskHook),
            feeToken: IERC20(address(0)),
            feeCollector: address(0),
            taskSLA: 60,
            consensus: ITaskMailboxTypes.Consensus({
                consensusType: ITaskMailboxTypes.ConsensusType.STAKE_PROPORTION_THRESHOLD,
                value: abi.encode(uint16(10_000))
            }),
            taskMetadata: bytes("")
        });
        ITaskMailbox(taskMailbox).setExecutorOperatorSetTaskConfig(OperatorSet(avs, 1), executorOperatorSetTaskConfig);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory executorOperatorSetTaskConfigStored =
            ITaskMailbox(taskMailbox).getExecutorOperatorSetTaskConfig(OperatorSet(avs, 1));
        console.log(
            "Executor Operator Set Task Config set with curve type:",
            uint8(executorOperatorSetTaskConfigStored.curveType),
            address(executorOperatorSetTaskConfigStored.taskHook)
        );

        vm.stopBroadcast();
    }
}
