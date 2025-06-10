package containerManager

import "errors"

// Configuration validation errors
var (
	ErrInvalidStartTimeout     = errors.New("invalid start timeout: must be non-negative")
	ErrInvalidStopTimeout      = errors.New("invalid stop timeout: must be non-negative")
	ErrInvalidHealthInterval   = errors.New("invalid health check interval: must be at least 1 second")
	ErrInvalidFailureThreshold = errors.New("invalid failure threshold: must be at least 1")
	ErrInvalidMaxRestarts      = errors.New("invalid max restarts: must be non-negative")
	ErrInvalidRestartDelay     = errors.New("invalid restart delay: must be non-negative")
	ErrInvalidResourceInterval = errors.New("invalid resource check interval: must be at least 1 second")
)

// Container operation errors
var (
	ErrContainerNotFound      = errors.New("container not found")
	ErrContainerNotRunning    = errors.New("container is not running")
	ErrContainerCreateFailed  = errors.New("failed to create container")
	ErrContainerStartFailed   = errors.New("failed to start container")
	ErrContainerStopFailed    = errors.New("failed to stop container")
	ErrContainerRestartFailed = errors.New("failed to restart container")
	ErrImagePullFailed        = errors.New("failed to pull image")
	ErrNetworkCreateFailed    = errors.New("failed to create network")
)

// Monitoring errors
var (
	ErrMonitorNotFound     = errors.New("monitor not found for container")
	ErrMonitoringDisabled  = errors.New("monitoring is disabled")
	ErrRestartPolicyFailed = errors.New("restart policy violation")
	ErrHealthCheckFailed   = errors.New("health check failed")
)

// Utility error types for better error handling
type ContainerError struct {
	ContainerID string
	Operation   string
	Err         error
}

func (e *ContainerError) Error() string {
	return e.Err.Error()
}

func (e *ContainerError) Unwrap() error {
	return e.Err
}

// NewContainerError creates a new container-specific error
func NewContainerError(containerID, operation string, err error) *ContainerError {
	return &ContainerError{
		ContainerID: containerID,
		Operation:   operation,
		Err:         err,
	}
}
