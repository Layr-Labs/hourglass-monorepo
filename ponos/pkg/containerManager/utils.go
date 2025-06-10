package containerManager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/docker/go-connections/nat"
)

// HashAvsAddress takes a sha256 hash of the AVS address and returns the first 6 chars
func HashAvsAddress(avsAddress string) string {
	hasher := sha256.New()
	hasher.Write([]byte(avsAddress))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)[0:6]
}

// CreateDefaultContainerConfig creates a default container configuration for AVS performers
func CreateDefaultContainerConfig(avsAddress, imageRepo, imageTag string, containerPort int, networkName string) *ContainerConfig {
	hostname := fmt.Sprintf("avs-performer-%s", HashAvsAddress(avsAddress))

	return &ContainerConfig{
		Hostname: hostname,
		Image:    fmt.Sprintf("%s:%s", imageRepo, imageTag),
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", containerPort)): struct{}{},
		},
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", containerPort)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "", // Let Docker assign a random port
				},
			},
		},
		NetworkName:   networkName,
		AutoRemove:    true,
		RestartPolicy: "no",
		User:          "", // Could be set to non-root user for security
		Privileged:    false,
		ReadOnly:      false,
		MemoryLimit:   0, // No limit by default, could be configurable
		CPUShares:     0, // No limit by default, could be configurable
	}
}

// GetContainerEndpoint returns the connection endpoint for a container
func GetContainerEndpoint(info *ContainerInfo, containerPort int, networkName string) (string, error) {
	containerPortProto := nat.Port(fmt.Sprintf("%d/tcp", containerPort))

	if networkName != "" {
		// When using custom network, use container hostname and container port
		return fmt.Sprintf("%s:%d", info.Hostname, containerPort), nil
	}

	// When using default bridge network, use localhost and mapped port
	if portMap, ok := info.Ports[containerPortProto]; ok && len(portMap) > 0 {
		return fmt.Sprintf("localhost:%s", portMap[0].HostPort), nil
	}

	return "", fmt.Errorf("no port mapping found for container port %d", containerPort)
}
