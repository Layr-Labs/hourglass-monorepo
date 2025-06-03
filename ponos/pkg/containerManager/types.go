package containerManager

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

// ContainerConfig holds configuration for creating a container
type ContainerConfig struct {
	// Basic container configuration
	Hostname   string
	Image      string
	Env        []string
	WorkingDir string

	// Port configuration
	ExposedPorts nat.PortSet
	PortBindings nat.PortMap

	// Network configuration
	NetworkName string

	// Resource limits
	MemoryLimit int64 // in bytes
	CPUShares   int64

	// Security settings
	User       string
	Privileged bool
	ReadOnly   bool

	// Lifecycle settings
	AutoRemove    bool
	RestartPolicy string
}

// ContainerInfo represents information about a running container
type ContainerInfo struct {
	ID       string
	Hostname string
	Status   string
	Ports    map[nat.Port][]nat.PortBinding
	Networks map[string]*network.EndpointSettings
}

// HealthCheckConfig defines how container health should be monitored
type HealthCheckConfig struct {
	Enabled          bool
	Interval         time.Duration
	Timeout          time.Duration
	Retries          int
	StartPeriod      time.Duration
	FailureThreshold int
}

// ContainerManager defines the interface for managing Docker containers
type ContainerManager interface {
	// Container lifecycle operations
	Create(ctx context.Context, config *ContainerConfig) (*ContainerInfo, error)
	Start(ctx context.Context, containerID string) error
	Stop(ctx context.Context, containerID string, timeout time.Duration) error
	Remove(ctx context.Context, containerID string, force bool) error

	// Container information and monitoring
	Inspect(ctx context.Context, containerID string) (*ContainerInfo, error)
	IsRunning(ctx context.Context, containerID string) (bool, error)
	WaitForRunning(ctx context.Context, containerID string, timeout time.Duration) error

	// Network operations
	CreateNetworkIfNotExists(ctx context.Context, networkName string) error
	RemoveNetwork(ctx context.Context, networkName string) error

	// Health checking
	StartHealthCheck(ctx context.Context, containerID string, config *HealthCheckConfig) (<-chan bool, error)
	StopHealthCheck(containerID string)

	// Cleanup
	Shutdown(ctx context.Context) error
}

// ContainerManagerConfig holds configuration for the container manager
type ContainerManagerConfig struct {
	// Docker client configuration
	DockerHost    string
	DockerVersion string

	// Default timeouts
	DefaultStartTimeout time.Duration
	DefaultStopTimeout  time.Duration

	// Health check defaults
	DefaultHealthCheckConfig *HealthCheckConfig
}