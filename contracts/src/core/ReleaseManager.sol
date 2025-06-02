// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {OwnableUpgradeable} from "@openzeppelin-upgradeable/contracts/access/OwnableUpgradeable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import {Checkpoints} from "@openzeppelin/contracts/utils/Checkpoints.sol";

import {IReleaseManager, PromotionStatus} from "../interfaces/core/IReleaseManager.sol";
import {IPermissionController} from "@eigenlayer-contracts/src/contracts/interfaces/IPermissionController.sol";
import {ReleaseManagerStorage} from "./ReleaseManagerStorage.sol";

/**
 * @title ReleaseManager
 * @author EigenLabs
 * @notice Contract for managing the lifecycle of artifacts and releases for AVS deployments.
 */
contract ReleaseManager is ReentrancyGuard, OwnableUpgradeable, ReleaseManagerStorage {
    using Checkpoints for Checkpoints.History;

    /**
     *
     *                         MODIFIERS
     *
     */

    modifier onlyRegistered(address avs) {
        if (!registeredAVS[avs]) revert AVSNotRegistered();
        _;
    }

    modifier onlyAuthorized(address avs) {
        if (!_isAuthorized(avs, msg.sender)) revert Unauthorized();
        _;
    }

    /**
     *
     *                         INITIALIZATION
     *
     */

    function initialize(address _permissionController) public initializer {
        __Ownable_init();
        permissionController = IPermissionController(_permissionController);
    }

    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */

    /// @inheritdoc IReleaseManager
    function register(address avs) external onlyAuthorized(avs) {
        if (registeredAVS[avs]) revert AVSAlreadyRegistered();
        registeredAVS[avs] = true;
        emit AVSRegistered(avs);
    }

    /// @inheritdoc IReleaseManager
    function deregister(address avs) external onlyAuthorized(avs) {
        if (!registeredAVS[avs]) revert AVSNotRegistered();
        _registerAvs(avs, false);
    }

    /// @inheritdoc IReleaseManager
    function publishArtifacts(
        address avs,
        Artifact[] calldata _artifacts
    ) external nonReentrant onlyRegistered(avs) onlyAuthorized(avs) {
        for (uint256 i = 0; i < _artifacts.length; i++) {
            Artifact memory artifact = _artifacts[i];
            artifact.publishedAt = block.timestamp;

            // Create unique key: keccak256(abi.encodePacked(avs, digest))
            bytes32 key = _getArtifactKey(avs, artifact.digest);

            // Store artifact
            artifacts[key] = artifact;
            artifactExists[key] = true;

            emit ArtifactPublished(
                avs,
                artifact.digest,
                artifact.registryUrl,
                artifact.architecture,
                artifact.os,
                artifact.artifactType
            );
        }
    }

    /// @inheritdoc IReleaseManager
    function promoteArtifacts(
        address avs,
        ArtifactPromotion[] calldata promotions,
        string calldata version,
        uint256 deploymentDeadline
    ) external nonReentrant onlyRegistered(avs) onlyAuthorized(avs) {
        if (deploymentDeadline <= block.timestamp) revert InvalidDeadline();

        bytes32[] memory digests = new bytes32[](promotions.length);
        PromotionStatus[] memory statuses = new PromotionStatus[](promotions.length);

        for (uint256 i = 0; i < promotions.length; i++) {
            ArtifactPromotion memory promotion = promotions[i];

            // Verify artifact exists
            bytes32 artifactKey = _getArtifactKey(avs, promotion.digest);
            if (!artifactExists[artifactKey]) revert ArtifactNotFound();

            digests[i] = promotion.digest;
            statuses[i] = promotion.promotionStatus;

            // Add to each specified operator set
            for (uint256 j = 0; j < promotion.operatorSetIds.length; j++) {
                bytes32 operatorSetId = promotion.operatorSetIds[j];

                PromotedArtifact memory promoted = PromotedArtifact({
                    digest: promotion.digest,
                    registryUrl: promotion.registryUrl,
                    status: promotion.promotionStatus,
                    version: version,
                    deploymentDeadline: deploymentDeadline,
                    promotedAt: block.timestamp
                });

                // Add to the permanent history array
                allPromotedArtifacts[avs][operatorSetId].push(promoted);

                // Update the checkpoint to point to the new index
                uint256 newIndex = allPromotedArtifacts[avs][operatorSetId].length - 1;
                Checkpoints.push(_promotionIndexHistory[avs][operatorSetId], newIndex);

                // Also track the initial status in status history
                bytes32 statusKey = keccak256(abi.encodePacked(avs, operatorSetId, promotion.digest));
                Checkpoints.push(_promotionStatusHistory[statusKey], uint256(promotion.promotionStatus));
            }
        }

        emit ArtifactsPromoted(avs, version, deploymentDeadline, digests, statuses);
    }

    /// @inheritdoc IReleaseManager
    function updatePromotionStatus(
        address avs,
        bytes32 digest,
        bytes32 operatorSetId,
        PromotionStatus newStatus
    ) external nonReentrant onlyRegistered(avs) onlyAuthorized(avs) {
        // Find the current active promotion for this operator set
        uint256 currentIndex = Checkpoints.latest(_promotionIndexHistory[avs][operatorSetId]);
        PromotedArtifact[] storage promoted = allPromotedArtifacts[avs][operatorSetId];

        if (promoted.length == 0) revert ArtifactNotFound();

        // Check if the digest matches the current promotion
        bool found = false;
        if (promoted[currentIndex].digest == digest) {
            // Update the status in the current promotion
            promoted[currentIndex].status = newStatus;
            found = true;

            // Record the status change in history
            bytes32 statusKey = keccak256(abi.encodePacked(avs, operatorSetId, digest));
            uint256 oldStatusValue = Checkpoints.latest(_promotionStatusHistory[statusKey]);
            PromotionStatus oldStatus = PromotionStatus(oldStatusValue);

            Checkpoints.push(_promotionStatusHistory[statusKey], uint256(newStatus));

            emit PromotionStatusUpdated(avs, digest, operatorSetId, oldStatus, newStatus);
        }

        if (!found) revert ArtifactNotFound();
    }

    /**
     *
     *                         INTERNAL FUNCTIONS
     *
     */

    /**
     * @notice Registers or deregisters an AVS
     * @param avs The AVS address to register or deregister
     * @param isRegistered Whether to register (true) or deregister (false) the AVS
     */
    function _registerAvs(address avs, bool isRegistered) internal {
        registeredAVS[avs] = isRegistered;
        if (isRegistered) {
            emit AVSRegistered(avs);
        } else {
            emit AVSDeregistered(avs);
        }
    }

    /**
     * @notice Check if caller is authorized to act on behalf of AVS
     * @param avs The AVS address
     * @param caller The caller address
     * @return True if authorized
     */
    function _isAuthorized(address avs, address caller) internal view returns (bool) {
        // Check if caller is the AVS itself
        if (avs == caller) return true;

        // Check if caller is an admin via permission controller
        if (address(permissionController) != address(0)) {
            return permissionController.isAdmin(avs, caller);
        }

        return false;
    }

    /**
     * @notice Generate unique key for artifact storage
     * @param avs The AVS address
     * @param digest The artifact digest
     * @return The unique key
     */
    function _getArtifactKey(address avs, bytes32 digest) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(avs, digest));
    }

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */

    /// @inheritdoc IReleaseManager
    function getArtifact(address avs, bytes32 digest) external view returns (Artifact memory) {
        bytes32 key = _getArtifactKey(avs, digest);
        if (!artifactExists[key]) revert ArtifactNotFound();
        return artifacts[key];
    }

    /// @inheritdoc IReleaseManager
    function getPromotedArtifacts(
        address avs,
        bytes32 operatorSetId
    ) external view returns (PromotedArtifact[] memory) {
        // Return only the currently active promotion based on the latest checkpoint
        uint256 currentIndex = Checkpoints.latest(_promotionIndexHistory[avs][operatorSetId]);
        PromotedArtifact[] memory promoted = allPromotedArtifacts[avs][operatorSetId];

        if (promoted.length == 0) {
            return new PromotedArtifact[](0);
        }

        // Return single element array with current promotion
        PromotedArtifact[] memory current = new PromotedArtifact[](1);
        current[0] = promoted[currentIndex];
        return current;
    }

    /// @inheritdoc IReleaseManager
    function getLatestPromotedArtifact(
        address avs,
        bytes32 operatorSetId
    ) external view returns (PromotedArtifact memory) {
        uint256 currentIndex = Checkpoints.latest(_promotionIndexHistory[avs][operatorSetId]);
        PromotedArtifact[] memory promoted = allPromotedArtifacts[avs][operatorSetId];
        if (promoted.length == 0) revert ArtifactNotFound();
        return promoted[currentIndex];
    }

    /**
     * @notice Get the promoted artifact that was active at a specific block
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @param blockNumber The block number to query
     * @return The promoted artifact that was active at that block
     */
    function getPromotedArtifactAtBlock(
        address avs,
        bytes32 operatorSetId,
        uint256 blockNumber
    ) external view returns (PromotedArtifact memory) {
        uint256 index = Checkpoints.getAtBlock(_promotionIndexHistory[avs][operatorSetId], blockNumber);
        PromotedArtifact[] memory promoted = allPromotedArtifacts[avs][operatorSetId];
        if (promoted.length == 0 || index >= promoted.length) revert ArtifactNotFound();
        return promoted[index];
    }

    /**
     * @notice Get the complete promotion history for an AVS and operator set
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @return The complete array of all historical promotions
     */
    function getPromotionHistory(
        address avs,
        bytes32 operatorSetId
    ) external view returns (PromotedArtifact[] memory) {
        return allPromotedArtifacts[avs][operatorSetId];
    }

    /**
     * @notice Get the number of checkpoints in the promotion history
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @return The number of promotion checkpoints
     */
    function getPromotionCheckpointCount(
        address avs,
        bytes32 operatorSetId
    ) external view returns (uint256) {
        return Checkpoints.length(_promotionIndexHistory[avs][operatorSetId]);
    }

    /**
     * @notice Get a specific checkpoint from the promotion history
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @param pos The position in the checkpoint array
     * @return blockNumber The block number of the checkpoint
     * @return artifactIndex The index of the promoted artifact
     */
    function getPromotionCheckpointAt(
        address avs,
        bytes32 operatorSetId,
        uint256 pos
    ) external view returns (uint256 blockNumber, uint256 artifactIndex) {
        require(pos < Checkpoints.length(_promotionIndexHistory[avs][operatorSetId]), "Invalid position");
        (bool exists, uint32 _blockNumber, uint224 _value) = Checkpoints.latestCheckpoint(_promotionIndexHistory[avs][operatorSetId]);
        require(exists, "No checkpoints");
        return (uint256(_blockNumber), uint256(_value));
    }

    /**
     * @notice Get the status history for a specific artifact in an operator set
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @param digest The artifact digest
     * @param blockNumber The block number to query
     * @return The promotion status at that block
     */
    function getPromotionStatusAtBlock(
        address avs,
        bytes32 operatorSetId,
        bytes32 digest,
        uint256 blockNumber
    ) external view returns (PromotionStatus) {
        bytes32 statusKey = keccak256(abi.encodePacked(avs, operatorSetId, digest));
        uint256 statusValue = Checkpoints.getAtBlock(_promotionStatusHistory[statusKey], blockNumber);
        return PromotionStatus(statusValue);
    }
}