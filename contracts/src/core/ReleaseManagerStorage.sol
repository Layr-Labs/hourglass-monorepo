// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import "@openzeppelin/contracts/utils/Checkpoints.sol";
import "../interfaces/core/IReleaseManager.sol";

/**
 * @title ReleaseManagerStorage
 * @author Your Organization
 * @notice Storage contract for the ReleaseManager contract.
 */
abstract contract ReleaseManagerStorage is IReleaseManager {
    /// @notice Permission controller for UAM integration
    IPermissionController public permissionController;

    /// @notice Mapping of AVS addresses to registration status
    mapping(address => bool) public registeredAVS;

    /// @notice Mapping of concatenated(avsAddress, digest) to artifact details
    mapping(bytes32 => Artifact) public artifacts;

    /// @notice Mapping of AVS to operatorSetId to array of ALL promoted artifacts (never deleted)
    mapping(address => mapping(bytes32 => PromotedArtifact[])) public allPromotedArtifacts;

    /// @notice Mapping to track if an artifact exists for quick lookup
    mapping(bytes32 => bool) public artifactExists;

    /// @notice Checkpoints tracking the index of the active promotion for each AVS/operatorSet
    /// @dev Uses History to store uint256 indices that point to allPromotedArtifacts array
    mapping(address => mapping(bytes32 => Checkpoints.History)) internal _promotionIndexHistory;

    /// @notice Checkpoints tracking promotion status changes for specific artifacts
    /// @dev Key is keccak256(avs, operatorSetId, digest), value is status enum cast to uint256
    mapping(bytes32 => Checkpoints.History) internal _promotionStatusHistory;

    /// @notice Storage gap for future upgrades
    uint256[41] private __gap;
}