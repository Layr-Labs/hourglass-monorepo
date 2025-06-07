package containerManager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
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

	// Pull image if not present locally
	if err := dcm.ensureImageExists(ctx, config.Image); err != nil {
		return nil, errors.Wrap(err, "failed to ensure image exists")
	}

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
		// Handle the case where container is already being removed (AutoRemove=true)
		if strings.Contains(err.Error(), "removal of container") && strings.Contains(err.Error(), "is already in progress") {
			dcm.logger.Debug("Container removal already in progress (likely AutoRemove)", zap.String("containerID", containerID))
			return nil
		}
		// Handle the case where container doesn't exist (already removed)
		if strings.Contains(err.Error(), "No such container") {
			dcm.logger.Debug("Container already removed", zap.String("containerID", containerID))
			return nil
		}
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
	} else {
		// Merge with defaults for any zero values
		if config.Interval == 0 {
			config.Interval = dcm.config.DefaultHealthCheckConfig.Interval
		}
		if config.FailureThreshold == 0 {
			config.FailureThreshold = dcm.config.DefaultHealthCheckConfig.FailureThreshold
		}
		if config.Timeout == 0 {
			config.Timeout = dcm.config.DefaultHealthCheckConfig.Timeout
		}
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
		dcm.logger.Debug("Stopped liveness monitor during shutdown", zap.String("containerID", containerID))

		// Close channels in a goroutine to avoid blocking
		go func(ch chan ContainerEvent) {
			time.Sleep(10 * time.Millisecond)
			close(ch)
		}(monitor.eventChan)
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
	} else {
		// Merge with defaults for any zero values
		if config.HealthCheckConfig.Interval == 0 {
			config.HealthCheckConfig.Interval = dcm.config.DefaultLivenessConfig.HealthCheckConfig.Interval
		}
		if config.HealthCheckConfig.FailureThreshold == 0 {
			config.HealthCheckConfig.FailureThreshold = dcm.config.DefaultLivenessConfig.HealthCheckConfig.FailureThreshold
		}
		if config.ResourceCheckInterval == 0 {
			config.ResourceCheckInterval = dcm.config.DefaultLivenessConfig.ResourceCheckInterval
		}
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

	// Start monitoring goroutines
	go dcm.monitorContainerLiveness(monitorCtx, monitor)

	if config.ResourceMonitoring {
		go dcm.monitorContainerResources(monitorCtx, monitor)
	}

	// Start Docker event monitoring to detect crashes and OOM kills
	// This is essential for detecting container failures that won't show up in health checks
	go dcm.monitorDockerEvents(monitorCtx, monitor)

	return eventChan, nil
}

// StopLivenessMonitoring stops liveness monitoring for a container
func (dcm *DockerContainerManager) StopLivenessMonitoring(containerID string) {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()

	if monitor, exists := dcm.livenessMonitors[containerID]; exists {
		monitor.cancelFunc()
		delete(dcm.livenessMonitors, containerID)

		// Close the channel in a goroutine to avoid blocking
		go func() {
			// Give some time for monitoring goroutines to exit
			time.Sleep(10 * time.Millisecond)
			close(monitor.eventChan)
		}()

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
		dcm.logger.Warn("No liveness monitor found for container, ignoring restart request",
			zap.String("containerID", containerID),
			zap.String("reason", reason),
		)
		return fmt.Errorf("no liveness monitor found for container %s", containerID)
	}

	if !monitor.restartPolicy.Enabled {
		return fmt.Errorf("restart policy is disabled for container %s", containerID)
	}

	dcm.logger.Info("Manual restart triggered",
		zap.String("containerID", containerID),
		zap.String("reason", reason),
	)

	// Send restart event to monitor (safely handle closed channel)
	event := ContainerEvent{
		ContainerID: containerID,
		Type:        EventUnhealthy,
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Manual restart triggered: %s", reason),
	}

	select {
	case monitor.eventChan <- event:
	default:
		// Channel might be closed, that's ok
	}

	// Actually perform the restart directly
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return dcm.attemptRestart(ctx, monitor, reason)
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
				// Check if container doesn't exist (was killed/removed)
				if strings.Contains(err.Error(), "No such container") {
					dcm.logger.Warn("Container no longer exists (likely killed)",
						zap.String("containerID", monitor.containerID),
						zap.Error(err),
					)
					// Treat missing container as a crash that needs restart
					failures = monitor.config.HealthCheckConfig.FailureThreshold
				} else {
					// TODO: Emit metric for health check error
					dcm.logger.Error("Health check failed",
						zap.String("containerID", monitor.containerID),
						zap.Error(err),
					)
					failures++
				}
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
					case <-ctx.Done():
						return
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

				// Trigger restart if policy allows and we haven't hit the limit
				if monitor.restartPolicy.Enabled && monitor.restartPolicy.RestartOnUnhealthy {
					if err := dcm.attemptRestart(ctx, monitor, "health check failures"); err != nil {
						dcm.logger.Error("Failed to restart unhealthy container",
							zap.String("containerID", monitor.containerID),
							zap.Error(err),
						)
					}
				}

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
					case <-ctx.Done():
						return
					default:
					}

					if err := dcm.attemptRestart(ctx, monitor, fmt.Sprintf("CPU usage %.1f%% exceeded threshold %.1f%%", usage.CPUPercent, thresholds.CPUThreshold)); err != nil {
						dcm.logger.Error("Failed to restart container due to CPU threshold",
							zap.String("containerID", monitor.containerID),
							zap.Error(err),
						)
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
					case <-ctx.Done():
						return
					default:
					}

					if err := dcm.attemptRestart(ctx, monitor, fmt.Sprintf("Memory usage %.1f%% exceeded threshold %.1f%%", usage.MemoryPercent, thresholds.MemoryThreshold)); err != nil {
						dcm.logger.Error("Failed to restart container due to memory threshold",
							zap.String("containerID", monitor.containerID),
							zap.Error(err),
						)
					}
				}
			}
		}
	}
}

// attemptRestart attempts to restart a container if restart policy allows
func (dcm *DockerContainerManager) attemptRestart(ctx context.Context, monitor *containerMonitor, reason string) error {
	// Check if we've exceeded the maximum restart count
	if monitor.restartPolicy.MaxRestarts > 0 && monitor.restartCount >= monitor.restartPolicy.MaxRestarts {
		dcm.logger.Warn("Container restart limit reached",
			zap.String("containerID", monitor.containerID),
			zap.Int("restartCount", monitor.restartCount),
			zap.Int("maxRestarts", monitor.restartPolicy.MaxRestarts),
			zap.String("reason", reason),
		)

		// Send restart failed event
		event := ContainerEvent{
			ContainerID: monitor.containerID,
			Type:        EventRestartFailed,
			Timestamp:   time.Now(),
			Message:     fmt.Sprintf("Restart limit reached (%d/%d): %s", monitor.restartCount, monitor.restartPolicy.MaxRestarts, reason),
		}

		select {
		case monitor.eventChan <- event:
		default:
		}

		return fmt.Errorf("restart limit reached: %d/%d", monitor.restartCount, monitor.restartPolicy.MaxRestarts)
	}

	// Apply restart delay with backoff
	delay := dcm.calculateRestartDelay(monitor)
	if delay > 0 {
		dcm.logger.Info("Applying restart delay",
			zap.String("containerID", monitor.containerID),
			zap.Duration("delay", delay),
			zap.Int("restartCount", monitor.restartCount),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	// Send restarting event (safely handle closed channel)
	event := ContainerEvent{
		ContainerID: monitor.containerID,
		Type:        EventRestarting,
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Restarting container (attempt %d/%d): %s", monitor.restartCount+1, monitor.restartPolicy.MaxRestarts, reason),
	}

	select {
	case monitor.eventChan <- event:
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Channel might be closed, that's ok
	}

	// Create restart context with timeout
	restartCtx := ctx
	if monitor.restartPolicy.RestartTimeout > 0 {
		var cancel context.CancelFunc
		restartCtx, cancel = context.WithTimeout(ctx, monitor.restartPolicy.RestartTimeout)
		defer cancel()
	}

	// Perform the restart
	if err := dcm.RestartContainer(restartCtx, monitor.containerID, dcm.config.DefaultStopTimeout); err != nil {
		// Check if container doesn't exist, in which case we need to recreate it
		if strings.Contains(err.Error(), "No such container") {
			dcm.logger.Info("Container doesn't exist, recreation needed but not supported in container manager",
				zap.String("containerID", monitor.containerID),
				zap.String("reason", reason),
			)

			// Send restart failed event indicating recreation is needed
			failEvent := ContainerEvent{
				ContainerID: monitor.containerID,
				Type:        EventRestartFailed,
				Timestamp:   time.Now(),
				Message:     fmt.Sprintf("Container recreation needed but not implemented (attempt %d): %s", monitor.restartCount+1, err.Error()),
			}

			select {
			case monitor.eventChan <- failEvent:
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Channel might be closed, that's ok
			}

			// For tests and backward compatibility, don't return error here
			// The application layer (serverPerformer) handles recreation
			dcm.logger.Debug("Container restart completed with recreation needed signal",
				zap.String("containerID", monitor.containerID),
			)
			return nil
		}

		// Send restart failed event for other errors
		failEvent := ContainerEvent{
			ContainerID: monitor.containerID,
			Type:        EventRestartFailed,
			Timestamp:   time.Now(),
			Message:     fmt.Sprintf("Restart failed (attempt %d): %s", monitor.restartCount+1, err.Error()),
		}

		select {
		case monitor.eventChan <- failEvent:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Channel might be closed, that's ok
		}

		return err
	}

	// Send restart success event (safely handle closed channel)
	successEvent := ContainerEvent{
		ContainerID: monitor.containerID,
		Type:        EventRestarted,
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Container restarted successfully (attempt %d)", monitor.restartCount),
	}

	select {
	case monitor.eventChan <- successEvent:
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Channel might be closed, that's ok
	}

	dcm.logger.Info("Container restarted successfully",
		zap.String("containerID", monitor.containerID),
		zap.String("reason", reason),
		zap.Int("restartCount", monitor.restartCount),
	)

	return nil
}

// calculateRestartDelay calculates the delay before attempting a restart based on backoff policy
func (dcm *DockerContainerManager) calculateRestartDelay(monitor *containerMonitor) time.Duration {
	if monitor.restartCount == 0 {
		return monitor.restartPolicy.RestartDelay
	}

	// Apply exponential backoff
	delay := monitor.restartPolicy.RestartDelay
	for i := 0; i < monitor.restartCount && delay < monitor.restartPolicy.MaxBackoffDelay; i++ {
		delay = time.Duration(float64(delay) * monitor.restartPolicy.BackoffMultiplier)
	}

	// Cap at max backoff delay
	if delay > monitor.restartPolicy.MaxBackoffDelay {
		delay = monitor.restartPolicy.MaxBackoffDelay
	}

	return delay
}

// monitorDockerEvents monitors Docker events for container crashes and OOM kills
func (dcm *DockerContainerManager) monitorDockerEvents(ctx context.Context, monitor *containerMonitor) {
	// Create event filters for this specific container
	eventOptions := events.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("container", monitor.containerID),
			filters.Arg("event", "die"),
			filters.Arg("event", "oom"),
		),
	}

	eventChan, errChan := dcm.client.Events(ctx, eventOptions)

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errChan:
			if err != nil {
				dcm.logger.Error("Docker event monitoring error",
					zap.String("containerID", monitor.containerID),
					zap.Error(err),
				)
				// Try to reconnect after a delay
				time.Sleep(5 * time.Second)
				eventChan, errChan = dcm.client.Events(ctx, eventOptions)
			}
		case event := <-eventChan:
			dcm.handleDockerEvent(ctx, monitor, event)
		}
	}
}

// handleDockerEvent processes Docker events and triggers appropriate actions
func (dcm *DockerContainerManager) handleDockerEvent(ctx context.Context, monitor *containerMonitor, dockerEvent events.Message) {
	dcm.logger.Debug("Docker event received",
		zap.String("containerID", monitor.containerID),
		zap.String("action", string(dockerEvent.Action)),
		zap.String("status", dockerEvent.Status),
		zap.Any("attributes", dockerEvent.Actor.Attributes),
	)

	switch string(dockerEvent.Action) {
	case "die":
		// Container crashed or was stopped
		exitCode := 0
		if exitCodeStr, exists := dockerEvent.Actor.Attributes["exitCode"]; exists {
			if parsedCode, err := strconv.Atoi(exitCodeStr); err == nil {
				exitCode = parsedCode
			}
		}

		isOOM := false
		if oomKilled, exists := dockerEvent.Actor.Attributes["oomKilled"]; exists {
			isOOM = oomKilled == "true"
		}

		var eventType ContainerEventType
		var message string

		if isOOM {
			eventType = EventOOMKilled
			message = "Container was killed due to OOM"
		} else if exitCode != 0 {
			eventType = EventCrashed
			message = fmt.Sprintf("Container crashed with exit code %d", exitCode)
		} else {
			// Normal shutdown, no action needed
			eventType = EventStopped
			message = "Container stopped normally"
		}

		// Send event
		event := ContainerEvent{
			ContainerID: monitor.containerID,
			Type:        eventType,
			Timestamp:   time.Now(),
			Message:     message,
			State: ContainerState{
				ExitCode:     exitCode,
				OOMKilled:    isOOM,
				RestartCount: monitor.restartCount,
			},
		}

		select {
		case monitor.eventChan <- event:
		default:
		}

		// Trigger restart if policy allows
		if monitor.restartPolicy.Enabled {
			shouldRestart := false
			reason := ""

			if isOOM && monitor.restartPolicy.RestartOnOOM {
				shouldRestart = true
				reason = "OOM kill detected"
			} else if exitCode != 0 && monitor.restartPolicy.RestartOnCrash {
				shouldRestart = true
				reason = fmt.Sprintf("crash detected (exit code %d)", exitCode)
			}

			if shouldRestart {
				if err := dcm.attemptRestart(ctx, monitor, reason); err != nil {
					dcm.logger.Error("Failed to restart container after Docker event",
						zap.String("containerID", monitor.containerID),
						zap.String("reason", reason),
						zap.Error(err),
					)
				}
			}
		}

	case "oom":
		// OOM event (may come before die event)
		dcm.logger.Warn("Container OOM event detected",
			zap.String("containerID", monitor.containerID),
		)

		event := ContainerEvent{
			ContainerID: monitor.containerID,
			Type:        EventOOMKilled,
			Timestamp:   time.Now(),
			Message:     "Container received OOM signal",
			State: ContainerState{
				OOMKilled:    true,
				RestartCount: monitor.restartCount,
			},
		}

		select {
		case monitor.eventChan <- event:
		default:
		}
	}
}

// ensureImageExists checks if an image exists locally and pulls it if not
func (dcm *DockerContainerManager) ensureImageExists(ctx context.Context, imageName string) error {
	// Check if image exists locally
	images, err := dcm.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list images")
	}

	// Check if the image is already present
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				dcm.logger.Debug("Image already exists locally", zap.String("image", imageName))
				return nil
			}
		}
	}

	// Image not found locally, pull it
	dcm.logger.Info("Pulling image", zap.String("image", imageName))

	pullOptions := image.PullOptions{}

	// Pull the image
	pullResponse, err := dcm.client.ImagePull(ctx, imageName, pullOptions)
	if err != nil {
		return errors.Wrap(err, "failed to pull image")
	}
	defer pullResponse.Close()

	// Read the pull response to completion (required for the pull to actually complete)
	_, err = io.Copy(io.Discard, pullResponse)
	if err != nil {
		return errors.Wrap(err, "failed to read pull response")
	}

	dcm.logger.Info("Image pulled successfully", zap.String("image", imageName))
	return nil
}
