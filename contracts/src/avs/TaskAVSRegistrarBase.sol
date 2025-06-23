// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IKeyRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";

import {ITaskAVSRegistrarBase} from "../interfaces/avs/ITaskAVSRegistrarBase.sol";
import {TaskAVSRegistrarBaseStorage} from "./TaskAVSRegistrarBaseStorage.sol";

/**
 * @title TaskAVSRegistrarBase
 * @author Layr Labs, Inc.
 * @notice Abstract contract that extends AVSRegistrarWithSocket for task-based AVSs
 * @dev This is a minimal wrapper that inherits from TaskAVSRegistrarBaseStorage
 */
abstract contract TaskAVSRegistrarBase is TaskAVSRegistrarBaseStorage {
    /**
     * @notice Constructs the TaskAVSRegistrarBase contract
     * @param _avs The address of the AVS
     * @param _allocationManager The AllocationManager contract address
     * @param _keyRegistrar The KeyRegistrar contract address
     */
    constructor(
        address _avs,
        IAllocationManager _allocationManager,
        IKeyRegistrar _keyRegistrar
    ) TaskAVSRegistrarBaseStorage(_avs, _allocationManager, _keyRegistrar) {}

    /// @inheritdoc ITaskAVSRegistrarBase
    function setAvsConfig(AvsConfig memory config) external {
        // TODO: require checks - Figure out what checks are needed.
        // 1. OperatorSets are valid
        // 2. Only AVS delegated address can set config.

        // Validate executorOperatorSetIds are monotonically increasing (sorted) to efficiently check for duplicates
        if (config.executorOperatorSetIds.length > 0) {
            // Check that first element is not the aggregator
            require(config.aggregatorOperatorSetId != config.executorOperatorSetIds[0], InvalidAggregatorOperatorSetId());
            
            // Check monotonically increasing order and no aggregator overlap in one pass
            for (uint256 i = 1; i < config.executorOperatorSetIds.length; i++) {
                require(config.aggregatorOperatorSetId != config.executorOperatorSetIds[i], InvalidAggregatorOperatorSetId());
                require(config.executorOperatorSetIds[i] > config.executorOperatorSetIds[i - 1], DuplicateExecutorOperatorSetId()); 
            }
        }

        avsConfig = config;
        emit AvsConfigSet(config.aggregatorOperatorSetId, config.executorOperatorSetIds);
    }

    /// @inheritdoc ITaskAVSRegistrarBase
    function getAvsConfig() external view returns (AvsConfig memory) {
        return avsConfig;
    }
}
