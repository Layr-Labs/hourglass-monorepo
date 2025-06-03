package containerManager

import (
	"context"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// DockerContainerManager implements ContainerManager using Docker
type DockerContainerManager struct {
	client       *client.Client
	config       *ContainerManagerConfig
	logger       *zap.Logger
	healthChecks map[string]context.CancelFunc
	mu           sync.RWMutex
}

// NewDockerContainerManager creates a new Docker-based container manager
func NewDockerContainerManager(config *ContainerManagerConfig, logger *zap.Logger) (*DockerContainerManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Docker client")
	}

	// Set default values if not provided
	if config == nil {
		config = &ContainerManagerConfig{}
	}
	if config.DefaultStartTimeout == 0 {
		config.DefaultStartTimeout = 30 * time.Second
	}
	if config.DefaultStopTimeout == 0 {
		config.DefaultStopTimeout = 10 * time.Second
	}
	if config.DefaultHealthCheckConfig == nil {
		config.DefaultHealthCheckConfig = &HealthCheckConfig{
			Enabled:          true,
			Interval:         5 * time.Second,
			Timeout:          2 * time.Second,
			Retries:          3,
			StartPeriod:      10 * time.Second,
			FailureThreshold: 3,
		}
	}

	return &DockerContainerManager{
		client:       dockerClient,
		config:       config,
		logger:       logger,
		healthChecks: make(map[string]context.CancelFunc),
	}, nil
}

// Create creates a new container with the given configuration
func (dcm *DockerContainerManager) Create(ctx context.Context, config *ContainerConfig) (*ContainerInfo, error) {
	dcm.logger.Debug("Creating container", zap.String("hostname", config.Hostname), zap.String("image", config.Image))

	// Negotiate API version
	dcm.client.NegotiateAPIVersion(ctx)

	// Create network if specified
	if config.NetworkName != "" {
		if err := dcm.CreateNetworkIfNotExists(ctx, config.NetworkName); err != nil {
			return nil, errors.Wrap(err, "failed to create network")
		}
	}

	// Build container configuration
	containerConfig := &container.Config{
		Hostname:     config.Hostname,
		Image:        config.Image,
		Env:          config.Env,
		WorkingDir:   config.WorkingDir,
		ExposedPorts: config.ExposedPorts,
		User:         config.User,
	}

	// Build host configuration
	hostConfig := &container.HostConfig{
		AutoRemove:   config.AutoRemove,
		PortBindings: config.PortBindings,
		Privileged:   config.Privileged,
		ReadonlyRootfs: config.ReadOnly,
	}

	// Set resource limits if specified
	if config.MemoryLimit > 0 {
		hostConfig.Memory = config.MemoryLimit
	}
	if config.CPUShares > 0 {
		hostConfig.CPUShares = config.CPUShares
	}

	// Set restart policy if specified
	if config.RestartPolicy != "" {
		hostConfig.RestartPolicy = container.RestartPolicy{
			Name: container.RestartPolicyMode(config.RestartPolicy),
		}
	}

	// Build network configuration
	var netConfig *network.NetworkingConfig
	if config.NetworkName != "" {
		netConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				config.NetworkName: {},
			},
		}
	}

	// Create the container
	resp, err := dcm.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		netConfig,
		nil,
		config.Hostname,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	dcm.logger.Info("Container created successfully",
		zap.String("containerID", resp.ID),
		zap.String("hostname", config.Hostname),
	)

	// Return container info
	return &ContainerInfo{
		ID:       resp.ID,
		Hostname: config.Hostname,
		Status:   "created",
	}, nil
}

// Start starts a container
func (dcm *DockerContainerManager) Start(ctx context.Context, containerID string) error {
	dcm.logger.Debug("Starting container", zap.String("containerID", containerID))

	if err := dcm.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return errors.Wrap(err, "failed to start container")
	}

	dcm.logger.Info("Container started successfully", zap.String("containerID", containerID))
	return nil
}

// Stop stops a container
func (dcm *DockerContainerManager) Stop(ctx context.Context, containerID string, timeout time.Duration) error {
	dcm.logger.Debug("Stopping container", zap.String("containerID", containerID))

	// Stop any health checks first
	dcm.StopHealthCheck(containerID)

	if timeout == 0 {
		timeout = dcm.config.DefaultStopTimeout
	}

	timeoutSeconds := int(timeout.Seconds())
	stopOptions := container.StopOptions{
		Timeout: &timeoutSeconds,
	}

	if err := dcm.client.ContainerStop(ctx, containerID, stopOptions); err != nil {
		return errors.Wrap(err, "failed to stop container")
	}

	dcm.logger.Info("Container stopped successfully", zap.String("containerID", containerID))
	return nil
}

// Remove removes a container
func (dcm *DockerContainerManager) Remove(ctx context.Context, containerID string, force bool) error {
	dcm.logger.Debug("Removing container", zap.String("containerID", containerID))

	removeOptions := container.RemoveOptions{
		Force: force,
	}

	if err := dcm.client.ContainerRemove(ctx, containerID, removeOptions); err != nil {
		return errors.Wrap(err, "failed to remove container")
	}

	dcm.logger.Info("Container removed successfully", zap.String("containerID", containerID))
	return nil
}

// Inspect returns information about a container
func (dcm *DockerContainerManager) Inspect(ctx context.Context, containerID string) (*ContainerInfo, error) {
	containerJSON, err := dcm.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to inspect container")
	}

	return &ContainerInfo{
		ID:       containerJSON.ID,
		Hostname: containerJSON.Config.Hostname,
		Status:   containerJSON.State.Status,
		Ports:    containerJSON.NetworkSettings.Ports,
		Networks: containerJSON.NetworkSettings.Networks,
	}, nil
}

// IsRunning checks if a container is running
func (dcm *DockerContainerManager) IsRunning(ctx context.Context, containerID string) (bool, error) {
	info, err := dcm.Inspect(ctx, containerID)
	if err != nil {
		return false, err
	}

	return info.Status == "running", nil
}

// WaitForRunning waits for a container to be running with ports exposed
func (dcm *DockerContainerManager) WaitForRunning(ctx context.Context, containerID string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = dcm.config.DefaultStartTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("timeout waiting for container to be running")
		case <-ticker.C:
			running, err := dcm.IsRunning(ctx, containerID)
			if err != nil {
				return errors.Wrap(err, "failed to check container status")
			}

			if running {
				// Additional check to ensure ports are exposed
				info, err := dcm.Inspect(ctx, containerID)
				if err != nil {
					return errors.Wrap(err, "failed to inspect container")
				}

				if len(info.Ports) > 0 {
					dcm.logger.Info("Container is running with ports exposed", zap.String("containerID", containerID))
					return nil
				}
			}
		}
	}
}

// CreateNetworkIfNotExists creates a Docker network if it doesn't already exist
func (dcm *DockerContainerManager) CreateNetworkIfNotExists(ctx context.Context, networkName string) error {
	networks, err := dcm.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list networks")
	}

	// Check if network already exists
	for _, net := range networks {
		if net.Name == networkName {
			dcm.logger.Debug("Network already exists", zap.String("networkName", networkName))
			return nil
		}
	}

	// Create the network
	_, err = dcm.client.NetworkCreate(
		ctx,
		networkName,
		network.CreateOptions{
			Driver: "bridge",
			Options: map[string]string{
				"com.docker.net.bridge.enable_icc": "true",
			},
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to create network")
	}

	dcm.logger.Info("Network created successfully", zap.String("networkName", networkName))
	return nil
}

// RemoveNetwork removes a Docker network
func (dcm *DockerContainerManager) RemoveNetwork(ctx context.Context, networkName string) error {
	if err := dcm.client.NetworkRemove(ctx, networkName); err != nil {
		return errors.Wrap(err, "failed to remove network")
	}

	dcm.logger.Info("Network removed successfully", zap.String("networkName", networkName))
	return nil
}

// StartHealthCheck starts a health check routine for a container
func (dcm *DockerContainerManager) StartHealthCheck(ctx context.Context, containerID string, config *HealthCheckConfig) (<-chan bool, error) {
	if config == nil {
		config = dcm.config.DefaultHealthCheckConfig
	}

	if !config.Enabled {
		return nil, nil
	}

	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	// Stop existing health check if any
	if cancelFunc, exists := dcm.healthChecks[containerID]; exists {
		cancelFunc()
	}

	healthCtx, cancel := context.WithCancel(ctx)
	dcm.healthChecks[containerID] = cancel

	healthChan := make(chan bool, 1)

	go func() {
		defer close(healthChan)
		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		failures := 0

		for {
			select {
			case <-healthCtx.Done():
				return
			case <-ticker.C:
				running, err := dcm.IsRunning(healthCtx, containerID)
				if err != nil {
					dcm.logger.Error("Health check failed", 
						zap.String("containerID", containerID), 
						zap.Error(err),
					)
					failures++
				} else if !running {
					dcm.logger.Warn("Container is not running", zap.String("containerID", containerID))
					failures++
				} else {
					if failures > 0 {
						dcm.logger.Info("Container health recovered", zap.String("containerID", containerID))
					}
					failures = 0
					select {
					case healthChan <- true:
					default:
					}
					continue
				}

				if failures >= config.FailureThreshold {
					dcm.logger.Error("Container health check failed threshold",
						zap.String("containerID", containerID),
						zap.Int("failures", failures),
						zap.Int("threshold", config.FailureThreshold),
					)
					select {
					case healthChan <- false:
					default:
					}
					failures = 0 // Reset to avoid spam
				}
			}
		}
	}()

	return healthChan, nil
}

// StopHealthCheck stops the health check for a container
func (dcm *DockerContainerManager) StopHealthCheck(containerID string) {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	if cancelFunc, exists := dcm.healthChecks[containerID]; exists {
		cancelFunc()
		delete(dcm.healthChecks, containerID)
		dcm.logger.Debug("Health check stopped", zap.String("containerID", containerID))
	}
}

// Shutdown stops all health checks and cleans up resources
func (dcm *DockerContainerManager) Shutdown(ctx context.Context) error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	// Stop all health checks
	for containerID, cancelFunc := range dcm.healthChecks {
		cancelFunc()
		dcm.logger.Debug("Stopped health check during shutdown", zap.String("containerID", containerID))
	}
	dcm.healthChecks = make(map[string]context.CancelFunc)

	// Close Docker client
	if dcm.client != nil {
		if err := dcm.client.Close(); err != nil {
			return errors.Wrap(err, "failed to close Docker client")
		}
	}

	dcm.logger.Info("Container manager shutdown completed")
	return nil
}