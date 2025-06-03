package containerManager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// containerMonitor holds monitoring state for a container
type containerMonitor struct {
	containerID   string
	config        *LivenessConfig
	restartCount  int
	lastRestart   time.Time
	eventChan     chan ContainerEvent
	cancelFunc    context.CancelFunc
	restartPolicy RestartPolicy
}

// DockerContainerManager implements ContainerManager using Docker
type DockerContainerManager struct {
	client *client.Client
	config *ContainerManagerConfig
	logger *zap.Logger

	// Legacy health checks
	healthChecks map[string]context.CancelFunc

	// Enhanced liveness monitoring
	livenessMonitors map[string]*containerMonitor

	mu sync.RWMutex
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
	if config.DefaultLivenessConfig == nil {
		config.DefaultLivenessConfig = &LivenessConfig{
			HealthCheckConfig: *config.DefaultHealthCheckConfig,
			RestartPolicy: RestartPolicy{
				Enabled:            true,
				MaxRestarts:        5,
				RestartDelay:       2 * time.Second,
				BackoffMultiplier:  2.0,
				MaxBackoffDelay:    30 * time.Second,
				RestartTimeout:     60 * time.Second,
				RestartOnCrash:     true,
				RestartOnOOM:       true,
				RestartOnUnhealthy: false, // Let application decide
			},
			ResourceThresholds: ResourceThresholds{
				CPUThreshold:    90.0,
				MemoryThreshold: 90.0,
				RestartOnCPU:    false,
				RestartOnMemory: false,
			},
			MonitorEvents:         true,
			ResourceMonitoring:    true,
			ResourceCheckInterval: 30 * time.Second,
		}
	}

	return &DockerContainerManager{
		client:           dockerClient,
		config:           config,
		logger:           logger,
		healthChecks:     make(map[string]context.CancelFunc),
		livenessMonitors: make(map[string]*containerMonitor),
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
		AutoRemove:     config.AutoRemove,
		PortBindings:   config.PortBindings,
		Privileged:     config.Privileged,
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

	dcm.logger.Debug("Waiting for container to be running", 
		zap.String("containerID", containerID),
		zap.Duration("timeout", timeout),
	)

	for {
		select {
		case <-ctx.Done():
			// Get final container state for debugging
			if info, err := dcm.Inspect(context.Background(), containerID); err == nil {
				dcm.logger.Error("Timeout waiting for container to be running",
					zap.String("containerID", containerID),
					zap.String("status", info.Status),
					zap.Int("portCount", len(info.Ports)),
					zap.Any("ports", info.Ports),
				)
			}
			return errors.New("timeout waiting for container to be running")
		case <-ticker.C:
			running, err := dcm.IsRunning(ctx, containerID)
			if err != nil {
				dcm.logger.Debug("Failed to check container status", 
					zap.String("containerID", containerID),
					zap.Error(err),
				)
				return errors.Wrap(err, "failed to check container status")
			}

			dcm.logger.Debug("Container status check", 
				zap.String("containerID", containerID),
				zap.Bool("running", running),
			)

			if running {
				// Additional check to ensure ports are exposed
				info, err := dcm.Inspect(ctx, containerID)
				if err != nil {
					return errors.Wrap(err, "failed to inspect container")
				}

				dcm.logger.Debug("Container port inspection", 
					zap.String("containerID", containerID),
					zap.Int("portCount", len(info.Ports)),
					zap.Any("ports", info.Ports),
				)

				// For custom networks, we don't need host port bindings
				// The container is accessible via hostname:containerPort
				if len(info.Networks) > 0 {
					// Check if container is on a custom network
					for networkName := range info.Networks {
						if networkName != "bridge" && networkName != "host" && networkName != "none" {
							dcm.logger.Info("Container is running on custom network", 
								zap.String("containerID", containerID),
								zap.String("network", networkName),
							)
							return nil
						}
					}
				}

				// For bridge network, check if ports are exposed
				if len(info.Ports) > 0 {
					dcm.logger.Info("Container is running with ports exposed", 
						zap.String("containerID", containerID),
						zap.Any("ports", info.Ports),
					)
					return nil
				}

				// If no ports but container is running, log warning and continue waiting
				dcm.logger.Warn("Container is running but no ports are exposed", 
					zap.String("containerID", containerID),
				)
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

	// Stop all liveness monitors
	for containerID, monitor := range dcm.livenessMonitors {
		monitor.cancelFunc()
		close(monitor.eventChan)
		dcm.logger.Debug("Stopped liveness monitor during shutdown", zap.String("containerID", containerID))
	}
	dcm.livenessMonitors = make(map[string]*containerMonitor)

	// Close Docker client
	if dcm.client != nil {
		if err := dcm.client.Close(); err != nil {
			return errors.Wrap(err, "failed to close Docker client")
		}
	}

	dcm.logger.Info("Container manager shutdown completed")
	return nil
}

// StartLivenessMonitoring starts comprehensive container monitoring with auto-restart
func (dcm *DockerContainerManager) StartLivenessMonitoring(ctx context.Context, containerID string, config *LivenessConfig) (<-chan ContainerEvent, error) {
	if config == nil {
		config = dcm.config.DefaultLivenessConfig
	}

	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	// Stop existing monitor if any
	if monitor, exists := dcm.livenessMonitors[containerID]; exists {
		monitor.cancelFunc()
		close(monitor.eventChan)
	}

	// Create new monitor
	monitorCtx, cancel := context.WithCancel(ctx)
	eventChan := make(chan ContainerEvent, 10) // Buffered channel

	monitor := &containerMonitor{
		containerID:   containerID,
		config:        config,
		restartCount:  0,
		eventChan:     eventChan,
		cancelFunc:    cancel,
		restartPolicy: config.RestartPolicy,
	}

	dcm.livenessMonitors[containerID] = monitor

	// TODO: Emit metric for liveness monitor started
	dcm.logger.Info("Started liveness monitoring",
		zap.String("containerID", containerID),
		zap.Bool("restartEnabled", config.RestartPolicy.Enabled),
		zap.Bool("eventMonitoring", config.MonitorEvents),
		zap.Bool("resourceMonitoring", config.ResourceMonitoring),
	)

	// Docker event monitoring is disabled as it provides no additional value
	// over health check polling and adds complexity. Container restarts are
	// handled by health check failures and application-level monitoring.
	_ = config.MonitorEvents // Ignore this setting

	// Start monitoring goroutines
	go dcm.monitorContainerLiveness(monitorCtx, monitor)

	if config.ResourceMonitoring {
		go dcm.monitorContainerResources(monitorCtx, monitor)
	}

	return eventChan, nil
}

// StopLivenessMonitoring stops liveness monitoring for a container
func (dcm *DockerContainerManager) StopLivenessMonitoring(containerID string) {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	if monitor, exists := dcm.livenessMonitors[containerID]; exists {
		monitor.cancelFunc()
		close(monitor.eventChan)
		delete(dcm.livenessMonitors, containerID)

		// TODO: Emit metric for liveness monitor stopped
		dcm.logger.Debug("Stopped liveness monitoring", zap.String("containerID", containerID))
	}
}

// GetContainerState returns detailed container state information
func (dcm *DockerContainerManager) GetContainerState(ctx context.Context, containerID string) (*ContainerState, error) {
	containerJSON, err := dcm.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to inspect container")
	}

	// Parse time strings from Docker API
	startedAt, err := time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt)
	if err != nil {
		startedAt = time.Now() // Fallback to current time if parse fails
	}

	state := &ContainerState{
		Status:     containerJSON.State.Status,
		ExitCode:   containerJSON.State.ExitCode,
		StartedAt:  startedAt,
		OOMKilled:  containerJSON.State.OOMKilled,
		Error:      containerJSON.State.Error,
		Restarting: containerJSON.State.Restarting,
	}

	if containerJSON.State.FinishedAt != "" {
		if finishedAt, err := time.Parse(time.RFC3339Nano, containerJSON.State.FinishedAt); err == nil {
			state.FinishedAt = &finishedAt
		}
	}

	// Get restart count from monitor if available
	dcm.mu.RLock()
	if monitor, exists := dcm.livenessMonitors[containerID]; exists {
		state.RestartCount = monitor.restartCount
	}
	dcm.mu.RUnlock()

	return state, nil
}

// RestartContainer restarts a container with the specified timeout
func (dcm *DockerContainerManager) RestartContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = dcm.config.DefaultStopTimeout
	}

	// TODO: Emit metric for manual container restart
	dcm.logger.Info("Restarting container",
		zap.String("containerID", containerID),
		zap.Duration("timeout", timeout),
	)

	if err := dcm.client.ContainerRestart(ctx, containerID, container.StopOptions{
		Timeout: func() *int { t := int(timeout.Seconds()); return &t }(),
	}); err != nil {
		// TODO: Emit metric for restart failure
		return errors.Wrap(err, "failed to restart container")
	}

	// Update restart count in monitor
	dcm.mu.Lock()
	if monitor, exists := dcm.livenessMonitors[containerID]; exists {
		monitor.restartCount++
		monitor.lastRestart = time.Now()
	}
	dcm.mu.Unlock()

	// TODO: Emit metric for successful restart
	dcm.logger.Info("Container restarted successfully", zap.String("containerID", containerID))
	return nil
}

// SetRestartPolicy updates the restart policy for a container
func (dcm *DockerContainerManager) SetRestartPolicy(containerID string, policy RestartPolicy) error {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	if monitor, exists := dcm.livenessMonitors[containerID]; exists {
		monitor.restartPolicy = policy
		dcm.logger.Info("Updated restart policy",
			zap.String("containerID", containerID),
			zap.Bool("enabled", policy.Enabled),
			zap.Int("maxRestarts", policy.MaxRestarts),
		)
		return nil
	}

	return fmt.Errorf("no liveness monitor found for container %s", containerID)
}

// GetResourceUsage returns current resource usage for a container
func (dcm *DockerContainerManager) GetResourceUsage(ctx context.Context, containerID string) (*ResourceUsage, error) {
	stats, err := dcm.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get container stats")
	}
	defer stats.Body.Close()

	decoder := json.NewDecoder(stats.Body)
	var stat map[string]interface{}
	if err := decoder.Decode(&stat); err != nil {
		return nil, errors.Wrap(err, "failed to decode container stats")
	}

	// Calculate CPU percentage - simplified version for now
	cpuPercent := 0.0 // TODO: Implement proper CPU calculation with generic stats

	// Extract memory stats from generic map
	memoryPercent := float64(0)
	memoryUsage := int64(0)
	memoryLimit := int64(0)
	
	if memoryStats, ok := stat["memory_stats"].(map[string]interface{}); ok {
		if usage, ok := memoryStats["usage"].(float64); ok {
			memoryUsage = int64(usage)
		}
		if limit, ok := memoryStats["limit"].(float64); ok {
			memoryLimit = int64(limit)
			if memoryLimit > 0 {
				memoryPercent = float64(memoryUsage) / float64(memoryLimit) * 100.0
			}
		}
	}

	// Extract network stats from generic map
	var networkRx, networkTx int64
	if networks, ok := stat["networks"].(map[string]interface{}); ok {
		for _, network := range networks {
			if netMap, ok := network.(map[string]interface{}); ok {
				if rxBytes, ok := netMap["rx_bytes"].(float64); ok {
					networkRx += int64(rxBytes)
				}
				if txBytes, ok := netMap["tx_bytes"].(float64); ok {
					networkTx += int64(txBytes)
				}
			}
		}
	}

	// Extract disk I/O stats from generic map
	var diskRead, diskWrite int64
	if blkioStats, ok := stat["blkio_stats"].(map[string]interface{}); ok {
		if ioServiceBytes, ok := blkioStats["io_service_bytes_recursive"].([]interface{}); ok {
			for _, blkio := range ioServiceBytes {
				if blkioMap, ok := blkio.(map[string]interface{}); ok {
					if op, ok := blkioMap["op"].(string); ok {
						if value, ok := blkioMap["value"].(float64); ok {
							if op == "Read" {
								diskRead += int64(value)
							} else if op == "Write" {
								diskWrite += int64(value)
							}
						}
					}
				}
			}
		}
	}

	return &ResourceUsage{
		CPUPercent:    cpuPercent,
		MemoryUsage:   memoryUsage,
		MemoryLimit:   memoryLimit,
		MemoryPercent: memoryPercent,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
		DiskRead:      diskRead,
		DiskWrite:     diskWrite,
		Timestamp:     time.Now(),
	}, nil
}

// TriggerRestart manually triggers a container restart (for serverPerformer to call)
func (dcm *DockerContainerManager) TriggerRestart(containerID string, reason string) error {
	dcm.mu.RLock()
	monitor, exists := dcm.livenessMonitors[containerID]
	dcm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no liveness monitor found for container %s", containerID)
	}

	if !monitor.restartPolicy.Enabled {
		return fmt.Errorf("restart policy is disabled for container %s", containerID)
	}

	// Send restart event to monitor
	event := ContainerEvent{
		ContainerID: containerID,
		Type:        EventUnhealthy,
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Manual restart triggered: %s", reason),
	}

	select {
	case monitor.eventChan <- event:
		dcm.logger.Info("Manual restart triggered",
			zap.String("containerID", containerID),
			zap.String("reason", reason),
		)
		return nil
	default:
		return fmt.Errorf("failed to send restart event for container %s", containerID)
	}
}

// monitorContainerLiveness monitors container health and triggers events
func (dcm *DockerContainerManager) monitorContainerLiveness(ctx context.Context, monitor *containerMonitor) {
	ticker := time.NewTicker(monitor.config.HealthCheckConfig.Interval)
	defer ticker.Stop()

	failures := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			running, err := dcm.IsRunning(ctx, monitor.containerID)
			if err != nil {
				// TODO: Emit metric for health check error
				dcm.logger.Error("Health check failed",
					zap.String("containerID", monitor.containerID),
					zap.Error(err),
				)
				failures++
			} else if !running {
				dcm.logger.Warn("Container is not running", zap.String("containerID", monitor.containerID))
				failures++
			} else {
				if failures > 0 {
					// Send healthy event
					event := ContainerEvent{
						ContainerID: monitor.containerID,
						Type:        EventHealthy,
						Timestamp:   time.Now(),
						Message:     "Container health recovered",
					}

					select {
					case monitor.eventChan <- event:
					default:
					}

					dcm.logger.Info("Container health recovered", zap.String("containerID", monitor.containerID))
				}
				failures = 0
				continue
			}

			if failures >= monitor.config.HealthCheckConfig.FailureThreshold {
				// Send unhealthy event
				event := ContainerEvent{
					ContainerID: monitor.containerID,
					Type:        EventUnhealthy,
					Timestamp:   time.Now(),
					Message:     fmt.Sprintf("Health check failed %d times", failures),
				}

				select {
				case monitor.eventChan <- event:
				default:
				}

				// TODO: Emit metric for health check failure threshold reached
				dcm.logger.Error("Container health check failed threshold",
					zap.String("containerID", monitor.containerID),
					zap.Int("failures", failures),
					zap.Int("threshold", monitor.config.HealthCheckConfig.FailureThreshold),
				)
				failures = 0 // Reset to avoid spam
			}
		}
	}
}

// monitorContainerResources monitors container resource usage
func (dcm *DockerContainerManager) monitorContainerResources(ctx context.Context, monitor *containerMonitor) {
	ticker := time.NewTicker(monitor.config.ResourceCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			usage, err := dcm.GetResourceUsage(ctx, monitor.containerID)
			if err != nil {
				// TODO: Emit metric for resource monitoring error
				dcm.logger.Debug("Failed to get resource usage",
					zap.String("containerID", monitor.containerID),
					zap.Error(err),
				)
				continue
			}

			// TODO: Emit metrics for resource usage
			dcm.logger.Debug("Container resource usage",
				zap.String("containerID", monitor.containerID),
				zap.Float64("cpuPercent", usage.CPUPercent),
				zap.Float64("memoryPercent", usage.MemoryPercent),
			)

			// Check thresholds
			thresholds := monitor.config.ResourceThresholds

			if thresholds.RestartOnCPU && usage.CPUPercent > thresholds.CPUThreshold {
				// TODO: Emit metric for CPU threshold exceeded
				dcm.logger.Warn("CPU threshold exceeded",
					zap.String("containerID", monitor.containerID),
					zap.Float64("usage", usage.CPUPercent),
					zap.Float64("threshold", thresholds.CPUThreshold),
				)

				// Trigger restart if enabled
				if monitor.restartPolicy.Enabled {
					event := ContainerEvent{
						ContainerID: monitor.containerID,
						Type:        EventUnhealthy,
						Timestamp:   time.Now(),
						Message:     fmt.Sprintf("CPU usage %.1f%% exceeded threshold %.1f%%", usage.CPUPercent, thresholds.CPUThreshold),
					}

					select {
					case monitor.eventChan <- event:
					default:
					}
				}
			}

			if thresholds.RestartOnMemory && usage.MemoryPercent > thresholds.MemoryThreshold {
				// TODO: Emit metric for memory threshold exceeded
				dcm.logger.Warn("Memory threshold exceeded",
					zap.String("containerID", monitor.containerID),
					zap.Float64("usage", usage.MemoryPercent),
					zap.Float64("threshold", thresholds.MemoryThreshold),
				)

				// Trigger restart if enabled
				if monitor.restartPolicy.Enabled {
					event := ContainerEvent{
						ContainerID: monitor.containerID,
						Type:        EventUnhealthy,
						Timestamp:   time.Now(),
						Message:     fmt.Sprintf("Memory usage %.1f%% exceeded threshold %.1f%%", usage.MemoryPercent, thresholds.MemoryThreshold),
					}

					select {
					case monitor.eventChan <- event:
					default:
					}
				}
			}
		}
	}
}
