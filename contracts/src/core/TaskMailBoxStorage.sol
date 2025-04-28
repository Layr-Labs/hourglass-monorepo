// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {ITaskMailbox} from "src/interfaces/ITaskMailbox.sol";

abstract contract TaskMailBoxStorage is ITaskMailbox {
    uint256 internal globalTaskCount;
    mapping(bytes32 taskHash => Task task) internal tasks;

    mapping(bytes32 operatorSetKey => bool registered) public isOperatorSetRegistered;
    mapping(bytes32 operatorSetKey => OperatorSetTaskConfig config) public operatorSetTaskConfig;
}
