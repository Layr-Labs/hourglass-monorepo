package containerManager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ContainerResult holds the result of a successful container creation and startup
type ContainerResult struct {
	Info     *ContainerInfo
	Endpoint string
}

// HashAvsAddress takes a sha256 hash of the AVS address and returns the first 6 chars
func HashAvsAddress(avsAddress string) string {
	hasher := sha256.New()
	hasher.Write([]byte(avsAddress))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)[0:6]
}

// CreateDefaultContainerConfig creates a default container configuration for AVS performers
func CreateDefaultContainerConfig(avsAddress, imageRepo, imageTag string, containerPort int, networkName string) *ContainerConfig {
	// Use predictable hostname for DNS resolution in Docker networks
	hostname := fmt.Sprintf("avs-performer-%s", HashAvsAddress(avsAddress))

	// Add timestamp to hostname to ensure uniqueness for blue-green deployments
	timestamp := time.Now().Unix()
	uniqueHostname := fmt.Sprintf("%s-%d", hostname, timestamp)

	return &ContainerConfig{
		Hostname: uniqueHostname,
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

// CreateAndStartDefaultContainer is a factory method that creates, starts, and prepares a container
// with default configurations.
// Returns ContainerResult with the container info and endpoint, or error with cleanup
func CreateAndStartDefaultContainer(
	ctx context.Context,
	manager ContainerManager,
	avsAddress string,
	imageRepo string,
	imageTag string,
	containerPort int,
	networkName string,
	logger *zap.Logger,
) (*ContainerResult, error) {
	// Validate required parameters
	if manager == nil {
		return nil, errors.New("container manager cannot be nil")
	}

	// Create container configuration
	containerConfig := CreateDefaultContainerConfig(
		avsAddress,
		imageRepo,
		imageTag,
		containerPort,
		networkName,
	)

	if logger != nil {
		logger.Info("Creating container with default configuration",
			zap.String("avsAddress", avsAddress),
			zap.String("hostname", containerConfig.Hostname),
			zap.String("image", containerConfig.Image),
			zap.String("networkName", containerConfig.NetworkName),
		)
	}

	// Create the container
	containerInfo, err := manager.Create(ctx, containerConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	// Start the container
	if err := manager.Start(ctx, containerInfo.ID); err != nil {
		// Clean up on failure
		if removeErr := manager.Remove(ctx, containerInfo.ID, true); removeErr != nil && logger != nil {
			logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to start container")
	}

	// Wait for the container to be running
	if err := manager.WaitForRunning(ctx, containerInfo.ID, 30*time.Second); err != nil {
		// Clean up on failure
		if removeErr := manager.Remove(ctx, containerInfo.ID, true); removeErr != nil && logger != nil {
			logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to wait for container to be running")
	}

	// Get updated container information with port mappings
	updatedInfo, err := manager.Inspect(ctx, containerInfo.ID)
	if err != nil {
		// Clean up on failure
		if removeErr := manager.Remove(ctx, containerInfo.ID, true); removeErr != nil && logger != nil {
			logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to inspect container")
	}

	// Get the container endpoint
	endpoint, err := GetContainerEndpoint(updatedInfo, containerPort, networkName)
	if err != nil {
		// Clean up on failure
		if removeErr := manager.Remove(ctx, containerInfo.ID, true); removeErr != nil && logger != nil {
			logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to get container endpoint")
	}

	if logger != nil {
		logger.Info("Container created and started successfully",
			zap.String("avsAddress", avsAddress),
			zap.String("containerID", updatedInfo.ID),
			zap.String("endpoint", endpoint),
		)
	}

	return &ContainerResult{
		Info:     updatedInfo,
		Endpoint: endpoint,
	}, nil
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
			Interval:         5 * time.Second,
			Timeout:          2 * time.Second,
			Retries:          3,
			StartPeriod:      10 * time.Second,
			FailureThreshold: 3,
		},
		RestartPolicy: RestartPolicy{
			Enabled:            true,
			MaxRestarts:        5,
			RestartDelay:       2 * time.Second,
			BackoffMultiplier:  2.0,
			MaxBackoffDelay:    30 * time.Second,
			RestartTimeout:     60 * time.Second,
			RestartOnCrash:     true,
			RestartOnOOM:       true,
			RestartOnUnhealthy: true,
		},
		ResourceThresholds: ResourceThresholds{
			CPUThreshold:    90.0,
			MemoryThreshold: 90.0,
			RestartOnCPU:    false, // Log warnings but don't auto-restart on resource thresholds
			RestartOnMemory: false, // Log warnings but don't auto-restart on resource thresholds
		},
		ResourceMonitoring:    true,
		ResourceCheckInterval: 30 * time.Second,
	}
}
