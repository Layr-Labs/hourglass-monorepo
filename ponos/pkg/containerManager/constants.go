package containerManager

import "time"

// Default timeouts and intervals
const (
	DefaultStartTimeout       = 30 * time.Second
	DefaultStopTimeout        = 10 * time.Second
	DefaultRestartTimeout     = 60 * time.Second
	DefaultHealthInterval     = 5 * time.Second
	DefaultResourceInterval   = 30 * time.Second
	DefaultRestartDelay       = 2 * time.Second
	DefaultMaxBackoffDelay    = 30 * time.Second
	DefaultBackoffMultiplier  = 2.0
	DefaultHealthTimeout      = 2 * time.Second
	DefaultHealthRetries      = 3
	DefaultHealthStartPeriod  = 10 * time.Second
	DefaultFailureThreshold   = 3
	DefaultMaxRestarts        = 5
	EventChannelTimeout       = 10 * time.Millisecond
	DockerEventReconnectDelay = 5 * time.Second
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
