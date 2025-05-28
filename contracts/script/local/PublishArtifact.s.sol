// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";

interface IArtifactRegistry {
    function publishArtifact(
        address avs,
        bytes calldata operatorSetId,
        bytes calldata digest
    ) external;
}

contract PublishArtifact is Script {
    function run(
        address avsAddress,
        address artifactRegistry,
        string memory imageRef
    ) public {
        // Get AVS private key from environment
        uint256 avsPrivateKey = vm.envUint("PRIVATE_KEY_AVS");
        
        // Start broadcasting transactions
        vm.startBroadcast(avsPrivateKey);
        
        console.log("=== Publishing Artifact ===");
        console.log("AVS Address:", avsAddress);
        console.log("Artifact Registry:", artifactRegistry);
        console.log("Operator Set ID: 1");
        console.log("Image Reference:", imageRef);
        
        // Call publishArtifact
        IArtifactRegistry(artifactRegistry).publishArtifact(
            avsAddress,
            bytes("1"), // Operator Set ID as bytes
            bytes(imageRef) // Image reference as bytes
        );
        
        console.log("Artifact published successfully!");
        
        vm.stopBroadcast();
    }
} 