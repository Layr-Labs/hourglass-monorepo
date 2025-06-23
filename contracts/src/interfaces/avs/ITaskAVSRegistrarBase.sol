// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

/**
 * @title ITaskAVSRegistrarBaseTypes
 * @notice Interface defining the type structures used in the TaskAVSRegistrarBase
 */
interface ITaskAVSRegistrarBaseTypes {
    /**
     * @notice Configuration for an AVS
     * @param aggregatorOperatorSetId The operator set ID responsible for aggregating results
     * @param executorOperatorSetIds Array of operator set IDs responsible for executing tasks
     */
    struct AvsConfig {
        // TODO: Pack storage efficiently.
        uint32 aggregatorOperatorSetId; // TODO: Add avs address too: Any AVS can be an aggregator.
        uint32[] executorOperatorSetIds;
    }
}

/**
 * @title ITaskAVSRegistrarBaseErrors
 * @notice Interface defining errors that can be thrown by the TaskAVSRegistrarBase
 */
interface ITaskAVSRegistrarBaseErrors {
    /// @notice Thrown when an aggregator operator set id is also an executor operator set id
    error InvalidAggregatorOperatorSetId();

    /// @notice Thrown when an executor operator set id is already in the set
    error DuplicateExecutorOperatorSetId();
}

/**
 * @title ITaskAVSRegistrarBaseEvents
 * @notice Interface defining events emitted by the TaskAVSRegistrarBase
 */
interface ITaskAVSRegistrarBaseEvents is ITaskAVSRegistrarBaseTypes {
    /**
     * @notice Emitted when the AVS configuration is set
     * @param caller Address that called the configuration function
     * @param aggregatorOperatorSetId The operator set ID responsible for aggregating results
     * @param executorOperatorSetIds Array of operator set IDs responsible for executing tasks
     */
    event AvsConfigSet(
        address indexed caller, uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds
    );
}

/**
 * @title ITaskAVSRegistrarBase
 * @author Layr Labs, Inc.
 * @notice Interface for TaskAVSRegistrarBase contract that manages AVS configuration
 */
interface ITaskAVSRegistrarBase is ITaskAVSRegistrarBaseErrors, ITaskAVSRegistrarBaseEvents {
    /**
     * @notice Sets the configuration for this AVS
     * @param config Configuration for the AVS
     */
    function setAvsConfig(AvsConfig memory config) external;

    /**
     * @notice Gets the configuration for this AVS
     * @return Configuration for the AVS
     */
    function getAvsConfig() external view returns (AvsConfig memory);
} 