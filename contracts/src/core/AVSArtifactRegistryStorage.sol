// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

abstract contract AVSArtifactRegistryStorage {
    /// @notice Structure for artifact metadata
    struct Artifact {
        bytes digest;
        bytes registryUrl;
    }

    /// @notice Structure for storing artifact versions for a specific operator set
    struct ArtifactVersions {
        bytes registryUrl;
        bytes[] digests;
    }

    /// @notice Structure for registry storage per AVS
    struct RegistryStorage {
        bytes avsId;
        mapping(string => ArtifactVersions) operatorSetReleases;
        string[] operatorSetIds; // Track all operator set IDs for iteration
    }

    /// @notice Mapping of AVS addresses to their registry storage
    mapping(address => RegistryStorage) public registries;
    
    /// @notice Array of all registered AVS addresses
    address[] public avsAddresses;
    
    /// @notice Mapping of operators to the AVSs they're associated with
    mapping(address => address[]) public operatorAvs;
} 