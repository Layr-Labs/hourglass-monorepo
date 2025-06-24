package serverPerformer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ContainerStatus int

const (
	ContainerStatusPending ContainerStatus = iota
	ContainerStatusActive
	ContainerStatusExpired
)

func (cs ContainerStatus) String() string {
	switch cs {
	case ContainerStatusPending:
		return "Pending"
	case ContainerStatusActive:
		return "Active"
	case ContainerStatusExpired:
		return "Expired"
	default:
		return "Unknown"
	}
}

type ContainerHealthState int

const (
	ContainerHealthUnknown ContainerHealthState = iota
	ContainerHealthHealthy
	ContainerHealthUnhealthy
	ContainerHealthStopped
	ContainerHealthCrashed
)

func (chs ContainerHealthState) String() string {
	switch chs {
	case ContainerHealthHealthy:
		return "Healthy"
	case ContainerHealthUnhealthy:
		return "Unhealthy"
	case ContainerHealthStopped:
		return "Stopped"
	case ContainerHealthCrashed:
		return "Crashed"
	default:
		return "Unknown"
	}
}

type ApplicationHealthState int

const (
	AppHealthUnknown ApplicationHealthState = iota
	AppHealthHealthy
	AppHealthUnhealthy
	AppHealthNotReady
)

func (ahs ApplicationHealthState) String() string {
	switch ahs {
	case AppHealthHealthy:
		return "Healthy"
	case AppHealthUnhealthy:
		return "Unhealthy"
	case AppHealthNotReady:
		return "NotReady"
	default:
		return "Unknown"
	}
}

type ContainerMetadata struct {
	Info            *containerManager.ContainerInfo
	Image           avsPerformer.PerformerImage
	ActivationTime  int64
	Status          ContainerStatus
	Client          performerV1.PerformerServiceClient
	Endpoint        string
	DeploymentID    string
	CreatedAt       time.Time
	ActivatedAt     *time.Time
	LastHealthCheck time.Time
	RegistryURL     string
	ArtifactDigest  string
}

type ContainerHealthStatus struct {
	ContainerID       string
	Status            ContainerStatus
	ActivationTime    int64
	ContainerHealth   ContainerHealthState
	ApplicationHealth ApplicationHealthState
	LastHealthCheck   time.Time
	Endpoint          string
	Image             string
}

type AvsPerformerServer struct {
	config           *avsPerformer.AvsPerformerConfig
	logger           *zap.Logger
	containerManager containerManager.ContainerManager

	// Legacy fields for backwards compatibility
	containerInfo   *containerManager.ContainerInfo
	performerClient performerV1.PerformerServiceClient

	// New container management
	containers       map[string]*ContainerMetadata
	currentContainer string
	containerMu      sync.RWMutex

	peeringFetcher  peering.IPeeringDataFetcher
	aggregatorPeers []*peering.OperatorPeerInfo

	// Application health check cancellation
	healthCheckCancel context.CancelFunc
	healthCheckMu     sync.Mutex
}

func NewAvsPerformerServer(
	config *avsPerformer.AvsPerformerConfig,
	peeringFetcher peering.IPeeringDataFetcher,
	logger *zap.Logger,
) (*AvsPerformerServer, error) {
	// Create container manager
	containerMgr, err := containerManager.NewDockerContainerManager(
		&containerManager.ContainerManagerConfig{
			DefaultStartTimeout: 30 * time.Second,
			DefaultStopTimeout:  10 * time.Second,
			DefaultHealthCheckConfig: &containerManager.HealthCheckConfig{
				Enabled:          true,
				Interval:         5 * time.Second,
				Timeout:          2 * time.Second,
				Retries:          3,
				StartPeriod:      10 * time.Second,
				FailureThreshold: 3,
			},
		},
		logger,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container manager")
	}

	return &AvsPerformerServer{
		config:           config,
		logger:           logger,
		containerManager: containerMgr,
		peeringFetcher:   peeringFetcher,
		containers:       make(map[string]*ContainerMetadata),
	}, nil
}

const containerPort = 8080

// logFields returns common logging fields for this server
func (aps *AvsPerformerServer) logFields() []zap.Field {
	return []zap.Field{
		zap.String("avsAddress", aps.config.AvsAddress),
	}
}

// logFieldsWithContainer returns common logging fields including container ID
func (aps *AvsPerformerServer) logFieldsWithContainer() []zap.Field {
	fields := aps.logFields()
	if aps.containerInfo != nil {
		fields = append(fields, zap.String("containerID", aps.containerInfo.ID))
	}
	return fields
}

// cleanupFailedContainer removes a failed container and logs errors
func (aps *AvsPerformerServer) cleanupFailedContainer(ctx context.Context, containerID string) {
	if removeErr := aps.containerManager.Remove(ctx, containerID, true); removeErr != nil {
		aps.logger.Error("Failed to remove failed container",
			zap.String("containerID", containerID),
			zap.Error(removeErr),
		)
	}
}

// createLivenessConfig creates a standard liveness configuration
func (aps *AvsPerformerServer) createLivenessConfig() *containerManager.LivenessConfig {
	return &containerManager.LivenessConfig{
		HealthCheckConfig: containerManager.HealthCheckConfig{
			Enabled:          true,
			Interval:         5 * time.Second,
			Timeout:          2 * time.Second,
			Retries:          3,
			StartPeriod:      10 * time.Second,
			FailureThreshold: 3,
		},
		RestartPolicy: containerManager.RestartPolicy{
			Enabled:            true,
			MaxRestarts:        5,
			RestartDelay:       2 * time.Second,
			BackoffMultiplier:  2.0,
			MaxBackoffDelay:    30 * time.Second,
			RestartTimeout:     60 * time.Second,
			RestartOnCrash:     true,
			RestartOnOOM:       true,
			RestartOnUnhealthy: true,
		},
		ResourceThresholds: containerManager.ResourceThresholds{
			CPUThreshold:    90.0,
			MemoryThreshold: 90.0,
			RestartOnCPU:    false,
			RestartOnMemory: false,
		},
		ResourceMonitoring:    true,
		ResourceCheckInterval: 30 * time.Second,
	}
}

// retryWithBackoff executes a function with exponential backoff retry logic
func (aps *AvsPerformerServer) retryWithBackoff(ctx context.Context, operation func() error, operationName string) error {
	retries := []uint64{1, 3, 5, 10, 20}
	for i, retry := range retries {
		err := operation()
		if err == nil {
			return nil
		}

		aps.logger.Error(fmt.Sprintf("Failed %s", operationName),
			append(aps.logFields(), zap.Error(err))...,
		)

		if i == len(retries)-1 {
			aps.logger.Info(fmt.Sprintf("Giving up on %s", operationName),
				append(aps.logFields(), zap.Error(err))...,
			)
			return err
		}

		time.Sleep(time.Duration(retry) * time.Second)
	}
	return fmt.Errorf("failed %s after retries", operationName)
}

// generateDeploymentID creates a unique deployment ID
func generateDeploymentID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes) // Error is safe to ignore for random bytes
	return hex.EncodeToString(bytes)
}

// getActiveContainer returns the current active container metadata
func (aps *AvsPerformerServer) getActiveContainer() *ContainerMetadata {
	aps.containerMu.RLock()
	defer aps.containerMu.RUnlock()

	if aps.currentContainer == "" {
		return nil
	}
	return aps.containers[aps.currentContainer]
}

// addPendingContainer adds a new container to the pending queue
func (aps *AvsPerformerServer) addPendingContainer(container *ContainerMetadata) {
	aps.containerMu.Lock()
	defer aps.containerMu.Unlock()

	aps.containers[container.Info.ID] = container
	aps.logger.Info("Added pending container",
		append(aps.logFields(),
			zap.String("containerID", container.Info.ID),
			zap.String("deploymentID", container.DeploymentID),
			zap.Int64("activationTime", container.ActivationTime),
		)...,
	)
}

// checkAndActivatePendingContainer performs lazy container activation
func (aps *AvsPerformerServer) checkAndActivatePendingContainer(currentTime int64) {
	aps.containerMu.Lock()
	defer aps.containerMu.Unlock()

	// Find the container that should be active now
	var targetContainer *ContainerMetadata
	for _, container := range aps.containers {
		if container.ActivationTime <= currentTime &&
			container.Status == ContainerStatusPending {
			if targetContainer == nil ||
				container.ActivationTime > targetContainer.ActivationTime {
				targetContainer = container
			}
		}
	}

	if targetContainer != nil && targetContainer.Info.ID != aps.currentContainer {
		// Unlock before calling activateContainer to avoid deadlock
		aps.containerMu.Unlock()
		err := aps.activateContainer(targetContainer.Info.ID)
		aps.containerMu.Lock()

		if err != nil {
			aps.logger.Error("Failed to activate pending container",
				append(aps.logFields(),
					zap.String("containerID", targetContainer.Info.ID),
					zap.Error(err),
				)...,
			)
		}
	}
}

// activateContainer safely switches to a new active container
func (aps *AvsPerformerServer) activateContainer(containerID string) error {
	aps.containerMu.Lock()
	defer aps.containerMu.Unlock()

	// Verify container exists and is pending
	metadata, exists := aps.containers[containerID]
	if !exists {
		return fmt.Errorf("container %s not found", containerID)
	}

	if metadata.Status != ContainerStatusPending {
		return fmt.Errorf("container %s is not pending (status: %s)", containerID, metadata.Status)
	}

	// TODO: Add health checks here
	// if !aps.isContainerHealthy(containerID) {
	//     return fmt.Errorf("cannot activate unhealthy container %s", containerID)
	// }

	// Mark old container as expired
	if aps.currentContainer != "" {
		if oldContainer, exists := aps.containers[aps.currentContainer]; exists {
			oldContainer.Status = ContainerStatusExpired
		}
	}

	// Activate new container
	aps.currentContainer = containerID
	metadata.Status = ContainerStatusActive
	now := time.Now()
	metadata.ActivatedAt = &now

	// Update legacy fields for backwards compatibility
	aps.containerInfo = metadata.Info
	aps.performerClient = metadata.Client

	aps.logger.Info("Activated new container",
		append(aps.logFields(),
			zap.String("newContainerID", containerID),
			zap.String("deploymentID", metadata.DeploymentID),
		)...,
	)

	return nil
}

// DeployNewPerformerVersion deploys a new container version with scheduled activation
func (aps *AvsPerformerServer) DeployNewPerformerVersion(
	ctx context.Context,
	registryURL string,
	digest string,
	activationTime int64,
) (string, error) {
	if activationTime <= time.Now().Unix() {
		return "", fmt.Errorf("activation time must be in the future")
	}

	// Create image config from RPC parameters
	image := avsPerformer.PerformerImage{
		Repository: registryURL,
		Tag:        digest,
	}

	// Create and start container
	containerInfo, endpoint, err := aps.createAndStartContainerWithImage(ctx, image)
	if err != nil {
		return "", err
	}

	// Create client
	client, err := aps.createPerformerClient(endpoint)
	if err != nil {
		aps.cleanupFailedContainer(ctx, containerInfo.ID)
		return "", err
	}

	deploymentID := generateDeploymentID()

	// Add to pending containers
	container := &ContainerMetadata{
		Info:           containerInfo,
		Image:          image,
		ActivationTime: activationTime,
		Status:         ContainerStatusPending,
		Client:         client,
		Endpoint:       endpoint,
		DeploymentID:   deploymentID,
		CreatedAt:      time.Now(),
		RegistryURL:    registryURL,
		ArtifactDigest: digest,
	}

	aps.addPendingContainer(container)

	aps.logger.Info("Deployed new performer version",
		append(aps.logFields(),
			zap.String("deploymentID", deploymentID),
			zap.String("containerID", containerInfo.ID),
			zap.String("registryURL", registryURL),
			zap.String("digest", digest),
			zap.Int64("activationTime", activationTime),
		)...,
	)

	return deploymentID, nil
}

// GetAllContainerHealth returns comprehensive health status of all containers
func (aps *AvsPerformerServer) GetAllContainerHealth(ctx context.Context) ([]ContainerHealthStatus, error) {
	aps.containerMu.RLock()
	defer aps.containerMu.RUnlock()

	var statuses []ContainerHealthStatus
	for containerID, metadata := range aps.containers {
		status := ContainerHealthStatus{
			ContainerID:     containerID,
			Status:          metadata.Status,
			ActivationTime:  metadata.ActivationTime,
			Endpoint:        metadata.Endpoint,
			Image:           fmt.Sprintf("%s:%s", metadata.Image.Repository, metadata.Image.Tag),
			LastHealthCheck: metadata.LastHealthCheck,
		}

		// Get container-level health
		status.ContainerHealth = aps.getContainerHealth(containerID)

		// Get application-level health (only if container is running)
		if status.ContainerHealth == ContainerHealthHealthy {
			status.ApplicationHealth = aps.getApplicationHealth(ctx, metadata.Client)
		}

		statuses = append(statuses, status)
	}
	return statuses, nil
}

// getContainerHealth checks the container-level health
func (aps *AvsPerformerServer) getContainerHealth(containerID string) ContainerHealthState {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get container info from Docker
	containerInfo, err := aps.containerManager.Inspect(ctx, containerID)
	if err != nil {
		return ContainerHealthUnknown
	}

	// Check container status
	switch strings.ToLower(containerInfo.Status) {
	case "running":
		return ContainerHealthHealthy
	case "exited", "dead":
		return ContainerHealthCrashed
	case "paused", "restarting":
		return ContainerHealthUnhealthy
	default:
		return ContainerHealthUnknown
	}
}

// getApplicationHealth checks the application-level health via gRPC
func (aps *AvsPerformerServer) getApplicationHealth(ctx context.Context, client performerV1.PerformerServiceClient) ApplicationHealthState {
	if client == nil {
		return AppHealthNotReady
	}

	// Create a short timeout for health checks
	healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := client.HealthCheck(healthCtx, &performerV1.HealthCheckRequest{})
	if err != nil {
		return AppHealthUnhealthy
	}

	// If we got a response without error, consider it healthy
	return AppHealthHealthy
}


// reapExpiredContainers removes containers that are no longer needed
func (aps *AvsPerformerServer) reapExpiredContainers() {
	aps.containerMu.Lock()
	defer aps.containerMu.Unlock()

	currentTime := time.Now().Unix()
	expiredContainers := []string{}

	for containerID, metadata := range aps.containers {
		// Skip current active container
		if containerID == aps.currentContainer {
			continue
		}

		// Check if container has been superseded
		if aps.isContainerExpired(metadata, currentTime) {
			expiredContainers = append(expiredContainers, containerID)
		}
	}

	// Clean up expired containers
	for _, containerID := range expiredContainers {
		aps.removeExpiredContainer(containerID)
	}

	if len(expiredContainers) > 0 {
		aps.logger.Info("Reaped expired containers",
			append(aps.logFields(),
				zap.Int("count", len(expiredContainers)),
				zap.Strings("containerIDs", expiredContainers),
			)...,
		)
	}
}

// isContainerExpired checks if a container should be reaped
func (aps *AvsPerformerServer) isContainerExpired(metadata *ContainerMetadata, currentTime int64) bool {
	// 1. Pending container that's way past its activation time
	if metadata.Status == ContainerStatusPending &&
		currentTime > metadata.ActivationTime+3600 { // 1 hour grace period
		return true
	}

	// 2. Previously active containers that have been replaced
	if metadata.Status == ContainerStatusExpired {
		return true
	}

	return false
}

// removeExpiredContainer stops and removes a container from the system
func (aps *AvsPerformerServer) removeExpiredContainer(containerID string) {
	metadata := aps.containers[containerID]

	aps.logger.Info("Removing expired container",
		append(aps.logFields(),
			zap.String("containerID", containerID),
			zap.String("deploymentID", metadata.DeploymentID),
			zap.Int64("activationTime", metadata.ActivationTime),
		)...,
	)

	// Stop and remove container
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := aps.containerManager.Stop(ctx, containerID, 10*time.Second); err != nil {
		aps.logger.Error("Failed to stop expired container",
			append(aps.logFields(),
				zap.String("containerID", containerID),
				zap.Error(err),
			)...,
		)
	}
	if err := aps.containerManager.Remove(ctx, containerID, true); err != nil {
		aps.logger.Error("Failed to remove expired container",
			append(aps.logFields(),
				zap.String("containerID", containerID),
				zap.Error(err),
			)...,
		)
	}

	// Remove from our store
	delete(aps.containers, containerID)
}

// RemoveContainer allows manual removal of specific containers
func (aps *AvsPerformerServer) RemoveContainer(containerID string) error {
	aps.containerMu.Lock()
	defer aps.containerMu.Unlock()

	// Don't allow removal of active container
	if containerID == aps.currentContainer {
		return fmt.Errorf("cannot remove active container %s", containerID)
	}

	// Check if container exists
	if _, exists := aps.containers[containerID]; !exists {
		return fmt.Errorf("container %s not found", containerID)
	}

	// Remove the container
	aps.removeExpiredContainer(containerID)

	return nil
}

func (aps *AvsPerformerServer) fetchAggregatorPeerInfo(ctx context.Context) ([]*peering.OperatorPeerInfo, error) {
	var result []*peering.OperatorPeerInfo
	err := aps.retryWithBackoff(ctx, func() error {
		aggPeers, err := aps.peeringFetcher.ListAggregatorOperators(ctx, aps.config.AvsAddress)
		if err != nil {
			return err
		}
		result = aggPeers
		return nil
	}, "fetch aggregator peers")
	return result, err
}

// createAndStartContainer creates, starts, and initializes a container with proper error handling
func (aps *AvsPerformerServer) createAndStartContainer(ctx context.Context) (*containerManager.ContainerInfo, string, error) {
	return aps.createAndStartContainerWithImage(ctx, avsPerformer.PerformerImage{
		Repository: aps.config.Image.Repository,
		Tag:        aps.config.Image.Tag,
	})
}

// createAndStartContainerWithImage creates, starts, and initializes a container with specific image
func (aps *AvsPerformerServer) createAndStartContainerWithImage(ctx context.Context, image avsPerformer.PerformerImage) (*containerManager.ContainerInfo, string, error) {
	// Create container configuration
	containerConfig := containerManager.CreateDefaultContainerConfig(
		aps.config.AvsAddress,
		image.Repository,
		image.Tag,
		containerPort,
		aps.config.PerformerNetworkName,
	)

	aps.logger.Info("Using container configuration",
		append(aps.logFields(),
			zap.String("hostname", containerConfig.Hostname),
			zap.String("image", containerConfig.Image),
			zap.String("networkName", containerConfig.NetworkName),
		)...,
	)

	// Create the container
	containerInfo, err := aps.containerManager.Create(ctx, containerConfig)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create container")
	}

	// Start the container with cleanup on failure
	if err := aps.containerManager.Start(ctx, containerInfo.ID); err != nil {
		aps.cleanupFailedContainer(ctx, containerInfo.ID)
		return nil, "", errors.Wrap(err, "failed to start container")
	}

	// Wait for the container to be running
	if err := aps.containerManager.WaitForRunning(ctx, containerInfo.ID, 30*time.Second); err != nil {
		aps.cleanupFailedContainer(ctx, containerInfo.ID)
		return nil, "", errors.Wrap(err, "failed to wait for container to be running")
	}

	// Get updated container information with port mappings
	updatedInfo, err := aps.containerManager.Inspect(ctx, containerInfo.ID)
	if err != nil {
		aps.cleanupFailedContainer(ctx, containerInfo.ID)
		return nil, "", errors.Wrap(err, "failed to inspect container")
	}

	// Get the container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, containerPort, aps.config.PerformerNetworkName)
	if err != nil {
		aps.cleanupFailedContainer(ctx, containerInfo.ID)
		return nil, "", errors.Wrap(err, "failed to get container endpoint")
	}

	aps.logger.Info("Container started successfully",
		append(aps.logFields(),
			zap.String("containerID", containerInfo.ID),
			zap.String("endpoint", endpoint),
		)...,
	)

	return updatedInfo, endpoint, nil
}

// createPerformerClient creates a new performer client with error handling
func (aps *AvsPerformerServer) createPerformerClient(endpoint string) (performerV1.PerformerServiceClient, error) {
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(endpoint, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create performer client")
	}
	return perfClient, nil
}

// startLivenessMonitoring starts liveness monitoring for a container
func (aps *AvsPerformerServer) startLivenessMonitoring(ctx context.Context, containerID string) {
	livenessConfig := aps.createLivenessConfig()
	eventChan, err := aps.containerManager.StartLivenessMonitoring(ctx, containerID, livenessConfig)
	if err != nil {
		aps.logger.Warn("Failed to start liveness monitoring", zap.Error(err))
	} else {
		go aps.monitorContainerEvents(ctx, eventChan)
	}
}

func (aps *AvsPerformerServer) Initialize(ctx context.Context) error {
	// Fetch aggregator peer information
	aggregatorPeers, err := aps.fetchAggregatorPeerInfo(ctx)
	if err != nil {
		return err
	}
	aps.aggregatorPeers = aggregatorPeers
	aps.logger.Info("Fetched aggregator peers",
		append(aps.logFields(), zap.Any("aggregatorPeers", aps.aggregatorPeers))...,
	)

	// Create and start initial container
	containerInfo, endpoint, err := aps.createAndStartContainer(ctx)
	if err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown after container creation failure")
		}
		return err
	}

	// Create performer client
	perfClient, err := aps.createPerformerClient(endpoint)
	if err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown container after client creation failure")
		}
		return err
	}

	// Create initial container metadata and add to store
	initialContainer := &ContainerMetadata{
		Info: containerInfo,
		Image: avsPerformer.PerformerImage{
			Repository: aps.config.Image.Repository,
			Tag:        aps.config.Image.Tag,
		},
		ActivationTime: time.Now().Unix(), // Already active
		Status:         ContainerStatusActive,
		Client:         perfClient,
		Endpoint:       endpoint,
		DeploymentID:   "initial",
		CreatedAt:      time.Now(),
		RegistryURL:    aps.config.Image.Repository,
		ArtifactDigest: aps.config.Image.Tag,
	}
	now := time.Now()
	initialContainer.ActivatedAt = &now

	// Add to container store and set as current
	aps.containerMu.Lock()
	aps.containers[containerInfo.ID] = initialContainer
	aps.currentContainer = containerInfo.ID
	aps.containerMu.Unlock()

	// Set legacy fields for backwards compatibility
	aps.containerInfo = containerInfo
	aps.performerClient = perfClient

	// Start liveness monitoring
	aps.startLivenessMonitoring(ctx, containerInfo.ID)

	// Start application-level health checking
	aps.startApplicationHealthCheck(ctx)

	return nil
}

// monitorContainerEvents monitors container lifecycle events and handles them appropriately
func (aps *AvsPerformerServer) monitorContainerEvents(ctx context.Context, eventChan <-chan containerManager.ContainerEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				aps.logger.Info("Container event channel closed")
				return
			}
			aps.handleContainerEvent(ctx, event)
		}
	}
}

// handleContainerEvent processes individual container events
func (aps *AvsPerformerServer) handleContainerEvent(ctx context.Context, event containerManager.ContainerEvent) {
	aps.logger.Info("Container event received",
		append(aps.logFields(),
			zap.String("containerID", event.ContainerID),
			zap.String("eventType", string(event.Type)),
			zap.String("message", event.Message),
			zap.Int("restartCount", event.State.RestartCount),
		)...,
	)

	switch event.Type {
	case containerManager.EventStarted:
		aps.logger.Info("Container started successfully",
			append(aps.logFields(), zap.String("containerID", event.ContainerID))...,
		)

	case containerManager.EventCrashed:
		aps.logger.Error("Container crashed",
			append(aps.logFields(),
				zap.String("containerID", event.ContainerID),
				zap.Int("exitCode", event.State.ExitCode),
				zap.Int("restartCount", event.State.RestartCount),
				zap.String("error", event.State.Error),
			)...,
		)
		// Auto-restart is handled by containerManager based on RestartPolicy

	case containerManager.EventOOMKilled:
		aps.logger.Error("Container killed due to OOM",
			append(aps.logFields(),
				zap.String("containerID", event.ContainerID),
				zap.Int("restartCount", event.State.RestartCount),
			)...,
		)
		// Auto-restart is handled by containerManager

	case containerManager.EventRestarted:
		aps.logger.Info("Container restarted successfully",
			append(aps.logFields(),
				zap.String("containerID", event.ContainerID),
				zap.Int("restartCount", event.State.RestartCount),
			)...,
		)
		// Recreate performer client connection after restart
		go aps.recreatePerformerClient(ctx)

	case containerManager.EventRestartFailed:
		aps.logger.Error("Container restart failed",
			append(aps.logFields(),
				zap.String("containerID", event.ContainerID),
				zap.String("error", event.Message),
				zap.Int("restartCount", event.State.RestartCount),
			)...,
		)
		// Could potentially signal the executor to take additional action

		// Check if restart failed because container doesn't exist (needs recreation)
		if strings.Contains(event.Message, "recreation needed") {
			aps.logger.Info("Container recreation needed, attempting to recreate",
				append(aps.logFields(), zap.String("containerID", event.ContainerID))...,
			)
			go aps.recreateContainer(ctx)
		}

	case containerManager.EventHealthy:
		aps.logger.Debug("Container health recovered",
			append(aps.logFields(), zap.String("containerID", event.ContainerID))...,
		)

	case containerManager.EventUnhealthy:
		aps.logger.Warn("Container is unhealthy",
			append(aps.logFields(),
				zap.String("containerID", event.ContainerID),
				zap.String("reason", event.Message),
			)...,
		)
		// The container manager will handle auto-restart based on policy
		// Application can decide to trigger manual restart if needed

	case containerManager.EventRestarting:
		aps.logger.Info("Container is being restarted",
			append(aps.logFields(),
				zap.String("containerID", event.ContainerID),
				zap.String("reason", event.Message),
			)...,
		)
	}
}

// recreateContainer recreates a container that was killed/removed
func (aps *AvsPerformerServer) recreateContainer(ctx context.Context) {
	aps.logger.Info("Starting container recreation",
		append(aps.logFieldsWithContainer(), zap.String("previousContainerID", aps.containerInfo.ID))...,
	)

	// Stop monitoring the old container
	if aps.containerInfo != nil {
		aps.containerManager.StopLivenessMonitoring(aps.containerInfo.ID)
	}

	// Create and start new container
	containerInfo, endpoint, err := aps.createAndStartContainer(ctx)
	if err != nil {
		aps.logger.Error("Failed to recreate container",
			append(aps.logFields(), zap.Error(err))...,
		)
		return
	}

	// Update container info
	aps.containerInfo = containerInfo

	// Create new performer client
	perfClient, err := aps.createPerformerClient(endpoint)
	if err != nil {
		aps.logger.Error("Failed to create performer client for new container",
			append(aps.logFields(),
				zap.String("containerID", containerInfo.ID),
				zap.String("endpoint", endpoint),
				zap.Error(err),
			)...,
		)
		aps.cleanupFailedContainer(ctx, containerInfo.ID)
		return
	}

	aps.performerClient = perfClient

	// Start liveness monitoring for the new container
	aps.startLivenessMonitoring(ctx, containerInfo.ID)

	// Start new application-level health checking for the recreated container
	aps.startApplicationHealthCheck(ctx)

	aps.logger.Info("Container recreation completed successfully",
		append(aps.logFields(),
			zap.String("newContainerID", containerInfo.ID),
			zap.String("endpoint", endpoint),
		)...,
	)
}

// recreatePerformerClient recreates the performer client connection after container restart
func (aps *AvsPerformerServer) recreatePerformerClient(ctx context.Context) {
	// Wait a moment for the container to fully start
	time.Sleep(2 * time.Second)

	// Get updated container information
	updatedInfo, err := aps.containerManager.Inspect(ctx, aps.containerInfo.ID)
	if err != nil {
		aps.logger.Error("Failed to inspect container after restart",
			append(aps.logFields(), zap.Error(err))...,
		)
		return
	}
	aps.containerInfo = updatedInfo

	// Get the new container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, containerPort, aps.config.PerformerNetworkName)
	if err != nil {
		aps.logger.Error("Failed to get container endpoint after restart",
			append(aps.logFields(), zap.Error(err))...,
		)
		return
	}

	// Create new performer client
	perfClient, err := aps.createPerformerClient(endpoint)
	if err != nil {
		aps.logger.Error("Failed to recreate performer client after restart",
			append(aps.logFields(), zap.Error(err))...,
		)
		return
	}

	aps.performerClient = perfClient
	aps.logger.Info("Performer client recreated successfully after container restart",
		append(aps.logFields(), zap.String("endpoint", endpoint))...,
	)
}

// TriggerContainerRestart allows the application to manually trigger a container restart
func (aps *AvsPerformerServer) TriggerContainerRestart(reason string) error {
	if aps.containerManager == nil || aps.containerInfo == nil {
		return fmt.Errorf("container manager or container info not available")
	}

	aps.logger.Info("Triggering manual container restart",
		append(aps.logFieldsWithContainer(), zap.String("reason", reason))...,
	)

	return aps.containerManager.TriggerRestart(aps.containerInfo.ID, reason)
}

// startApplicationHealthCheck performs application-level health checks via gRPC
func (aps *AvsPerformerServer) startApplicationHealthCheck(ctx context.Context) {
	// Stop any existing health check
	aps.stopApplicationHealthCheck()

	// Create cancellable context for this health check
	healthCtx, cancel := context.WithCancel(ctx)

	aps.healthCheckMu.Lock()
	aps.healthCheckCancel = cancel
	aps.healthCheckMu.Unlock()

	aps.logger.Info("Starting application health check",
		aps.logFieldsWithContainer()...,
	)

	// Start the health check in a goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		consecutiveFailures := 0
		const maxConsecutiveFailures = 3

		// Add reaping counter - reap less frequently than health checks
		reapCounter := 0
		const reapInterval = 60 // Reap every 60 health checks (60 seconds)

		for {
			select {
			case <-healthCtx.Done():
				aps.logger.Debug("Application health check cancelled",
					aps.logFields()...,
				)
				return
			case <-ticker.C:
				// 1. Perform container reaping periodically
				reapCounter++
				if reapCounter >= reapInterval {
					aps.reapExpiredContainers()
					reapCounter = 0
				}

				// 2. Continue with existing health check logic
				if aps.performerClient == nil {
					continue
				}

				res, err := aps.performerClient.HealthCheck(healthCtx, &performerV1.HealthCheckRequest{})
				if err != nil {
					consecutiveFailures++
					aps.logger.Error("Failed to get health from performer",
						append(aps.logFields(),
							zap.Error(err),
							zap.Int("consecutiveFailures", consecutiveFailures),
						)...,
					)

					// Trigger container restart if we've had too many consecutive failures
					if consecutiveFailures >= maxConsecutiveFailures {
						aps.logger.Error("Application health check failed multiple times, triggering container restart",
							append(aps.logFields(), zap.Int("consecutiveFailures", consecutiveFailures))...,
						)

						if err := aps.TriggerContainerRestart(fmt.Sprintf("application health check failed %d consecutive times", consecutiveFailures)); err != nil {
							aps.logger.Error("Failed to trigger container restart",
								append(aps.logFields(), zap.Error(err))...,
							)
						}

						// Reset counter after triggering restart
						consecutiveFailures = 0
					}
					continue
				}

				// Reset failure counter on successful health check
				if consecutiveFailures > 0 {
					aps.logger.Info("Application health check recovered",
						append(aps.logFields(), zap.Int("previousFailures", consecutiveFailures))...,
					)
					consecutiveFailures = 0
				}

				aps.logger.Debug("Got health response",
					append(aps.logFields(), zap.String("status", res.Status.String()))...,
				)
			}
		}
	}()
}

// stopApplicationHealthCheck stops the current application health check
func (aps *AvsPerformerServer) stopApplicationHealthCheck() {
	aps.healthCheckMu.Lock()
	defer aps.healthCheckMu.Unlock()

	if aps.healthCheckCancel != nil {
		aps.healthCheckCancel()
		aps.healthCheckCancel = nil
		aps.logger.Debug("Stopped application health check",
			aps.logFields()...,
		)
	}
}

func (aps *AvsPerformerServer) ValidateTaskSignature(t *performerTask.PerformerTask) error {
	sig, err := bn254.NewSignatureFromBytes(t.Signature)
	if err != nil {
		aps.logger.Error("Failed to create signature from bytes",
			append(aps.logFields(), zap.Error(err))...,
		)
		return err
	}
	peer := util.Find(aps.aggregatorPeers, func(p *peering.OperatorPeerInfo) bool {
		return strings.EqualFold(p.OperatorAddress, t.AggregatorAddress)
	})
	if peer == nil {
		aps.logger.Error("Failed to find peer for task",
			append(aps.logFields(), zap.String("aggregatorAddress", t.AggregatorAddress))...,
		)
		return fmt.Errorf("failed to find peer for task")
	}

	isVerified := false

	// TODO(seanmcgary): this should verify the key against the expected aggregator operatorSetID
	for _, opset := range peer.OperatorSets {
		verfied, err := sig.Verify(opset.PublicKey, t.Payload)
		if err != nil {
			aps.logger.Error("Error verifying signature",
				append(aps.logFields(),
					zap.String("aggregatorAddress", t.AggregatorAddress),
					zap.Error(err),
				)...,
			)
			continue
		}
		if !verfied {
			aps.logger.Error("Failed to verify signature",
				append(aps.logFields(),
					zap.String("aggregatorAddress", t.AggregatorAddress),
					zap.Error(err),
				)...,
			)
			continue
		}
		aps.logger.Info("Signature verified with operator set",
			append(aps.logFields(),
				zap.String("aggregatorAddress", t.AggregatorAddress),
				zap.Uint32("opsetID", opset.OperatorSetID),
			)...,
		)
		isVerified = true
	}

	if !isVerified {
		aps.logger.Error("Failed to verify signature with any operator set",
			append(aps.logFields(), zap.String("aggregatorAddress", t.AggregatorAddress))...,
		)
		return fmt.Errorf("failed to verify signature with any operator set")
	}

	return nil
}

func (aps *AvsPerformerServer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	// 1. Lazy evaluation: activate pending containers if deadline reached
	aps.checkAndActivatePendingContainer(time.Now().Unix())

	// 2. Get current active container
	activeContainer := aps.getActiveContainer()
	if activeContainer == nil {
		return nil, fmt.Errorf("no active container available")
	}

	aps.logger.Info("Processing task", append(aps.logFields(), zap.Any("task", task))...)

	res, err := activeContainer.Client.ExecuteTask(ctx, &performerV1.TaskRequest{
		TaskId:  []byte(task.TaskID),
		Payload: task.Payload,
	})
	if err != nil {
		aps.logger.Error("Performer failed to handle task",
			append(aps.logFields(), zap.Error(err))...,
		)
		return nil, err
	}

	return performerTask.NewTaskResultFromResultProto(res), nil
}

func (aps *AvsPerformerServer) Shutdown() error {
	// Stop application health check first
	aps.stopApplicationHealthCheck()

	if aps.containerManager == nil {
		return nil
	}

	aps.logger.Info("Shutting down AVS performer server",
		aps.logFields()...,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Stop and remove all containers
	aps.containerMu.Lock()
	containerIDs := make([]string, 0, len(aps.containers))
	for containerID := range aps.containers {
		containerIDs = append(containerIDs, containerID)
	}
	aps.containerMu.Unlock()

	for _, containerID := range containerIDs {
		// Stop the container
		if err := aps.containerManager.Stop(ctx, containerID, 10*time.Second); err != nil {
			aps.logger.Error("Failed to stop container",
				append(aps.logFields(),
					zap.String("containerID", containerID),
					zap.Error(err),
				)...,
			)
		}

		// Remove the container
		if err := aps.containerManager.Remove(ctx, containerID, true); err != nil {
			aps.logger.Error("Failed to remove container",
				append(aps.logFields(),
					zap.String("containerID", containerID),
					zap.Error(err),
				)...,
			)
		}
	}

	// Clear container store
	aps.containerMu.Lock()
	aps.containers = make(map[string]*ContainerMetadata)
	aps.currentContainer = ""
	aps.containerMu.Unlock()

	// Shutdown the container manager
	if err := aps.containerManager.Shutdown(ctx); err != nil {
		aps.logger.Error("Failed to shutdown container manager",
			append(aps.logFields(), zap.Error(err))...,
		)
		return err
	}

	aps.logger.Info("AVS performer server shutdown completed",
		aps.logFields()...,
	)

	return nil
}
