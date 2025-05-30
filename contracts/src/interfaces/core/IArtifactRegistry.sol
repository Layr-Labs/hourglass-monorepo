// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IArtifactRegistry {

    enum ArtifactType {
        Container,
        Binary
    }

    enum Architecture {
        AMD64,
        ARM64,
        ARM,
        X86,
        RISCV
    }

    enum OperatingSystem {
        Linux,
        Windows,
        MacOS
    }

    enum Distribution {
        Unknown,
        Debian,
        Ubuntu,
        Alpine,
        CentOS,
        Arch
    }

    struct Artifact {
        ArtifactType artifactType;
        Architecture architecture;
        OperatingSystem os;
        Distribution distro;
        bytes digest;
        bytes registryUrl;
    }

    struct ArtifactReleases {
        Artifact[] artifacts;
    }

    event PublishedArtifact(
        address indexed avs,
        bytes indexed operatorSetId,
        Artifact newArtifact,
        Artifact previousArtifact
    );

    event RegisteredAvs(address indexed avs);
    event DeregisteredAvs(address indexed avs);

    function register(address avs) external;

    function deregister(address avs) external;

    function publishArtifact(
        address avs,
        bytes calldata operatorSetId,
        Artifact calldata newArtifact
    ) external;

    function getArtifact(
        address avs,
        bytes calldata operatorSetId
    ) external view returns (Artifact memory);

    function listArtifacts(
        address avs
    ) external view returns (ArtifactReleases[] memory);
}
