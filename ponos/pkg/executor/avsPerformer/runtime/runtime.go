package runtime

import (
	"context"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"go.uber.org/zap"
)

const containerPort = 8080

// ContainerConfig holds configuration for creating a container
type ContainerConfig struct {
	Name        string
	Image       string
	NetworkName string
	Labels      map[string]string
	Logger      *zap.Logger
}

// ContainerResult holds the result of creating a container
type ContainerResult struct {
	ContainerID     string
	ExposedPort     string
	ContainerHost   string
	PerformerClient performerV1.PerformerServiceClient
}

// HealthCheckConfig holds configuration for health checks
type HealthCheckConfig struct {
	Client      performerV1.PerformerServiceClient
	Identifier  string
	ContainerID string
}

// IContainerRuntimeController defines the interface for container runtime management
type IContainerRuntimeController interface {
	// CreateAndStartContainer creates and starts a container with the given configuration
	CreateAndStartContainer(ctx context.Context, config *ContainerConfig) (*ContainerResult, error)

	// StopAndRemoveContainer stops and removes a container by ID
	StopAndRemoveContainer(containerID string)

	// CreateNetworkIfNotExists creates a network if it doesn't already exist
	CreateNetworkIfNotExists(ctx context.Context, networkName string) error

	// StartHealthCheck starts monitoring health of a performer
	StartHealthCheck(ctx context.Context, config *HealthCheckConfig)

	// CheckHealth performs a single health check
	CheckHealth(ctx context.Context, client performerV1.PerformerServiceClient, identifier string, containerID string)
}
