package containerManager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

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
// If imageDigest is provided (non-empty), it will be used with @ notation (repo@digest)
// Otherwise, imageTag will be used with : notation (repo:tag)
func CreateDefaultContainerConfig(avsAddress, imageRepo, imageTag, imageDigest string, containerPort int, networkName string, envs []string) *ContainerConfig {
	// Use predictable hostname for DNS resolution in Docker networks
	hostname := fmt.Sprintf("avs-performer-%s", HashAvsAddress(avsAddress))

	// Add timestamp to hostname to ensure uniqueness for blue-green deployments
	timestamp := time.Now().Unix()
	uniqueHostname := fmt.Sprintf("%s-%d", hostname, timestamp)

	// Construct the image string based on whether we have a digest or tag
	var imageStr string
	if imageDigest != "" {
		// Use digest notation: repository@sha256:hash
		imageStr = fmt.Sprintf("%s@%s", imageRepo, imageDigest)
	} else {
		// Use tag notation: repository:tag
		imageStr = fmt.Sprintf("%s:%s", imageRepo, imageTag)
	}

	return &ContainerConfig{
		Hostname: uniqueHostname,
		Image:    imageStr,
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
		Env:           envs,
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

// NewDefaultAvsPerformerLivenessConfig creates a default liveness configuration
// optimized for AVS performer containers with aggressive health monitoring
// and auto-restart capabilities
func NewDefaultAvsPerformerLivenessConfig() *LivenessConfig {
	return &LivenessConfig{
		HealthCheckConfig: HealthCheckConfig{
			Enabled:          true,
			Interval:         DefaultHealthInterval,
			Timeout:          DefaultHealthTimeout,
			Retries:          DefaultHealthRetries,
			StartPeriod:      DefaultHealthStartPeriod,
			FailureThreshold: DefaultFailureThreshold,
		},
		RestartPolicy: RestartPolicy{
			Enabled:            DefaultRestartEnabled,
			MaxRestarts:        DefaultMaxRestarts,
			RestartDelay:       DefaultRestartDelay,
			BackoffMultiplier:  DefaultBackoffMultiplier,
			MaxBackoffDelay:    DefaultMaxBackoffDelay,
			RestartTimeout:     DefaultRestartTimeout,
			RestartOnCrash:     DefaultRestartOnCrash,
			RestartOnOOM:       DefaultRestartOnOOM,
			RestartOnUnhealthy: DefaultRestartOnUnhealthy,
		},
		ResourceThresholds: ResourceThresholds{
			CPUThreshold:    DefaultCPUThreshold,
			MemoryThreshold: DefaultMemoryThreshold,
			RestartOnCPU:    DefaultResourceRestartOnCPU,
			RestartOnMemory: DefaultRestartOnMemory,
		},
		MonitorEvents:         DefaultMonitorEvents,
		ResourceMonitoring:    DefaultResourceMonitoring,
		ResourceCheckInterval: DefaultResourceInterval,
	}
}
