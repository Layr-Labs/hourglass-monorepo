package containerManager

import (
	"context"
	"time"
)

// ContainerLifecycleManager handles basic container lifecycle operations
type ContainerLifecycleManager interface {
	Create(ctx context.Context, config *ContainerConfig) (*ContainerInfo, error)
	Start(ctx context.Context, containerID string) error
	Stop(ctx context.Context, containerID string, timeout time.Duration) error
	Remove(ctx context.Context, containerID string, force bool) error
	RestartContainer(ctx context.Context, containerID string, timeout time.Duration) error
}

// ContainerInspector handles container inspection and state queries
type ContainerInspector interface {
	Inspect(ctx context.Context, containerID string) (*ContainerInfo, error)
	IsRunning(ctx context.Context, containerID string) (bool, error)
	GetContainerState(ctx context.Context, containerID string) (*ContainerState, error)
	GetResourceUsage(ctx context.Context, containerID string) (*ResourceUsage, error)
}

// ContainerWaiter handles waiting for container state changes
type ContainerWaiter interface {
	WaitForRunning(ctx context.Context, containerID string, timeout time.Duration) error
}

// NetworkManager handles Docker network operations
type NetworkManager interface {
	CreateNetworkIfNotExists(ctx context.Context, networkName string) error
	RemoveNetwork(ctx context.Context, networkName string) error
}

// HealthMonitor handles container health checking
type HealthMonitor interface {
	StartHealthCheck(ctx context.Context, containerID string, config *HealthCheckConfig) (<-chan bool, error)
	StopHealthCheck(containerID string)
}

// LivenessMonitor handles comprehensive container monitoring and auto-restart
type LivenessMonitor interface {
	StartLivenessMonitoring(ctx context.Context, containerID string, config *LivenessConfig) (<-chan ContainerEvent, error)
	StopLivenessMonitoring(containerID string)
	TriggerRestart(containerID string, reason string) error
	SetRestartPolicy(containerID string, policy RestartPolicy) error
}

// Note: ContainerManager interface is defined in types.go to avoid circular dependencies
// It should be updated to compose these smaller interfaces for better modularity

// Ensure DockerContainerManager implements all interfaces
var (
	_ ContainerLifecycleManager = (*DockerContainerManager)(nil)
	_ ContainerInspector        = (*DockerContainerManager)(nil)
	_ ContainerWaiter           = (*DockerContainerManager)(nil)
	_ NetworkManager            = (*DockerContainerManager)(nil)
	_ HealthMonitor             = (*DockerContainerManager)(nil)
	_ LivenessMonitor           = (*DockerContainerManager)(nil)
	_ ContainerManager          = (*DockerContainerManager)(nil)
)
