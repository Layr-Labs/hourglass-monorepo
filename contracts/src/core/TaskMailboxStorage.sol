// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {ITaskMailbox} from "../interfaces/core/ITaskMailbox.sol";

abstract contract TaskMailboxStorage is ITaskMailbox {
    uint256 internal globalTaskCount;
    mapping(bytes32 taskHash => Task task) internal tasks;

    mapping(address avs => bool isRegistered) public isAvsRegistered;
    mapping(address avs => AvsConfig config) public avsConfigs;
    mapping(bytes32 operatorSetKey => bool isRegistered) public isExecutorOperatorSetRegistered;
    mapping(bytes32 operatorSetKey => ExecutorOperatorSetTaskConfig config) public executorOperatorSetTaskConfigs;
}
