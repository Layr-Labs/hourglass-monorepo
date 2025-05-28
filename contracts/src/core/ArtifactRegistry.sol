// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {ArtifactRegistryStorage} from "./ArtifactRegistryStorage.sol";

contract ArtifactRegistry is ArtifactRegistryStorage {
    /// @notice Event emitted when a new artifact is published
    event PublishedArtifact(
        address indexed avs,
        bytes indexed operatorSetId,
        Artifact newArtifact,
        Artifact previousArtifact
    );

    /// @notice Publish a new artifact for an AVS and operator set
    /// @param avs The address of the AVS
    /// @param operatorSetId The ID of the operator set
    /// @param digest The digest of the artifact
    function publishArtifact(
        address avs,
        bytes calldata registryUrl,
        bytes calldata operatorSetId,
        bytes calldata digest
    ) external {
        // For this implementation, we're using a fixed registry URL
        string memory operatorSetIdString = string(operatorSetId);
        
        // Create new artifact
        Artifact memory newArtifact = Artifact({
            digest: digest,
            registryUrl: registryUrl
        });
        
        // Get previous artifact if exists
        Artifact memory previousArtifact;
        if (registries[avs].operatorSetReleases[operatorSetIdString].digests.length > 0) {
            uint256 lastIndex = registries[avs].operatorSetReleases[operatorSetIdString].digests.length - 1;
            bytes memory lastDigest = registries[avs].operatorSetReleases[operatorSetIdString].digests[lastIndex];
            previousArtifact = Artifact({
                digest: lastDigest,
                registryUrl: registries[avs].operatorSetReleases[operatorSetIdString].registryUrl
            });
        }
        
        // If this is the first time seeing this AVS, add it to the list
        if (registries[avs].avsId.length == 0) {
            registries[avs].avsId = abi.encodePacked(avs);
            avsAddresses.push(avs);
        }
        
        // If this is the first time seeing this operator set for this AVS, add it to the list
        bool operatorSetExists = false;
        for (uint256 i = 0; i < registries[avs].operatorSetIds.length; i++) {
            if (keccak256(bytes(registries[avs].operatorSetIds[i])) == keccak256(bytes(operatorSetIdString))) {
                operatorSetExists = true;
                break;
            }
        }
        
        if (!operatorSetExists) {
            registries[avs].operatorSetIds.push(operatorSetIdString);
        }
        
        // Update registry
        registries[avs].operatorSetReleases[operatorSetIdString].registryUrl = registryUrl;
        registries[avs].operatorSetReleases[operatorSetIdString].digests.push(digest);
        
        emit PublishedArtifact(avs, operatorSetId, newArtifact, previousArtifact);
    }

    /// @notice Get the latest artifact for all operator sets of an AVS
    /// @param avs The address of the AVS
    /// @return Array of latest artifact digests
    function getLatestArtifact(address avs) external view returns (bytes[] memory) {
        string[] memory operatorSetIds = registries[avs].operatorSetIds;
        uint256 validSetsCount = 0;
        
        // First count how many operator sets have at least one artifact
        for (uint256 i = 0; i < operatorSetIds.length; i++) {
            if (registries[avs].operatorSetReleases[operatorSetIds[i]].digests.length > 0) {
                validSetsCount++;
            }
        }
        
        bytes[] memory digests = new bytes[](validSetsCount);
        uint256 currentIndex = 0;
        
        // Then collect the latest digest for each valid operator set
        for (uint256 i = 0; i < operatorSetIds.length; i++) {
            bytes[] memory setDigests = registries[avs].operatorSetReleases[operatorSetIds[i]].digests;
            if (setDigests.length > 0) {
                digests[currentIndex] = setDigests[setDigests.length - 1];
                currentIndex++;
            }
        }
        
        return digests;
    }

    /// @notice List all artifacts for an operator
    /// @param operator The address of the operator
    /// @return Array of artifacts
    function listArtifacts(address operator) external view returns (Artifact[] memory) {
        address[] memory avsForOperator = operatorAvs[operator];
        
        // Count total artifacts
        uint256 totalArtifacts = 0;
        for (uint256 i = 0; i < avsForOperator.length; i++) {
            address avs = avsForOperator[i];
            for (uint256 j = 0; j < registries[avs].operatorSetIds.length; j++) {
                string memory operatorSetId = registries[avs].operatorSetIds[j];
                totalArtifacts += registries[avs].operatorSetReleases[operatorSetId].digests.length;
            }
        }
        
        // Create result array
        Artifact[] memory artifacts = new Artifact[](totalArtifacts);
        
        // Fill result array
        uint256 currentIndex = 0;
        for (uint256 i = 0; i < avsForOperator.length; i++) {
            address avs = avsForOperator[i];
            for (uint256 j = 0; j < registries[avs].operatorSetIds.length; j++) {
                string memory operatorSetId = registries[avs].operatorSetIds[j];
                ArtifactVersions storage releases = registries[avs].operatorSetReleases[operatorSetId];
                
                for (uint256 k = 0; k < releases.digests.length; k++) {
                    artifacts[currentIndex] = Artifact({
                        digest: releases.digests[k],
                        registryUrl: releases.registryUrl
                    });
                    currentIndex++;
                }
            }
        }
        
        return artifacts;
    }
    
    /// @notice Associate an operator with an AVS
    /// @param operator The operator address
    /// @param avs The AVS address
    function associateOperatorWithAVS(address operator, address avs) external {
        // Check if AVS exists
        require(registries[avs].avsId.length > 0, "AVS does not exist");
        
        // Check if association already exists
        address[] storage associatedAVS = operatorAvs[operator];
        for (uint256 i = 0; i < associatedAVS.length; i++) {
            if (associatedAVS[i] == avs) {
                return; // Already associated
            }
        }
        
        // Add association
        operatorAvs[operator].push(avs);
    }

    /// @notice Get the latest artifact for a specific operator set of an AVS
    /// @param avs The address of the AVS
    /// @param operatorSetId The ID of the operator set
    /// @return The latest artifact
    function getLatestArtifact(address avs, bytes calldata operatorSetId) external view returns (Artifact memory) {
        string memory operatorSetIdString = string(operatorSetId);
        ArtifactVersions storage releases = registries[avs].operatorSetReleases[operatorSetIdString];
        
        require(releases.digests.length > 0, "No artifacts for this operator set");
        
        uint256 lastIndex = releases.digests.length - 1;
        
        Artifact memory artifact = Artifact({
            digest: releases.digests[lastIndex],
            registryUrl: releases.registryUrl
        });
        
        return artifact;
    }
} 