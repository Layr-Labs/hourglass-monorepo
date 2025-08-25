package avsContainerPerformer

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	healthV1 "github.com/Layr-Labs/protocol-apis/gen/protos/grpc/health/v1"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	internalContainerPort                   = 8080
	maxConsecutiveApplicationHealthFailures = 3
	defaultApplicationHealthCheckInterval   = 15 * time.Second
	defaultDeploymentTimeout                = 1 * time.Minute
	defaultCleanupTimeout                   = 5 * time.Second
	defaultRunningWaitTimeout               = 30 * time.Second
)

// PerformerContainer holds all information about a container
type PerformerContainer struct {
	performerID     string
	info            *containerManager.ContainerInfo
	client          *avsPerformerClient.PerformerClient
	eventChan       <-chan containerManager.ContainerEvent
	performerHealth *avsPerformer.PerformerHealth
	statusChan      chan avsPerformer.PerformerStatusEvent
	image           avsPerformer.PerformerImage
	status          avsPerformer.PerformerResourceStatus
}

type AvsContainerPerformer struct {
	config *avsPerformer.AvsPerformerConfig
	logger *zap.Logger

	// Container tracking
	containerManager      containerManager.ContainerManager
	currentContainer      atomic.Value
	nextContainer         *PerformerContainer
	performerContainersMu sync.Mutex

	// Task tracking
	taskWaitGroups   map[string]*sync.WaitGroup
	taskWaitGroupsMu sync.Mutex

	// Draining tracking
	drainingPerformers   map[string]struct{}
	drainingPerformersMu sync.Mutex

	// Deployment tracking
	activeDeploymentMu sync.Mutex
}

// NewAvsContainerPerformerWithContainerManager creates a new AvsContainerPerformer with the provided container manager
// Necessary for injection of container manager behavior.
func NewAvsContainerPerformerWithContainerManager(
	config *avsPerformer.AvsPerformerConfig,
	logger *zap.Logger,
	containerManager containerManager.ContainerManager,
) *AvsContainerPerformer {
	// Set default health check interval if not specified
	if config.ApplicationHealthCheckInterval == 0 {
		config.ApplicationHealthCheckInterval = defaultApplicationHealthCheckInterval
	}

	return &AvsContainerPerformer{
		config:             config,
		logger:             logger,
		containerManager:   containerManager,
		taskWaitGroups:     make(map[string]*sync.WaitGroup),
		drainingPerformers: make(map[string]struct{}),
	}
}

// NewAvsContainerPerformer creates a new AvsContainerPerformer
func NewAvsContainerPerformer(
	config *avsPerformer.AvsPerformerConfig,
	logger *zap.Logger,
) (*AvsContainerPerformer, error) {
	containerMgr, err := containerManager.NewDockerContainerManager(
		containerManager.DefaultContainerManagerConfig(),
		logger,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create container manager for executor: %v", err)
	}
	// Set default health check interval if not specified
	if config.ApplicationHealthCheckInterval == 0 {
		config.ApplicationHealthCheckInterval = defaultApplicationHealthCheckInterval
	}

	return &AvsContainerPerformer{
		config:             config,
		logger:             logger,
		containerManager:   containerMgr,
		taskWaitGroups:     make(map[string]*sync.WaitGroup),
		drainingPerformers: make(map[string]struct{}),
	}, nil
}

// cleanupFailedContainer removes a container using a fresh context to ensure cleanup succeeds
// even if the original context is cancelled or timed out
func (aps *AvsContainerPerformer) cleanupFailedContainer(containerID string, failureReason string) {
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), defaultCleanupTimeout)
	defer cleanupCancel()

	if err := aps.containerManager.Remove(cleanupCtx, containerID, true); err != nil {
		aps.logger.Error("Failed to remove container during cleanup",
			zap.String("containerID", containerID),
			zap.String("failureReason", failureReason),
			zap.Error(err),
		)
	}
}

// generatePerformerID generates a unique performer ID
func (aps *AvsContainerPerformer) generatePerformerID() string {
	return fmt.Sprintf("performer-%s-%s", aps.config.AvsAddress, uuid.New().String())
}

func (aps *AvsContainerPerformer) buildDockerEnvsFromConfig(image avsPerformer.PerformerImage) []string {
	dockerEnvs := make([]string, 0)
	for _, env := range image.Envs {
		val := env.Value
		if env.ValueFromEnv != "" {
			val = os.Getenv(env.ValueFromEnv)
		}
		dockerEnvs = append(dockerEnvs, fmt.Sprintf("%s=%s", env.Name, val))
	}
	return dockerEnvs
}

// createAndStartContainer creates, starts, and prepares a container for the AVS performer
func (aps *AvsContainerPerformer) createAndStartContainer(
	ctx context.Context,
	avsAddress string,
	image avsPerformer.PerformerImage,
	containerConfig *containerManager.ContainerConfig,
	livenessConfig *containerManager.LivenessConfig,
) (*PerformerContainer, error) {

	// Create the container
	containerInfo, err := aps.containerManager.Create(ctx, containerConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	// Start the container
	if err := aps.containerManager.Start(ctx, containerInfo.ID); err != nil {
		// Clean up on failure
		aps.cleanupFailedContainer(containerInfo.ID, "failed to start container")
		return nil, errors.Wrap(err, "failed to start container")
	}

	// Wait for the container to be running
	if err := aps.containerManager.WaitForRunning(ctx, containerInfo.ID, defaultRunningWaitTimeout); err != nil {
		// Clean up on failure
		aps.cleanupFailedContainer(containerInfo.ID, "failed to wait for container to be running")
		return nil, errors.Wrap(err, "failed to wait for container to be running")
	}

	// Get updated container information with port mappings
	updatedInfo, err := aps.containerManager.Inspect(ctx, containerInfo.ID)
	if err != nil {
		// Clean up on failure
		aps.cleanupFailedContainer(containerInfo.ID, "failed to inspect container")
		return nil, errors.Wrap(err, "failed to inspect container")
	}

	// Get the container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, internalContainerPort, containerConfig.NetworkName)
	if err != nil {
		// Clean up on failure
		aps.cleanupFailedContainer(containerInfo.ID, "failed to get container endpoint")
		return nil, errors.Wrap(err, "failed to get container endpoint")
	}

	aps.logger.Info("Container created and started successfully",
		zap.String("avsAddress", avsAddress),
		zap.String("containerID", updatedInfo.ID),
		zap.String("endpoint", endpoint),
	)

	// Create performer client
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(endpoint, true)
	if err != nil {
		// Clean up on failure
		aps.cleanupFailedContainer(updatedInfo.ID, "failed to create performer client")
		return nil, errors.Wrap(err, "failed to create performer client")
	}

	// Start liveness monitoring for this container
	eventChan, err := aps.containerManager.StartLivenessMonitoring(ctx, updatedInfo.ID, livenessConfig)
	if err != nil {
		// Clean up on failure
		aps.cleanupFailedContainer(updatedInfo.ID, "failed to start liveness monitoring")
		return nil, errors.Wrap(err, "failed to start liveness monitoring")
	}

	performerID := aps.generatePerformerID()
	aps.logger.Info("Container created and monitoring started",
		zap.String("avsAddress", avsAddress),
		zap.String("performerID", performerID),
		zap.String("containerID", updatedInfo.ID),
		zap.String("endpoint", endpoint),
	)
	// Create the container instance with all components
	return &PerformerContainer{
		performerID: performerID,
		info:        updatedInfo,
		client:      perfClient,
		eventChan:   eventChan,
		image:       image,
		statusChan:  make(chan avsPerformer.PerformerStatusEvent, 10),
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy: true,
			LastHealthCheck:    time.Now(),
		},
	}, nil
}

func (aps *AvsContainerPerformer) Initialize(ctx context.Context) error {
	aps.performerContainersMu.Lock()
	defer aps.performerContainersMu.Unlock()

	// Check if we should start with a container loaded
	// Skip container creation if image info is empty (for deployment-based initialization)
	if aps.config.Image.Repository == "" || aps.config.Image.Tag == "" {
		aps.logger.Info("Starting PerformerServer without initial container.",
			zap.String("avsAddress", aps.config.AvsAddress),
		)
		return nil
	}

	// Create and start container
	performerContainer, err := aps.createAndStartContainer(
		ctx,
		aps.config.AvsAddress,
		aps.config.Image,
		containerManager.CreateDefaultContainerConfig(
			aps.config.AvsAddress,
			aps.config.Image.Repository,
			aps.config.Image.Tag,
			aps.config.Image.Digest,
			internalContainerPort,
			aps.config.PerformerNetworkName,
			aps.buildDockerEnvsFromConfig(aps.config.Image),
		),
		containerManager.NewDefaultAvsPerformerLivenessConfig(),
	)
	if err != nil {
		return err
	}

	performerContainer.status = avsPerformer.PerformerResourceStatusInService
	aps.currentContainer.Store(performerContainer)

	// Start monitoring events for the new container
	go aps.monitorContainerEvents(ctx, performerContainer)

	return nil
}

// monitorContainerEvents monitors container lifecycle events and performs periodic application health checks
func (aps *AvsContainerPerformer) monitorContainerEvents(ctx context.Context, container *PerformerContainer) {
	// Create a ticker for periodic application health checks
	appHealthCheckTicker := time.NewTicker(aps.config.ApplicationHealthCheckInterval)
	defer appHealthCheckTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-container.eventChan:
			if !ok {
				aps.logger.Info("Container event channel closed",
					zap.String("performerID", container.performerID),
					zap.String("containerID", container.info.ID),
				)
				return
			}
			aps.handleContainerEvent(ctx, event, container)
		case <-appHealthCheckTicker.C:
			// Perform periodic application health checks for containers that are Docker-healthy
			aps.performPeriodicApplicationHealthChecks(ctx)
		}
	}
}

// handleContainerEvent processes individual container events
func (aps *AvsContainerPerformer) handleContainerEvent(ctx context.Context, event containerManager.ContainerEvent, targetContainer *PerformerContainer) {
	aps.logger.Info("Container event received",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", targetContainer.performerID),
		zap.String("containerID", event.ContainerID),
		zap.String("eventType", string(event.Type)),
		zap.String("message", event.Message),
		zap.Int("restartCount", event.State.RestartCount),
	)

	aps.performerContainersMu.Lock()
	defer aps.performerContainersMu.Unlock()

	switch event.Type {
	case containerManager.EventStarted:
		aps.logger.Info("Container started successfully",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
		)
		targetContainer.performerHealth.ContainerIsHealthy = true
		return

	case containerManager.EventHealthy:
		aps.logger.Debug("Container Docker healthy signal received",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
		)
		// Update the container health status
		targetContainer.performerHealth.ContainerIsHealthy = true
		return

	case containerManager.EventCrashed:
		aps.logger.Error("Container crashed",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.Int("exitCode", event.State.ExitCode),
			zap.Int("restartCount", event.State.RestartCount),
			zap.String("error", event.State.Error),
		)
		// Auto-restart is handled by containerManager based on RestartPolicy
		targetContainer.performerHealth.ContainerIsHealthy = false
		targetContainer.performerHealth.ApplicationIsHealthy = false

	case containerManager.EventOOMKilled:
		aps.logger.Error("Container killed due to OOM",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.Int("restartCount", event.State.RestartCount),
		)
		// Auto-restart is handled by containerManager
		targetContainer.performerHealth.ContainerIsHealthy = false
		targetContainer.performerHealth.ApplicationIsHealthy = false

	case containerManager.EventRestarted:
		aps.logger.Info("Container restarted successfully",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.Int("restartCount", event.State.RestartCount),
		)
		// Reset health context since container was restarted
		targetContainer.performerHealth.ContainerIsHealthy = false
		targetContainer.performerHealth.ApplicationIsHealthy = false
		targetContainer.performerHealth.ConsecutiveApplicationHealthFailures = 0
		// Recreate performer client connection for the specific container that restarted after briefly waiting
		time.Sleep(2 * time.Second)
		aps.recreatePerformerClientForContainer(ctx, targetContainer)

	case containerManager.EventRestartFailed:
		aps.logger.Error("Container restart failed",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.String("error", event.Message),
			zap.Int("restartCount", event.State.RestartCount),
		)
		targetContainer.performerHealth.ContainerIsHealthy = false
		targetContainer.performerHealth.ApplicationIsHealthy = false

		if strings.Contains(event.Message, "recreation needed") {
			aps.logger.Info("Container recreation needed, attempting to recreate",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("containerID", event.ContainerID),
			)
			// Recreate container synchronously since we already hold the mutex
			aps.recreateContainer(ctx, targetContainer)
		}

	case containerManager.EventUnhealthy:
		aps.logger.Warn("Container is unhealthy",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.String("reason", event.Message),
		)
		// Update the container health status
		targetContainer.performerHealth.ContainerIsHealthy = false
		targetContainer.performerHealth.ApplicationIsHealthy = false

	case containerManager.EventRestarting:
		aps.logger.Info("Container is being restarted",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.String("reason", event.Message),
		)
		// Container is restarting, mark as unhealthy
		targetContainer.performerHealth.ContainerIsHealthy = false
		targetContainer.performerHealth.ApplicationIsHealthy = false
	}

	// Only reach this point if the event indicates an unhealthy container.
	if targetContainer.statusChan != nil {
		select {
		case targetContainer.statusChan <- avsPerformer.PerformerStatusEvent{
			Status:      avsPerformer.PerformerUnhealthy,
			PerformerID: targetContainer.performerID,
			Message:     "Container is unhealthy",
			Timestamp:   time.Now(),
		}:
			aps.logger.Info("Sent unhealthy status event",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("performerID", targetContainer.performerID),
			)
		default:
		}
	}
}

// recreateContainer recreates a container that was killed/removed by updating the fields in place
func (aps *AvsContainerPerformer) recreateContainer(ctx context.Context, targetContainer *PerformerContainer) {
	// Stop monitoring the old container
	aps.containerManager.StopLivenessMonitoring(targetContainer.info.ID)

	aps.logger.Info("Starting container recreation",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", targetContainer.performerID),
		zap.String("previousContainerID", targetContainer.info.ID),
		zap.String("repository", targetContainer.image.Repository),
		zap.String("tag", targetContainer.image.Tag),
		zap.String("digest", targetContainer.image.Digest),
	)

	// Create and start new container
	newContainer, err := aps.createAndStartContainer(
		ctx, aps.config.AvsAddress,
		targetContainer.image,
		containerManager.CreateDefaultContainerConfig(
			aps.config.AvsAddress,
			targetContainer.image.Repository,
			targetContainer.image.Tag,
			targetContainer.image.Digest,
			internalContainerPort,
			aps.config.PerformerNetworkName,
			aps.buildDockerEnvsFromConfig(targetContainer.image),
		), containerManager.NewDefaultAvsPerformerLivenessConfig())
	if err != nil {
		aps.logger.Error("Failed to recreate container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", targetContainer.performerID),
			zap.Error(err),
		)
		return
	}

	// Update the fields in the existing PerformerContainer reference
	// This keeps the same PerformerContainer object but with new container details
	targetContainer.info = newContainer.info
	targetContainer.client = newContainer.client
	targetContainer.eventChan = newContainer.eventChan
	targetContainer.performerHealth = newContainer.performerHealth

	aps.logger.Info("Container recreation completed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", targetContainer.performerID),
		zap.String("newContainerID", targetContainer.info.ID),
	)
}

// recreatePerformerClientForContainer recreates the performer client connection after container restart
func (aps *AvsContainerPerformer) recreatePerformerClientForContainer(ctx context.Context, container *PerformerContainer) {
	// Get updated container information
	updatedInfo, err := aps.containerManager.Inspect(ctx, container.info.ID)
	if err != nil {
		aps.logger.Error("Failed to inspect container after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", container.info.ID),
			zap.Error(err),
		)
		return
	}
	// Get the new container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, internalContainerPort, aps.config.PerformerNetworkName)
	if err != nil {
		aps.logger.Error("Failed to get container endpoint after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", container.info.ID),
			zap.Error(err),
		)
		return
	}

	// Create new performer client
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(endpoint, true)
	if err != nil {
		aps.logger.Error("Failed to recreate performer client after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", container.info.ID),
			zap.Error(err),
		)
		return
	}
	container.client = perfClient

	aps.logger.Info("Performer client recreated successfully after container restart",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", container.info.ID),
		zap.String("endpoint", endpoint),
	)
}

// TriggerContainerRestart allows the application to manually trigger a restart for a specific container
func (aps *AvsContainerPerformer) TriggerContainerRestart(container *PerformerContainer, reason string) error {
	if container == nil || container.info == nil {
		return fmt.Errorf("container info not available")
	}

	containerID := container.info.ID

	aps.logger.Info("Triggering manual container restart",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", containerID),
		zap.String("reason", reason),
	)

	return aps.containerManager.TriggerRestart(containerID, reason)
}

// checkApplicationHealth performs a single health check on the specified container
func (aps *AvsContainerPerformer) checkApplicationHealth(ctx context.Context, container *PerformerContainer) error {
	healthCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	res, err := container.client.HealthClient.Check(healthCtx, &healthV1.HealthCheckRequest{})
	if err != nil {
		return err
	}

	aps.logger.Sugar().Debug("Application health check successful",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", container.info.ID),
		zap.String("status", res.Status.String()),
	)

	return nil
}

// CreatePerformer creates a new performer and returns the creation result
// Always updates the nextContainer slot
func (aps *AvsContainerPerformer) CreatePerformer(
	ctx context.Context,
	image avsPerformer.PerformerImage,
) (*avsPerformer.PerformerCreationResult, error) {
	aps.performerContainersMu.Lock()
	defer aps.performerContainersMu.Unlock()

	aps.logger.Info("Starting performer deployment",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("imageRepository", image.Repository),
		zap.String("imageTag", image.Tag),
	)

	// Check if next container slot is already occupied
	if aps.nextContainer != nil {
		aps.logger.Error("Cannot create new performer, next container slot is already occupied",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("existingNextPerformerID", aps.nextContainer.performerID),
			zap.String("requestedImage", fmt.Sprintf("%s:%s", image.Repository, image.Tag)),
		)
		return nil, fmt.Errorf("a next performer already exists (ID: %s). Please remove it explicitly before creating a new one", aps.nextContainer.performerID)
	}

	// Create the new container instance
	newContainer, err := aps.createAndStartContainer(
		ctx,
		aps.config.AvsAddress,
		image,
		containerManager.CreateDefaultContainerConfig(
			aps.config.AvsAddress,
			image.Repository,
			image.Tag,
			image.Digest,
			internalContainerPort,
			aps.config.PerformerNetworkName,
			aps.buildDockerEnvsFromConfig(image),
		), containerManager.NewDefaultAvsPerformerLivenessConfig())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	// Always deploy as next container
	aps.nextContainer = newContainer
	aps.nextContainer.status = avsPerformer.PerformerResourceStatusStaged

	aps.logger.Info("Deployed as next performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", newContainer.performerID),
		zap.String("containerID", newContainer.info.ID),
	)

	// Start monitoring events for the new container
	go aps.monitorContainerEvents(ctx, newContainer)

	// Get endpoint for logging
	endpoint, _ := containerManager.GetContainerEndpoint(newContainer.info, internalContainerPort, aps.config.PerformerNetworkName)

	aps.logger.Info("Performer deployment started",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", newContainer.performerID),
		zap.String("containerID", newContainer.info.ID),
		zap.String("endpoint", endpoint),
	)

	return &avsPerformer.PerformerCreationResult{
		PerformerID: newContainer.performerID,
		StatusChan:  newContainer.statusChan,
	}, nil
}

// RemovePerformer removes a performer from the server by its performerID.
func (aps *AvsContainerPerformer) RemovePerformer(ctx context.Context, performerID string) error {
	aps.performerContainersMu.Lock()
	defer aps.performerContainersMu.Unlock()

	// Determine which container to remove
	var targetContainer *PerformerContainer

	current := aps.currentContainer.Load()
	if current != nil {
		currentContainer, ok := current.(*PerformerContainer)
		if !ok || currentContainer == nil {
			aps.logger.Error("Invalid type in currentContainer atomic.Value during removal")
			return fmt.Errorf("invalid performer type stored in currentContainer")
		}
		if currentContainer.performerID == performerID {
			targetContainer = currentContainer
			aps.currentContainer.Store((*PerformerContainer)(nil))
		}
	}

	if targetContainer == nil && aps.nextContainer != nil && aps.nextContainer.performerID == performerID {
		targetContainer = aps.nextContainer
		aps.nextContainer = nil
	}

	if targetContainer == nil {
		// Performer not found
		aps.logger.Error("Performer not found",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", performerID),
		)
		return fmt.Errorf("performer with ID %s not found", performerID)
	}

	// Close the status channel if it exists
	if targetContainer.statusChan != nil {
		close(targetContainer.statusChan)
	}

	// Shutdown the container (this will stop monitoring and remove it)
	if err := aps.shutdownContainer(ctx, aps.config.AvsAddress, targetContainer); err != nil {
		aps.logger.Error("Failed to shutdown container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", performerID),
			zap.String("containerID", targetContainer.info.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to remove performer: %w", err)
	}

	aps.logger.Info("Performer removed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", performerID),
	)

	return nil
}

// PromotePerformer promotes the specified performer to currentContainer
// If the performer is already current, it's a no-op success
// If the performer is next and healthy, it's promoted to current
// If the performer is not found or unhealthy, an error is returned
func (aps *AvsContainerPerformer) PromotePerformer(ctx context.Context, performerID string) error {
	aps.performerContainersMu.Lock()
	defer aps.performerContainersMu.Unlock()

	// Check if the performer is already the current container
	if current := aps.currentContainer.Load(); current != nil {
		currentContainer := current.(*PerformerContainer)
		if currentContainer != nil && currentContainer.performerID == performerID {
			aps.logger.Info("Performer is already the current container",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("performerID", performerID),
			)
			return nil
		}
	}

	// Check if the performer is the next container
	if aps.nextContainer == nil || aps.nextContainer.performerID != performerID {
		return fmt.Errorf("performer %s is not in the next deployment slot", performerID)
	}

	// Verify nextContainer is healthy
	if !aps.nextContainer.performerHealth.ContainerIsHealthy || !aps.nextContainer.performerHealth.ApplicationIsHealthy {
		return fmt.Errorf("cannot promote unhealthy performer %s (container health: %v, application health: %v)",
			performerID,
			aps.nextContainer.performerHealth.ContainerIsHealthy,
			aps.nextContainer.performerHealth.ApplicationIsHealthy)
	}

	aps.logger.Info("Promoting performer to current",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", performerID),
		zap.String("containerID", aps.nextContainer.info.ID),
	)

	// Initiate draining of the old container
	if current := aps.currentContainer.Load(); current != nil {
		oldContainer := current.(*PerformerContainer)
		if oldContainer != nil && oldContainer.info != nil {
			aps.logger.Info("Initiating drain of old performer",
				zap.String("oldPerformerID", oldContainer.performerID),
				zap.String("oldContainerID", oldContainer.info.ID),
			)

			// Start draining in a separate goroutine
			aps.startDrainAndRemove(oldContainer)
		}
	}

	// Promote next to current and update status
	aps.nextContainer.status = avsPerformer.PerformerResourceStatusInService
	aps.currentContainer.Store(aps.nextContainer)
	aps.nextContainer = nil

	aps.logger.Info("Performer promotion completed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", performerID),
	)

	return nil
}

func (aps *AvsContainerPerformer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	aps.logger.Sugar().Infow("Processing task", zap.Any("task", task))

	// Load current container using atomic accessor
	current := aps.currentContainer.Load()
	if current == nil {
		return nil, fmt.Errorf("no current container available to execute task")
	}

	currentContainer, ok := current.(*PerformerContainer)
	if !ok || currentContainer == nil || currentContainer.client == nil {
		return nil, fmt.Errorf("no current container available to execute task")
	}

	// Track this task with the performer's WaitGroup
	wg := aps.getOrCreateTaskWaitGroup(currentContainer.performerID)
	wg.Add(1)
	defer wg.Done()

	// Execute the task
	res, err := currentContainer.client.PerformerClient.ExecuteTask(ctx, &performerV1.TaskRequest{
		TaskId:  []byte(task.TaskID),
		Payload: task.Payload,
	})
	if err != nil {
		aps.logger.Sugar().Errorw("Performer failed to handle task",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", currentContainer.performerID),
			zap.Error(err),
		)
		return nil, err
	}
	aps.logger.Sugar().Infow("Performer handled task")

	return performerTask.NewTaskResultFromResultProto(res), nil
}

// shutdownContainer handles the shutdown of a single container instance
func (aps *AvsContainerPerformer) shutdownContainer(ctx context.Context, avsAddress string, container *PerformerContainer) error {
	if container == nil || container.info == nil || aps.containerManager == nil {
		return nil
	}

	aps.logger.Info("Shutting down container",
		zap.String("avsAddress", avsAddress),
		zap.String("performerID", container.performerID),
		zap.String("containerID", container.info.ID),
	)

	// Stop liveness monitoring for this container
	aps.containerManager.StopLivenessMonitoring(container.info.ID)

	// Stop the container
	if err := aps.containerManager.Stop(ctx, container.info.ID, 10*time.Second); err != nil {
		aps.logger.Error("Failed to stop container",
			zap.String("avsAddress", avsAddress),
			zap.String("performerID", container.performerID),
			zap.String("containerID", container.info.ID),
			zap.Error(err),
		)
	}

	// Remove the container
	if err := aps.containerManager.Remove(ctx, container.info.ID, true); err != nil {
		aps.logger.Error("Failed to remove container",
			zap.String("avsAddress", avsAddress),
			zap.String("performerID", container.performerID),
			zap.String("containerID", container.info.ID),
			zap.Error(err),
		)
		return err
	}

	aps.logger.Info("Container shutdown completed",
		zap.String("avsAddress", avsAddress),
		zap.String("performerID", container.performerID),
		zap.String("containerID", container.info.ID),
	)

	return nil
}

func (aps *AvsContainerPerformer) Shutdown() error {
	aps.performerContainersMu.Lock()
	defer aps.performerContainersMu.Unlock()

	aps.logger.Info("Shutting down AVS performer",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	// Clear draining performers to prevent further drain operations
	aps.drainingPerformersMu.Lock()
	for performerID := range aps.drainingPerformers {
		aps.logger.Info("Cancelling drain for performer",
			zap.String("performerID", performerID),
		)
		delete(aps.drainingPerformers, performerID)
	}
	aps.drainingPerformersMu.Unlock()

	// Shutdown both current and next containers
	var errs []error

	if current := aps.currentContainer.Load(); current != nil {
		currentContainer, ok := current.(*PerformerContainer)
		if !ok {
			aps.logger.Error("Invalid type in currentContainer atomic.Value during shutdown")
			errs = append(errs, fmt.Errorf("invalid performer type stored in currentContainer"))
		}
		if currentContainer != nil && currentContainer.info != nil {
			aps.logger.Info("Shutting down current container",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("performerID", currentContainer.performerID),
				zap.String("containerID", currentContainer.info.ID),
			)
			// Wait for running tasks to complete
			aps.logger.Info("Waiting for running tasks to complete",
				zap.String("performerID", currentContainer.performerID),
			)
			aps.waitForTaskCompletion(currentContainer.performerID)
			// Then shutdown container
			if err := aps.shutdownContainer(context.Background(), aps.config.AvsAddress, currentContainer); err != nil {
				errs = append(errs, fmt.Errorf("failed to shutdown current container: %w", err))
			}
			aps.cleanupTaskWaitGroup(currentContainer.performerID)
			aps.currentContainer.Store((*PerformerContainer)(nil))
		}
	}

	if aps.nextContainer != nil {
		aps.logger.Info("Shutting down next container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", aps.nextContainer.performerID),
			zap.String("containerID", aps.nextContainer.info.ID),
		)
		aps.waitForTaskCompletion(aps.nextContainer.performerID)
		if err := aps.shutdownContainer(context.Background(), aps.config.AvsAddress, aps.nextContainer); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown next container: %w", err))
		}
		aps.cleanupTaskWaitGroup(aps.nextContainer.performerID)
		aps.nextContainer = nil
	}

	if len(errs) > 0 {
		return stderrors.Join(errs...)
	}

	aps.logger.Info("AVS performer shutdown completed",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	return nil
}

// ListPerformers returns information about the current and next containers for this AVS performer
func (aps *AvsContainerPerformer) ListPerformers() []avsPerformer.PerformerMetadata {
	aps.performerContainersMu.Lock()
	defer aps.performerContainersMu.Unlock()

	var performers []avsPerformer.PerformerMetadata

	// Add current container info if exists
	if current := aps.currentContainer.Load(); current != nil {
		currentContainer, ok := current.(*PerformerContainer)
		if !ok || currentContainer == nil || currentContainer.info == nil {
			aps.logger.Error("Invalid type in currentContainer atomic.Value during listing")
			return performers
		}
		performers = append(performers, convertPerformerContainer(aps.config.AvsAddress, currentContainer))
	}

	// Add next container info if exists
	if aps.nextContainer != nil {
		performers = append(performers, convertPerformerContainer(aps.config.AvsAddress, aps.nextContainer))
	}

	return performers
}

// getOrCreateTaskWaitGroup returns the WaitGroup for a performer, creating it if needed
func (aps *AvsContainerPerformer) getOrCreateTaskWaitGroup(performerID string) *sync.WaitGroup {
	aps.taskWaitGroupsMu.Lock()
	defer aps.taskWaitGroupsMu.Unlock()

	wg, exists := aps.taskWaitGroups[performerID]
	if !exists {
		wg = &sync.WaitGroup{}
		aps.taskWaitGroups[performerID] = wg
	}
	return wg
}

// waitForTaskCompletion waits for all tasks on a performer to complete
func (aps *AvsContainerPerformer) waitForTaskCompletion(performerID string) {
	aps.taskWaitGroupsMu.Lock()
	wg, exists := aps.taskWaitGroups[performerID]
	aps.taskWaitGroupsMu.Unlock()

	if exists && wg != nil {
		wg.Wait()
	}
}

// cleanupTaskWaitGroup removes the WaitGroup for a performer
func (aps *AvsContainerPerformer) cleanupTaskWaitGroup(performerID string) {
	aps.taskWaitGroupsMu.Lock()
	defer aps.taskWaitGroupsMu.Unlock()
	delete(aps.taskWaitGroups, performerID)
}

// startDrainAndRemove initiates draining of a performer container in a separate goroutine
func (aps *AvsContainerPerformer) startDrainAndRemove(container *PerformerContainer) {
	performerID := container.performerID

	// Check if already draining
	aps.drainingPerformersMu.Lock()
	if _, exists := aps.drainingPerformers[performerID]; exists {
		aps.drainingPerformersMu.Unlock()
		aps.logger.Warn("Performer is already draining",
			zap.String("performerID", performerID),
		)
		return
	}
	aps.drainingPerformers[performerID] = struct{}{}
	aps.drainingPerformersMu.Unlock()

	go func() {
		aps.logger.Info("Starting performer drain",
			zap.String("performerID", performerID),
			zap.String("containerID", container.info.ID),
		)

		// Wait for all tasks to complete
		aps.waitForTaskCompletion(performerID)

		aps.logger.Info("Performer drained, removing container",
			zap.String("performerID", performerID),
			zap.String("containerID", container.info.ID),
		)

		// Stop liveness monitoring
		aps.containerManager.StopLivenessMonitoring(container.info.ID)

		// Stop and remove container
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := aps.containerManager.Stop(ctx, container.info.ID, 10*time.Second); err != nil {
			aps.logger.Warn("Failed to stop drained container",
				zap.String("performerID", performerID),
				zap.String("containerID", container.info.ID),
				zap.Error(err),
			)
		}
		if err := aps.containerManager.Remove(ctx, container.info.ID, true); err != nil {
			aps.logger.Warn("Failed to remove drained container",
				zap.String("performerID", performerID),
				zap.String("containerID", container.info.ID),
				zap.Error(err),
			)
		}
		cancel()

		// Remove from draining set and clean up WaitGroup
		aps.drainingPerformersMu.Lock()
		delete(aps.drainingPerformers, performerID)
		aps.drainingPerformersMu.Unlock()

		aps.cleanupTaskWaitGroup(performerID)

		aps.logger.Info("Performer drain completed",
			zap.String("performerID", performerID),
		)
	}()
}

// performPeriodicApplicationHealthChecks checks application health for containers that are Docker-healthy
func (aps *AvsContainerPerformer) performPeriodicApplicationHealthChecks(ctx context.Context) {
	aps.performerContainersMu.Lock()
	defer aps.performerContainersMu.Unlock()

	// Create a slice of containers to check
	var containersToCheck []*PerformerContainer

	if current := aps.currentContainer.Load(); current != nil {
		currentContainer := current.(*PerformerContainer)
		if currentContainer.performerHealth.ContainerIsHealthy && currentContainer.client != nil {
			containersToCheck = append(containersToCheck, currentContainer)
		}
	}

	if aps.nextContainer != nil && aps.nextContainer.performerHealth.ContainerIsHealthy && aps.nextContainer.client != nil {
		containersToCheck = append(containersToCheck, aps.nextContainer)
	}

	// Process all containers with the same logic
	for _, container := range containersToCheck {
		containerID := container.info.ID

		// Perform the application health check
		err := aps.checkApplicationHealth(ctx, container)
		container.performerHealth.LastHealthCheck = time.Now()

		if err != nil {
			// Health check failed
			container.performerHealth.ApplicationIsHealthy = false
			container.performerHealth.ConsecutiveApplicationHealthFailures++

			aps.logger.Warn("Application health check failed",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("containerID", containerID),
				zap.Error(err),
				zap.Int("consecutiveFailures", container.performerHealth.ConsecutiveApplicationHealthFailures),
			)

			// Handle consecutive failures
			if container.performerHealth.ConsecutiveApplicationHealthFailures >= maxConsecutiveApplicationHealthFailures {
				consecutiveFailures := container.performerHealth.ConsecutiveApplicationHealthFailures
				container.performerHealth.ConsecutiveApplicationHealthFailures = 0

				aps.logger.Error("Container application health failed multiple times, triggering restart",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("containerID", containerID),
					zap.Int("consecutiveFailures", consecutiveFailures),
				)

				// Send unhealthy status event
				if container.statusChan != nil {
					select {
					case container.statusChan <- avsPerformer.PerformerStatusEvent{
						Status:      avsPerformer.PerformerUnhealthy,
						PerformerID: container.performerID,
						Message:     fmt.Sprintf("Container unhealthy after %d consecutive health check failures", consecutiveFailures),
						Timestamp:   time.Now(),
					}:
						aps.logger.Info("Sent unhealthy status event",
							zap.String("avsAddress", aps.config.AvsAddress),
							zap.String("performerID", container.performerID),
						)
					default:
					}
				}

				if restartErr := aps.TriggerContainerRestart(container, fmt.Sprintf("application health check failed %d consecutive times", consecutiveFailures)); restartErr != nil {
					aps.logger.Error("Failed to trigger container restart",
						zap.String("avsAddress", aps.config.AvsAddress),
						zap.String("containerID", containerID),
						zap.Error(restartErr),
					)
				}
			}
		} else {
			// Health check succeeded
			container.performerHealth.ApplicationIsHealthy = true
			container.performerHealth.ConsecutiveApplicationHealthFailures = 0

			if container.statusChan != nil {
				select {
				case container.statusChan <- avsPerformer.PerformerStatusEvent{
					Status:      avsPerformer.PerformerHealthy,
					PerformerID: container.performerID,
					Message:     "Container is healthy and ready",
					Timestamp:   time.Now(),
				}:
					aps.logger.Info("Sent healthy status event",
						zap.String("avsAddress", aps.config.AvsAddress),
						zap.String("performerID", container.performerID),
					)
				default:
				}
			}
		}
	}
}

// Deploy performs a synchronous deployment - creates container, waits for health, and promotes
func (aps *AvsContainerPerformer) Deploy(ctx context.Context, image avsPerformer.PerformerImage) (*avsPerformer.DeploymentResult, error) {
	// Use deployment mutex to prevent concurrent deployments
	if !aps.activeDeploymentMu.TryLock() {
		return nil, fmt.Errorf("deployment in progress for avs %s", aps.config.AvsAddress)
	}
	defer aps.activeDeploymentMu.Unlock()

	// Generate deployment ID
	deploymentID := fmt.Sprintf("deployment-%s-%s", aps.config.AvsAddress, uuid.New().String())

	// Use default deployment timeout
	timeout := defaultDeploymentTimeout

	deploymentCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := &avsPerformer.DeploymentResult{
		ID:        deploymentID,
		Status:    avsPerformer.DeploymentStatusPending,
		Image:     image,
		StartTime: time.Now(),
	}

	aps.logger.Info("Starting deployment",
		zap.String("deploymentID", deploymentID),
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("repository", image.Repository),
		zap.String("tag", image.Tag),
	)

	// Create the performer
	creationResult, err := aps.CreatePerformer(deploymentCtx, image)
	if err != nil {
		result.Status = avsPerformer.DeploymentStatusFailed
		result.EndTime = time.Now()
		result.Error = err
		result.Message = fmt.Sprintf("Failed to create performer: %v", err)
		return result, err
	}

	result.PerformerID = creationResult.PerformerID
	result.Status = avsPerformer.DeploymentStatusInProgress

	// Monitor deployment until healthy
	healthyCtx, healthyCancel := context.WithTimeout(deploymentCtx, timeout)
	defer healthyCancel()

	if err := aps.waitForHealthy(healthyCtx, creationResult.StatusChan); err != nil {
		// Deployment failed, clean up
		aps.logger.Error("Deployment health check failed",
			zap.String("deploymentID", deploymentID),
			zap.Error(err),
		)

		// Clean up the failed deployment
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()

		if removeErr := aps.RemovePerformer(cleanupCtx, creationResult.PerformerID); removeErr != nil {
			aps.logger.Error("Failed to clean up failed deployment",
				zap.String("deploymentID", deploymentID),
				zap.Error(removeErr),
			)
		}

		result.Status = avsPerformer.DeploymentStatusFailed
		result.EndTime = time.Now()
		result.Error = err
		result.Message = fmt.Sprintf("Deployment failed: %v", err)
		return result, err
	}

	// Promote the performer
	if err := aps.PromotePerformer(deploymentCtx, creationResult.PerformerID); err != nil {
		aps.logger.Error("Failed to promote performer",
			zap.String("deploymentID", deploymentID),
			zap.Error(err),
		)

		// Clean up the failed deployment
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()

		if removeErr := aps.RemovePerformer(cleanupCtx, creationResult.PerformerID); removeErr != nil {
			aps.logger.Error("Failed to clean up after promotion failure",
				zap.String("deploymentID", deploymentID),
				zap.Error(removeErr),
			)
		}

		result.Status = avsPerformer.DeploymentStatusFailed
		result.EndTime = time.Now()
		result.Error = err
		result.Message = fmt.Sprintf("Failed to promote performer: %v", err)
		return result, err
	}

	// Deployment successful
	result.Status = avsPerformer.DeploymentStatusCompleted
	result.EndTime = time.Now()
	result.Message = "Deployment completed successfully"

	aps.logger.Info("Deployment completed successfully",
		zap.String("deploymentID", deploymentID),
		zap.String("performerID", creationResult.PerformerID),
		zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
	)

	return result, nil
}

// waitForHealthy waits for the performer to become healthy by monitoring the status channel
func (aps *AvsContainerPerformer) waitForHealthy(ctx context.Context, statusChan <-chan avsPerformer.PerformerStatusEvent) error {
	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("deployment timeout")
			}
			return ctx.Err()

		case status, ok := <-statusChan:
			if !ok {
				return fmt.Errorf("status channel closed unexpectedly")
			}

			switch status.Status {
			case avsPerformer.PerformerHealthy:
				aps.logger.Info("Performer is healthy",
					zap.String("performerID", status.PerformerID),
				)
				return nil

			case avsPerformer.PerformerUnhealthy:
				aps.logger.Warn("Performer is unhealthy, continuing to monitor",
					zap.String("performerID", status.PerformerID),
					zap.String("message", status.Message),
				)
			case avsPerformer.PerformerHealthUnknown:
				aps.logger.Warn("Performer health is unknown, continuing to monitor")
			}
		}
	}
}

func convertPerformerContainer(avsAddress string, container *PerformerContainer) avsPerformer.PerformerMetadata {
	return avsPerformer.PerformerMetadata{
		PerformerID:        container.performerID,
		AvsAddress:         avsAddress,
		Status:             container.status,
		ArtifactRegistry:   container.image.Repository,
		ArtifactTag:        container.image.Tag,
		ArtifactDigest:     container.image.Digest,
		ContainerHealthy:   container.performerHealth.ContainerIsHealthy,
		ApplicationHealthy: container.performerHealth.ApplicationIsHealthy,
		LastHealthCheck:    container.performerHealth.LastHealthCheck,
		ResourceID:         container.info.ID,
	}
}
