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
	Enabled          bool // false = disabled (default), true = enabled (must be explicitly set)
	Interval         time.Duration
	Timeout          time.Duration
	Retries          int
	StartPeriod      time.Duration
	FailureThreshold int
}

// ContainerState represents detailed container state information
type ContainerState struct {
	Status       string // running, stopped, crashed, oom-killed, etc.
	ExitCode     int    // Non-zero indicates crash
	StartedAt    time.Time
	FinishedAt   *time.Time
	RestartCount int
	OOMKilled    bool
	Error        string
	Restarting   bool
}

// ContainerEventType represents the type of container lifecycle event
type ContainerEventType string

const (
	EventStarted       ContainerEventType = "started"
	EventStopped       ContainerEventType = "stopped"
	EventCrashed       ContainerEventType = "crashed"
	EventOOMKilled     ContainerEventType = "oom-killed"
	EventRestarted     ContainerEventType = "restarted"
	EventHealthy       ContainerEventType = "healthy"
	EventUnhealthy     ContainerEventType = "unhealthy"
	EventRestarting    ContainerEventType = "restarting"
	EventRestartFailed ContainerEventType = "restart-failed"
)

// ContainerEvent represents container lifecycle events
type ContainerEvent struct {
	ContainerID string
	Type        ContainerEventType
	State       ContainerState
	Timestamp   time.Time
	Message     string
}

// RestartPolicy defines container restart behavior
type RestartPolicy struct {
	Enabled            bool
	MaxRestarts        int           // -1 for unlimited
	RestartDelay       time.Duration // Initial delay between restarts
	BackoffMultiplier  float64       // Exponential backoff multiplier
	MaxBackoffDelay    time.Duration // Maximum backoff delay
	RestartTimeout     time.Duration // Max time to wait for restart
	RestartOnCrash     bool          // Restart on non-zero exit codes
	RestartOnOOM       bool          // Restart on OOM kills
	RestartOnUnhealthy bool          // Restart on health check failures
}

// ResourceUsage represents container resource utilization
type ResourceUsage struct {
	CPUPercent    float64 // CPU usage percentage
	MemoryUsage   int64   // Memory usage in bytes
	MemoryLimit   int64   // Memory limit in bytes
	MemoryPercent float64 // Memory usage percentage
	NetworkRx     int64   // Network bytes received
	NetworkTx     int64   // Network bytes transmitted
	DiskRead      int64   // Disk bytes read
	DiskWrite     int64   // Disk bytes written
	Timestamp     time.Time
}

// ResourceThresholds defines thresholds for proactive container management
type ResourceThresholds struct {
	CPUThreshold    float64 // CPU percentage threshold for alerts
	MemoryThreshold float64 // Memory percentage threshold for alerts
	RestartOnCPU    bool    // Restart container if CPU threshold exceeded
	RestartOnMemory bool    // Restart container if memory threshold exceeded
}

// LivenessConfig extends HealthCheckConfig with restart capabilities and monitoring
type LivenessConfig struct {
	HealthCheckConfig
	RestartPolicy         RestartPolicy
	ResourceThresholds    ResourceThresholds
	MonitorEvents         bool          // Deprecated: Docker event monitoring adds no value over health polling
	ResourceMonitoring    bool          // Monitor CPU/memory usage
	ResourceCheckInterval time.Duration // How often to check resource usage
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

	// Basic health checking (legacy)
	StartHealthCheck(ctx context.Context, containerID string, config *HealthCheckConfig) (<-chan bool, error)
	StopHealthCheck(containerID string)

	// Enhanced liveness monitoring with auto-restart
	StartLivenessMonitoring(ctx context.Context, containerID string, config *LivenessConfig) (<-chan ContainerEvent, error)
	StopLivenessMonitoring(containerID string)
	GetContainerState(ctx context.Context, containerID string) (*ContainerState, error)

	// Restart capabilities
	RestartContainer(ctx context.Context, containerID string, timeout time.Duration) error
	SetRestartPolicy(containerID string, policy RestartPolicy) error

	// Resource monitoring
	GetResourceUsage(ctx context.Context, containerID string) (*ResourceUsage, error)

	// Manual restart trigger (for serverPerformer to call)
	TriggerRestart(containerID string, reason string) error

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

	// Health check defaults (legacy)
	DefaultHealthCheckConfig *HealthCheckConfig

	// Liveness monitoring defaults
	DefaultLivenessConfig *LivenessConfig
}
