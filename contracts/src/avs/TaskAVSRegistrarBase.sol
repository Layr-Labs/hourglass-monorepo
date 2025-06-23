// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IKeyRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";

import {ITaskAVSRegistrarBase} from "../interfaces/avs/ITaskAVSRegistrarBase.sol";
import {TaskAVSRegistrarBaseStorage} from "./TaskAVSRegistrarBaseStorage.sol";

/**
 * @title TaskAVSRegistrarBase
 * @author Layr Labs, Inc.
 * @notice Abstract AVS Registrar for task-based AVSs
 */
abstract contract TaskAVSRegistrarBase is TaskAVSRegistrarBaseStorage, Ownable {
    /**
     * @dev Constructor that passes parameters to parent and sets the owner
     * @param _avs The address of the AVS
     * @param _allocationManager The AllocationManager contract address
     * @param _keyRegistrar The KeyRegistrar contract address
     * @param _owner The owner of the contract
     * @param _initialConfig The initial AVS configuration
     */
    constructor(
        address _avs,
        IAllocationManager _allocationManager,
        IKeyRegistrar _keyRegistrar,
        address _owner,
        AvsConfig memory _initialConfig
    ) TaskAVSRegistrarBaseStorage(_avs, _allocationManager, _keyRegistrar) Ownable() {
        _transferOwnership(_owner);
        _setAvsConfig(_initialConfig);
    }

    /// @inheritdoc ITaskAVSRegistrarBase
    function setAvsConfig(
        AvsConfig memory config
    ) external onlyOwner {
        _setAvsConfig(config);
    }

    /// @inheritdoc ITaskAVSRegistrarBase
    function getAvsConfig() external view returns (AvsConfig memory) {
        return avsConfig;
    }

    /**
     * @notice Internal function to set the AVS configuration
     * @param config The AVS configuration to set
     * @dev The executorOperatorSetIds must be monotonically increasing.
     */
    function _setAvsConfig(
        AvsConfig memory config
    ) internal {
        // Require at least one executor operator set
        require(config.executorOperatorSetIds.length > 0, ExecutorOperatorSetIdsEmpty());

        // Check that first element is not the aggregator
        require(config.aggregatorOperatorSetId != config.executorOperatorSetIds[0], InvalidAggregatorOperatorSetId());

        // Check monotonically increasing order and no aggregator overlap in one pass
        for (uint256 i = 1; i < config.executorOperatorSetIds.length; i++) {
            require(
                config.aggregatorOperatorSetId != config.executorOperatorSetIds[i], InvalidAggregatorOperatorSetId()
            );
            require(
                config.executorOperatorSetIds[i] > config.executorOperatorSetIds[i - 1],
                DuplicateExecutorOperatorSetId()
            );
        }

        avsConfig = config;
        emit AvsConfigSet(config.aggregatorOperatorSetId, config.executorOperatorSetIds);
    }
}
