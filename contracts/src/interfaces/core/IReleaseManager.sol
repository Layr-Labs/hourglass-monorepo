// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import "@eigenlayer-contracts/src/contracts/interfaces/IPermissionController.sol";

// ============ Types ============
enum Architecture {
    AMD64,
    ARM64
}

enum OperatingSystem {
    Linux,
    Darwin,
    Windows
}

enum ArtifactType {
    Binary,
    Container
}

enum PromotionStatus {
    Outdated,
    Deprecated,
    Preferred,
    Required
}

interface IReleaseManagerErrors {
    /// @notice Thrown when an AVS is not registered
    error AVSNotRegistered();
    /// @notice Thrown when an AVS is already registered
    error AVSAlreadyRegistered();
    /// @notice Thrown when an artifact digest is not found
    error ArtifactNotFound();
    /// @notice Thrown when trying to promote with an invalid deadline
    error InvalidDeadline();
    /// @notice Thrown when arrays have mismatched lengths
    error ArrayLengthMismatch();
    /// @notice Thrown when caller lacks required permissions
    error Unauthorized();
}

interface IReleaseManagerEvents {
    /// @notice Emitted when an artifact is published
    event ArtifactPublished(
        address indexed avs,
        bytes32 indexed digest,
        string registryUrl,
        Architecture architecture,
        OperatingSystem os,
        ArtifactType artifactType
    );

    /// @notice Emitted when artifacts are promoted
    event ArtifactsPromoted(
        address indexed avs,
        string indexed version,
        uint256 deploymentDeadline,
        bytes32[] digests,
        PromotionStatus[] statuses
    );

    /// @notice Emitted when a promotion status is updated
    event PromotionStatusUpdated(
        address indexed avs,
        bytes32 indexed digest,
        bytes32 indexed operatorSetId,
        PromotionStatus oldStatus,
        PromotionStatus newStatus
    );

    /// @notice Emitted when an AVS registers
    event AVSRegistered(address indexed avs);

    /// @notice Emitted when an AVS deregisters
    event AVSDeregistered(address indexed avs);
}

interface IReleaseManager is IReleaseManagerErrors, IReleaseManagerEvents {
    struct Artifact {
        ArtifactType artifactType;
        Architecture architecture;
        OperatingSystem os;
        bytes32 digest;
        string registryUrl;
        uint256 publishedAt;
    }

    struct ArtifactPromotion {
        PromotionStatus promotionStatus;
        bytes32 digest;
        string registryUrl;
        bytes32[] operatorSetIds;
    }

    struct PromotedArtifact {
        bytes32 digest;
        string registryUrl;
        PromotionStatus status;
        string version;
        uint256 deploymentDeadline;
        uint256 promotedAt;
    }

    // ============ Functions ============

    /**
     * @notice Register an AVS to use the ReleaseManager
     * @param avs The address of the AVS to register
     */
    function register(address avs) external;

    /**
     * @notice Deregister an AVS from the ReleaseManager
     * @param avs The address of the AVS to deregister
     */
    function deregister(address avs) external;

    /**
     * @notice Publish artifacts for an AVS
     * @param avs The address of the AVS publishing artifacts
     * @param artifacts Array of artifacts to publish
     */
    function publishArtifacts(address avs, Artifact[] calldata artifacts) external;

    /**
     * @notice Promote artifacts to specific operator sets with a deployment deadline
     * @param avs The address of the AVS promoting artifacts
     * @param promotions Array of artifact promotions
     * @param version Semantic version string
     * @param deploymentDeadline UTC timestamp deadline for deployment
     */
    function promoteArtifacts(
        address avs,
        ArtifactPromotion[] calldata promotions,
        string calldata version,
        uint256 deploymentDeadline
    ) external;

    /**
     * @notice Update the promotion status of a specific artifact
     * @param avs The address of the AVS
     * @param digest The digest of the artifact
     * @param operatorSetId The operator set ID
     * @param newStatus The new promotion status
     */
    function updatePromotionStatus(
        address avs,
        bytes32 digest,
        bytes32 operatorSetId,
        PromotionStatus newStatus
    ) external;

    /**
     * @notice Get artifact details by AVS and digest
     * @param avs The address of the AVS
     * @param digest The digest of the artifact
     * @return The artifact details
     */
    function getArtifact(address avs, bytes32 digest) external view returns (Artifact memory);

    /**
     * @notice Get all promoted artifacts for an AVS and operator set
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @return Array of promoted artifacts
     */
    function getPromotedArtifacts(
        address avs,
        bytes32 operatorSetId
    ) external view returns (PromotedArtifact[] memory);

    /**
     * @notice Get the latest promoted artifact for an AVS and operator set
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @return The latest promoted artifact
     */
    function getLatestPromotedArtifact(
        address avs,
        bytes32 operatorSetId
    ) external view returns (PromotedArtifact memory);

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
    ) external view returns (PromotedArtifact memory);

    /**
     * @notice Get the complete promotion history for an AVS and operator set
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @return The complete array of all historical promotions
     */
    function getPromotionHistory(
        address avs,
        bytes32 operatorSetId
    ) external view returns (PromotedArtifact[] memory);

    /**
     * @notice Get the number of checkpoints in the promotion history
     * @param avs The address of the AVS
     * @param operatorSetId The operator set ID
     * @return The number of promotion checkpoints
     */
    function getPromotionCheckpointCount(
        address avs,
        bytes32 operatorSetId
    ) external view returns (uint256);

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
    ) external view returns (uint256 blockNumber, uint256 artifactIndex);

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
    ) external view returns (PromotionStatus);
}