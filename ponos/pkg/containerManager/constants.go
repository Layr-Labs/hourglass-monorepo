package containerManager

import "time"

// Default timeouts and intervals
const (
	DefaultStartTimeout         = 30 * time.Second
	DefaultStopTimeout          = 10 * time.Second
	DefaultRestartTimeout       = 60 * time.Second
	DefaultHealthInterval       = 5 * time.Second
	DefaultResourceInterval     = 30 * time.Second
	DefaultResourceRestartOnCPU = false
	DefaultRestartOnMemory      = false
	DefaultRestartDelay         = 2 * time.Second
	DefaultMaxBackoffDelay      = 30 * time.Second
	DefaultRestartEnabled       = true
	DefaultRestartOnCrash       = true
	DefaultRestartOnOOM         = true
	DefaultRestartOnUnhealthy   = true
	DefaultBackoffMultiplier    = 2.0
	DefaultHealthTimeout        = 2 * time.Second
	DefaultHealthRetries        = 3
	DefaultHealthStartPeriod    = 10 * time.Second
	DefaultHealthEnabled        = true
	DefaultFailureThreshold     = 3
	DefaultMaxRestarts          = 5
	EventChannelTimeout         = 10 * time.Millisecond
	DockerEventReconnectDelay   = 5 * time.Second
	DefaultMonitorEvents        = true
	DefaultResourceMonitoring   = true
)

// Default resource thresholds
const (
	DefaultCPUThreshold    = 90.0
	DefaultMemoryThreshold = 90.0
)

// Event channel buffer size
const (
	EventChannelBufferSize = 10
)

// Docker API constants
const (
	DockerNetworkBridge = "bridge"
	DockerNetworkHost   = "host"
	DockerNetworkNone   = "none"
)
