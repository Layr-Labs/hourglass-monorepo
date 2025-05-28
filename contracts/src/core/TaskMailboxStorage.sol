// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {ITaskMailbox} from "../interfaces/core/ITaskMailbox.sol";

/**
 * @title TaskMailboxStorage
 * @author Layr Labs, Inc.
 * @notice Storage contract for the TaskMailbox contract.
 */
abstract contract TaskMailboxStorage is ITaskMailbox {
    /// @notice Global counter for tasks created across the TaskMailbox
    uint256 internal globalTaskCount;

    /// @notice Mapping from task hash to task details
    mapping(bytes32 taskHash => Task task) internal tasks;

    /// @notice Mapping to track registered AVSs
    mapping(address avs => bool isRegistered) public isAvsRegistered;

    /// @notice Mapping from AVS address to its configuration
    mapping(address avs => AvsConfig config) public avsConfigs;

    /// @notice Mapping to track registered executor operator sets by their keys
    mapping(bytes32 operatorSetKey => bool isRegistered) public isExecutorOperatorSetRegistered;

    /// @notice Mapping from executor operator set key to its task configuration
    mapping(bytes32 operatorSetKey => ExecutorOperatorSetTaskConfig config) public executorOperatorSetTaskConfigs;
}
