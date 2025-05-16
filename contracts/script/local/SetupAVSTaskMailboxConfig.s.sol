// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {ITaskMailbox, ITaskMailboxTypes} from "../../src/interfaces/core/ITaskMailbox.sol";
import {IAVSTaskHook} from "../../src/interfaces/avs/l2/IAVSTaskHook.sol";
import {IBN254CertificateVerifier} from "../../src/interfaces/avs/l2/IBN254CertificateVerifier.sol";

contract SetupAVSTaskMailboxConfig is Script {
    function setUp() public {}

    function run(address taskMailbox, address certificateVerifier, address taskHook) public {
        // Load the private key from the environment variable
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");
        address avs = vm.addr(avsPrivateKey);

        vm.startBroadcast(avsPrivateKey);
        console.log("AVS address:", avs);

        // 1. Set the AVS config
        uint32[] memory executorOperatorSetIds = new uint32[](1);
        executorOperatorSetIds[0] = 1;
        ITaskMailboxTypes.AvsConfig memory avsConfig =
            ITaskMailboxTypes.AvsConfig({aggregatorOperatorSetId: 0, executorOperatorSetIds: executorOperatorSetIds});
        ITaskMailbox(taskMailbox).setAvsConfig(avs, avsConfig);
        ITaskMailboxTypes.AvsConfig memory avsConfigStored = ITaskMailbox(taskMailbox).getAvsConfig(avs);
        console.log(
            "AVS config set:", avsConfigStored.aggregatorOperatorSetId, avsConfigStored.executorOperatorSetIds[0]
        );

        // 2. Set the Executor Operator Set Task Config
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory executorOperatorSetTaskConfig = ITaskMailboxTypes
            .ExecutorOperatorSetTaskConfig({
            certificateVerifier: certificateVerifier,
            taskHook: IAVSTaskHook(taskHook),
            feeToken: IERC20(address(0)),
            feeCollector: address(0),
            taskSLA: 60,
            stakeProportionThreshold: 10_000,
            taskMetadata: bytes("")
        });
        ITaskMailbox(taskMailbox).setExecutorOperatorSetTaskConfig(OperatorSet(avs, 1), executorOperatorSetTaskConfig);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory executorOperatorSetTaskConfigStored =
            ITaskMailbox(taskMailbox).getExecutorOperatorSetTaskConfig(OperatorSet(avs, 1));
        console.log(
            "Executor Operator Set Task Config set:",
            executorOperatorSetTaskConfigStored.certificateVerifier,
            address(executorOperatorSetTaskConfigStored.taskHook)
        );

        vm.stopBroadcast();
    }
}
