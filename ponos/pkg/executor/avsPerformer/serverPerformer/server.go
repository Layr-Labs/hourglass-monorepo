package serverPerformer

import (
	"context"
	stderrors "errors"
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
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	containerPort                           = 8080
	maxConsecutiveApplicationHealthFailures = 3
	defaultApplicationHealthCheckInterval   = 15 * time.Second
)

// PerformerContainer holds all information about a container
type PerformerContainer struct {
	PerformerID     string
	Info            *containerManager.ContainerInfo
	Client          performerV1.PerformerServiceClient
	EventChan       <-chan containerManager.ContainerEvent
	PerformerHealth *avsPerformer.PerformerHealth
	StatusChan      chan<- avsPerformer.PerformerStatusEvent
	Image           avsPerformer.PerformerImage
	Status          avsPerformer.PerformerContainerStatus
}

type AvsPerformerServer struct {
	config         *avsPerformer.AvsPerformerConfig
	logger         *zap.Logger
	peeringFetcher peering.IPeeringDataFetcher

	containerManager containerManager.ContainerManager
	currentContainer *PerformerContainer
	nextContainer    *PerformerContainer
	deploymentMu     sync.Mutex

	aggregatorPeers                []*peering.OperatorPeerInfo
	applicationHealthCheckInterval time.Duration
}

// NewAvsPerformerServer creates a new AvsPerformerServer with the provided container manager
func NewAvsPerformerServer(
	config *avsPerformer.AvsPerformerConfig,
	peeringFetcher peering.IPeeringDataFetcher,
	logger *zap.Logger,
	containerMgr containerManager.ContainerManager,
) *AvsPerformerServer {
	return NewAvsPerformerServerWithHealthCheckInterval(
		config,
		peeringFetcher,
		logger,
		containerMgr,
		defaultApplicationHealthCheckInterval,
	)
}

// NewAvsPerformerServerWithHealthCheckInterval creates a new AvsPerformerServer with custom health check interval
func NewAvsPerformerServerWithHealthCheckInterval(
	config *avsPerformer.AvsPerformerConfig,
	peeringFetcher peering.IPeeringDataFetcher,
	logger *zap.Logger,
	containerMgr containerManager.ContainerManager,
	applicationHealthCheckInterval time.Duration,
) *AvsPerformerServer {
	return &AvsPerformerServer{
		config:                         config,
		logger:                         logger,
		peeringFetcher:                 peeringFetcher,
		containerManager:               containerMgr,
		applicationHealthCheckInterval: applicationHealthCheckInterval,
	}
}

func (aps *AvsPerformerServer) fetchAggregatorPeerInfo(ctx context.Context) ([]*peering.OperatorPeerInfo, error) {
	retries := []uint64{1, 3, 5, 10, 20}
	for i, retry := range retries {
		aggPeers, err := aps.peeringFetcher.ListAggregatorOperators(ctx, aps.config.AvsAddress)
		if err != nil {
			aps.logger.Sugar().Errorw("Failed to fetch aggregator peers",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.Error(err),
			)
			if i == len(retries)-1 {
				aps.logger.Sugar().Infow("Giving up on fetching aggregator peers",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.Error(err),
				)
				return nil, err
			}
			time.Sleep(time.Duration(retry) * time.Second)
			continue
		}
		return aggPeers, nil
	}
	return nil, fmt.Errorf("failed to fetch aggregator peers after retries")
}

// generatePerformerID generates a unique performer ID
func (aps *AvsPerformerServer) generatePerformerID() string {
	return fmt.Sprintf("performer-%s-%s", aps.config.AvsAddress, uuid.New().String())
}

// createAndStartContainer creates, starts, and prepares a container for the AVS performer
func (aps *AvsPerformerServer) createAndStartContainer(
	ctx context.Context,
	avsAddress string,
	containerConfig *containerManager.ContainerConfig,
) (*PerformerContainer, error) {

	// Create the container
	containerInfo, err := aps.containerManager.Create(ctx, containerConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	// Start the container
	if err := aps.containerManager.Start(ctx, containerInfo.ID); err != nil {
		// Clean up on failure
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to start container")
	}

	// Wait for the container to be running
	if err := aps.containerManager.WaitForRunning(ctx, containerInfo.ID, 30*time.Second); err != nil {
		// Clean up on failure
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to wait for container to be running")
	}

	// Get updated container information with port mappings
	updatedInfo, err := aps.containerManager.Inspect(ctx, containerInfo.ID)
	if err != nil {
		// Clean up on failure
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to inspect container")
	}

	// Get the container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, containerPort, containerConfig.NetworkName)
	if err != nil {
		// Clean up on failure
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
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
		if removeErr := aps.containerManager.Remove(ctx, updatedInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", updatedInfo.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to create performer client")
	}

	// Start liveness monitoring for this container
	livenessConfig := containerManager.NewDefaultAvsPerformerLivenessConfig()
	eventChan, err := aps.containerManager.StartLivenessMonitoring(ctx, updatedInfo.ID, livenessConfig)
	if err != nil {
		// Clean up on failure
		if removeErr := aps.containerManager.Remove(ctx, updatedInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove container during monitoring setup failure",
				zap.String("containerID", updatedInfo.ID),
				zap.Error(removeErr),
			)
		}
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
	container := &PerformerContainer{
		PerformerID: performerID,
		Info:        updatedInfo,
		Client:      perfClient,
		EventChan:   eventChan,
		PerformerHealth: &avsPerformer.PerformerHealth{
			ContainerHealth: true,
			LastHealthCheck: time.Now(),
		},
	}

	return container, nil
}

func (aps *AvsPerformerServer) Initialize(ctx context.Context) error {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()
	// Fetch aggregator peer information
	aggregatorPeers, err := aps.fetchAggregatorPeerInfo(ctx)
	if err != nil {
		return err
	}
	aps.aggregatorPeers = aggregatorPeers
	aps.logger.Sugar().Infow("Fetched aggregator peers",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.Any("aggregatorPeers", aps.aggregatorPeers),
	)

	// Check if we should start with a container loaded
	// Skip container creation if image info is empty (for deployment-based initialization)
	if aps.config.Image.Repository == "" || aps.config.Image.Tag == "" {
		aps.logger.Info("Starting PerformerServer without initial container.",
			zap.String("avsAddress", aps.config.AvsAddress),
		)
		return nil
	}

	// Create and start container
	containerInstance, err := aps.createAndStartContainer(
		ctx,
		aps.config.AvsAddress,
		containerManager.CreateDefaultContainerConfig(
			aps.config.AvsAddress,
			aps.config.Image.Repository,
			aps.config.Image.Tag,
			containerPort,
			aps.config.PerformerNetworkName,
		),
	)
	if err != nil {
		return err
	}

	// Store the image info in the container
	containerInstance.Image = aps.config.Image
	aps.currentContainer = containerInstance
	aps.currentContainer.Status = avsPerformer.PerformerContainerStatusInService

	// Start monitoring events for the new container
	go aps.monitorContainerEvents(ctx, aps.currentContainer)

	return nil
}

// monitorContainerEvents monitors container lifecycle events and performs periodic application health checks
func (aps *AvsPerformerServer) monitorContainerEvents(ctx context.Context, container *PerformerContainer) {
	// Create a ticker for periodic application health checks
	appHealthCheckTicker := time.NewTicker(aps.applicationHealthCheckInterval)
	defer appHealthCheckTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-container.EventChan:
			if !ok {
				aps.logger.Info("Container event channel closed",
					zap.String("performerID", container.PerformerID),
					zap.String("containerID", container.Info.ID),
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
func (aps *AvsPerformerServer) handleContainerEvent(ctx context.Context, event containerManager.ContainerEvent, targetContainer *PerformerContainer) {
	aps.logger.Info("Container event received",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", targetContainer.PerformerID),
		zap.String("containerID", event.ContainerID),
		zap.String("eventType", string(event.Type)),
		zap.String("message", event.Message),
		zap.Int("restartCount", event.State.RestartCount),
	)

	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	switch event.Type {
	case containerManager.EventStarted:
		aps.logger.Info("Container started successfully",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
		)
		targetContainer.PerformerHealth.ContainerHealth = true
		return

	case containerManager.EventHealthy:
		aps.logger.Debug("Container Docker healthy signal received",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
		)
		// Update the container health status
		targetContainer.PerformerHealth.ContainerHealth = true
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
		targetContainer.PerformerHealth.ContainerHealth = false
		targetContainer.PerformerHealth.ApplicationHealth = false

	case containerManager.EventOOMKilled:
		aps.logger.Error("Container killed due to OOM",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.Int("restartCount", event.State.RestartCount),
		)
		// Auto-restart is handled by containerManager
		targetContainer.PerformerHealth.ContainerHealth = false
		targetContainer.PerformerHealth.ApplicationHealth = false

	case containerManager.EventRestarted:
		aps.logger.Info("Container restarted successfully",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.Int("restartCount", event.State.RestartCount),
		)
		// Reset health context since container was restarted
		targetContainer.PerformerHealth.ContainerHealth = false
		targetContainer.PerformerHealth.ApplicationHealth = false
		targetContainer.PerformerHealth.ConsecutiveApplicationHealthFailures = 0
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
		targetContainer.PerformerHealth.ContainerHealth = false
		targetContainer.PerformerHealth.ApplicationHealth = false

		// Handle restart failures differently for current vs next container
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
		targetContainer.PerformerHealth.ContainerHealth = false
		targetContainer.PerformerHealth.ApplicationHealth = false

	case containerManager.EventRestarting:
		aps.logger.Info("Container is being restarted",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.String("reason", event.Message),
		)
		// Container is restarting, mark as unhealthy
		targetContainer.PerformerHealth.ContainerHealth = false
		targetContainer.PerformerHealth.ApplicationHealth = false
	}

	// Only reach this point if the event indicates an unhealthy container.
	if targetContainer.StatusChan != nil {
		select {
		case targetContainer.StatusChan <- avsPerformer.PerformerStatusEvent{
			Status:      avsPerformer.PerformerUnhealthy,
			PerformerID: targetContainer.PerformerID,
			Message:     "Container is unhealthy",
			Timestamp:   time.Now(),
		}:
			aps.logger.Info("Sent unhealthy status event",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("performerID", targetContainer.PerformerID),
			)
		default:
		}
	}
}

// recreateContainer recreates a container that was killed/removed by updating the fields in place
func (aps *AvsPerformerServer) recreateContainer(ctx context.Context, targetContainer *PerformerContainer) {
	// Stop monitoring the old container
	aps.containerManager.StopLivenessMonitoring(targetContainer.Info.ID)

	aps.logger.Info("Starting container recreation",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", targetContainer.PerformerID),
		zap.String("previousContainerID", targetContainer.Info.ID),
		zap.String("image", fmt.Sprintf("%s:%s", targetContainer.Image.Repository, targetContainer.Image.Tag)),
	)

	// Create and start new container
	newContainer, err := aps.createAndStartContainer(
		ctx,
		aps.config.AvsAddress,
		containerManager.CreateDefaultContainerConfig(
			aps.config.AvsAddress,
			targetContainer.Image.Repository,
			targetContainer.Image.Tag,
			containerPort,
			aps.config.PerformerNetworkName,
		),
	)
	if err != nil {
		aps.logger.Error("Failed to recreate container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", targetContainer.PerformerID),
			zap.Error(err),
		)
		return
	}

	// Update the fields in the existing PerformerContainer reference
	// This keeps the same PerformerContainer object but with new container details
	targetContainer.Info = newContainer.Info
	targetContainer.Client = newContainer.Client
	targetContainer.EventChan = newContainer.EventChan
	targetContainer.PerformerHealth = newContainer.PerformerHealth

	aps.logger.Info("Container recreation completed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", targetContainer.PerformerID),
		zap.String("newContainerID", targetContainer.Info.ID),
	)
}

// recreatePerformerClientForContainer recreates the performer client connection after container restart
func (aps *AvsPerformerServer) recreatePerformerClientForContainer(ctx context.Context, container *PerformerContainer) {
	// Get updated container information
	updatedInfo, err := aps.containerManager.Inspect(ctx, container.Info.ID)
	if err != nil {
		aps.logger.Error("Failed to inspect container after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
		return
	}
	container.Info = updatedInfo

	// Get the new container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, containerPort, aps.config.PerformerNetworkName)
	if err != nil {
		aps.logger.Error("Failed to get container endpoint after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
		return
	}

	// Create new performer client
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(endpoint, true)
	if err != nil {
		aps.logger.Error("Failed to recreate performer client after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
		return
	}

	container.Info = updatedInfo
	container.Client = perfClient

	aps.logger.Info("Performer client recreated successfully after container restart",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", container.Info.ID),
		zap.String("endpoint", endpoint),
	)
}

// TriggerContainerRestart allows the application to manually trigger a restart for a specific container
func (aps *AvsPerformerServer) TriggerContainerRestart(container *PerformerContainer, reason string) error {
	if container == nil || container.Info == nil {
		return fmt.Errorf("container info not available")
	}

	containerID := container.Info.ID

	aps.logger.Info("Triggering manual container restart",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", containerID),
		zap.String("reason", reason),
	)

	return aps.containerManager.TriggerRestart(containerID, reason)
}

// checkApplicationHealth performs a single health check on the specified container
func (aps *AvsPerformerServer) checkApplicationHealth(ctx context.Context, container *PerformerContainer) error {
	healthCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	res, err := container.Client.HealthCheck(healthCtx, &performerV1.HealthCheckRequest{})
	if err != nil {
		return err
	}

	aps.logger.Debug("Application health check successful",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", container.Info.ID),
		zap.String("status", res.Status.String()),
	)

	return nil
}

// CreatePerformer creates a new performer and returns the creation result
// Always deploys to the nextContainer slot
func (aps *AvsPerformerServer) CreatePerformer(
	ctx context.Context,
	image avsPerformer.PerformerImage,
) (*avsPerformer.PerformerCreationResult, error) {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	aps.logger.Info("Starting performer deployment",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("imageRepository", image.Repository),
		zap.String("imageTag", image.Tag),
	)

	// Check if next container slot is already occupied
	if aps.nextContainer != nil {
		aps.logger.Error("Cannot create new performer, next container slot is already occupied",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("existingNextPerformerID", aps.nextContainer.PerformerID),
			zap.String("requestedImage", fmt.Sprintf("%s:%s", image.Repository, image.Tag)),
		)
		return nil, fmt.Errorf("a next performer already exists (ID: %s). Please remove it explicitly before creating a new one", aps.nextContainer.PerformerID)
	}

	// Create the new container instance
	newContainer, err := aps.createAndStartContainer(
		ctx,
		aps.config.AvsAddress,
		containerManager.CreateDefaultContainerConfig(
			aps.config.AvsAddress,
			image.Repository,
			image.Tag,
			containerPort,
			aps.config.PerformerNetworkName,
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	// Store the image info in the container
	newContainer.Image = image
	// Create status channel for deployment monitoring
	statusChan := make(chan avsPerformer.PerformerStatusEvent, 10)
	newContainer.StatusChan = statusChan

	// Always deploy as next container
	aps.nextContainer = newContainer
	aps.nextContainer.Status = avsPerformer.PerformerContainerStatusStaged

	aps.logger.Info("Deployed as next performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", newContainer.PerformerID),
		zap.String("containerID", newContainer.Info.ID),
	)

	// Start monitoring events for the new container
	go aps.monitorContainerEvents(ctx, newContainer)

	// Get endpoint for logging
	endpoint, _ := containerManager.GetContainerEndpoint(newContainer.Info, containerPort, aps.config.PerformerNetworkName)

	aps.logger.Info("Performer deployment started",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", newContainer.PerformerID),
		zap.String("containerID", newContainer.Info.ID),
		zap.String("endpoint", endpoint),
	)

	return &avsPerformer.PerformerCreationResult{
		PerformerID: newContainer.PerformerID,
		StatusChan:  statusChan,
	}, nil
}

// RemovePerformer removes a performer from the server by its performerID.
func (aps *AvsPerformerServer) RemovePerformer(ctx context.Context, performerID string) error {
	aps.deploymentMu.Lock()

	// Determine which container to remove
	var targetContainer *PerformerContainer
	var containerType string

	if aps.currentContainer != nil && aps.currentContainer.PerformerID == performerID {
		defer func() {
			aps.currentContainer = nil
			aps.deploymentMu.Unlock()
		}()
		targetContainer = aps.currentContainer
		containerType = "current"
	} else if aps.nextContainer != nil && aps.nextContainer.PerformerID == performerID {
		defer func() {
			aps.nextContainer = nil
			aps.deploymentMu.Unlock()
		}()
		targetContainer = aps.nextContainer
		containerType = "next"
	} else {
		defer aps.deploymentMu.Unlock()
		// Performer not found
		aps.logger.Error("Performer not found",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", performerID),
		)
		return fmt.Errorf("performer with ID %s not found", performerID)
	}

	// Log the removal
	aps.logger.Info("Removing performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", performerID),
		zap.String("containerID", targetContainer.Info.ID),
		zap.String("containerType", containerType),
	)

	// Close the status channel if it exists
	if targetContainer.StatusChan != nil {
		close(targetContainer.StatusChan)
	}

	// Shutdown the container (this will stop monitoring and remove it)
	if err := aps.shutdownContainer(ctx, aps.config.AvsAddress, targetContainer); err != nil {
		aps.logger.Error("Failed to shutdown container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", performerID),
			zap.String("containerID", targetContainer.Info.ID),
			zap.String("containerType", containerType),
			zap.Error(err),
		)
		return fmt.Errorf("failed to remove %s performer: %w", containerType, err)
	}

	aps.logger.Info("Performer removed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", performerID),
		zap.String("containerType", containerType),
	)

	return nil
}

// PromotePerformer promotes the specified performer to currentContainer
// If the performer is already current, it's a no-op success
// If the performer is next and healthy, it's promoted to current
// If the performer is not found or unhealthy, an error is returned
func (aps *AvsPerformerServer) PromotePerformer(ctx context.Context, performerID string) error {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	// Check if the performer is already the current container
	if aps.currentContainer != nil && aps.currentContainer.PerformerID == performerID {
		aps.logger.Info("Performer is already the current container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", performerID),
		)
		return nil
	}

	// Check if the performer is the next container
	if aps.nextContainer == nil || aps.nextContainer.PerformerID != performerID {
		return fmt.Errorf("performer %s is not in the next deployment slot", performerID)
	}

	// Verify nextContainer is healthy
	if !aps.nextContainer.PerformerHealth.ContainerHealth || !aps.nextContainer.PerformerHealth.ApplicationHealth {
		return fmt.Errorf("cannot promote unhealthy performer %s (container health: %v, application health: %v)",
			performerID,
			aps.nextContainer.PerformerHealth.ContainerHealth,
			aps.nextContainer.PerformerHealth.ApplicationHealth)
	}

	aps.logger.Info("Promoting performer to current",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", performerID),
		zap.String("containerID", aps.nextContainer.Info.ID),
	)

	// Stop the old container gracefully
	oldContainer := aps.currentContainer
	if oldContainer != nil && oldContainer.Info != nil {
		// Stop liveness monitoring
		aps.containerManager.StopLivenessMonitoring(oldContainer.Info.ID)

		// Stop and remove old container
		if err := aps.containerManager.Stop(ctx, oldContainer.Info.ID, 10*time.Second); err != nil {
			aps.logger.Warn("Failed to stop old container",
				zap.String("performerID", oldContainer.PerformerID),
				zap.String("containerID", oldContainer.Info.ID),
				zap.Error(err),
			)
		}
		if err := aps.containerManager.Remove(ctx, oldContainer.Info.ID, true); err != nil {
			aps.logger.Warn("Failed to remove old container",
				zap.String("performerID", oldContainer.PerformerID),
				zap.String("containerID", oldContainer.Info.ID),
				zap.Error(err),
			)
		}
	}

	// Promote next to current and update status
	aps.currentContainer = aps.nextContainer
	aps.currentContainer.Status = avsPerformer.PerformerContainerStatusInService
	aps.nextContainer = nil

	aps.logger.Info("Performer promotion completed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("performerID", performerID),
	)

	return nil
}

func (aps *AvsPerformerServer) ValidateTaskSignature(t *performerTask.PerformerTask) error {
	sig, err := bn254.NewSignatureFromBytes(t.Signature)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to create signature from bytes",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}
	peer := util.Find(aps.aggregatorPeers, func(p *peering.OperatorPeerInfo) bool {
		return strings.EqualFold(p.OperatorAddress, t.AggregatorAddress)
	})
	if peer == nil {
		aps.logger.Sugar().Errorw("Failed to find peer for task",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("aggregatorAddress", t.AggregatorAddress),
		)
		return fmt.Errorf("failed to find peer for task")
	}

	isVerified := false

	// TODO(seanmcgary): this should verify the key against the expected aggregator operatorSetID
	for _, opset := range peer.OperatorSets {
		verfied, err := sig.Verify(opset.PublicKey, t.Payload)
		if err != nil {
			aps.logger.Sugar().Errorw("Error verifying signature",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("aggregatorAddress", t.AggregatorAddress),
				zap.Error(err),
			)
			continue
		}
		if !verfied {
			aps.logger.Sugar().Errorw("Failed to verify signature",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("aggregatorAddress", t.AggregatorAddress),
				zap.Error(err),
			)
			continue
		}
		aps.logger.Sugar().Infow("Signature verified with operator set",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("aggregatorAddress", t.AggregatorAddress),
			zap.Uint32("opsetID", opset.OperatorSetID),
		)
		isVerified = true
	}

	if !isVerified {
		aps.logger.Sugar().Errorw("Failed to verify signature with any operator set",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("aggregatorAddress", t.AggregatorAddress),
		)
		return fmt.Errorf("failed to verify signature with any operator set")
	}

	return nil
}

func (aps *AvsPerformerServer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	aps.logger.Sugar().Infow("Processing task", zap.Any("task", task))

	aps.deploymentMu.Lock()
	if aps.currentContainer == nil || aps.currentContainer.Client == nil {
		aps.deploymentMu.Unlock()
		return nil, fmt.Errorf("no current container available to execute task")
	}

	// Make the call while holding the lock, or copy the client reference
	client := aps.currentContainer.Client
	aps.deploymentMu.Unlock()

	// Use the local client reference
	res, err := client.ExecuteTask(ctx, &performerV1.TaskRequest{
		TaskId:  []byte(task.TaskID),
		Payload: task.Payload,
	})
	if err != nil {
		aps.logger.Sugar().Errorw("Performer failed to handle task",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return nil, err
	}

	return performerTask.NewTaskResultFromResultProto(res), nil
}

// shutdownContainer handles the shutdown of a single container instance
func (aps *AvsPerformerServer) shutdownContainer(ctx context.Context, avsAddress string, container *PerformerContainer) error {
	if container == nil || container.Info == nil || aps.containerManager == nil {
		return nil
	}

	aps.logger.Info("Shutting down container",
		zap.String("avsAddress", avsAddress),
		zap.String("performerID", container.PerformerID),
		zap.String("containerID", container.Info.ID),
	)

	// Stop liveness monitoring for this container
	aps.containerManager.StopLivenessMonitoring(container.Info.ID)

	// Stop the container
	if err := aps.containerManager.Stop(ctx, container.Info.ID, 10*time.Second); err != nil {
		aps.logger.Error("Failed to stop container",
			zap.String("avsAddress", avsAddress),
			zap.String("performerID", container.PerformerID),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
	}

	// Remove the container
	if err := aps.containerManager.Remove(ctx, container.Info.ID, true); err != nil {
		aps.logger.Error("Failed to remove container",
			zap.String("avsAddress", avsAddress),
			zap.String("performerID", container.PerformerID),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
		return err
	}

	aps.logger.Info("Container shutdown completed",
		zap.String("avsAddress", avsAddress),
		zap.String("performerID", container.PerformerID),
		zap.String("containerID", container.Info.ID),
	)

	return nil
}

func (aps *AvsPerformerServer) Shutdown() error {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	aps.logger.Info("Shutting down AVS performer",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	// Shutdown both current and next containers
	var errs []error

	if aps.currentContainer != nil {
		aps.logger.Info("Shutting down current container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", aps.currentContainer.PerformerID),
			zap.String("containerID", aps.currentContainer.Info.ID),
		)
		if err := aps.shutdownContainer(context.Background(), aps.config.AvsAddress, aps.currentContainer); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown current container: %w", err))
		}
		aps.currentContainer = nil
	}

	if aps.nextContainer != nil {
		aps.logger.Info("Shutting down next container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("performerID", aps.nextContainer.PerformerID),
			zap.String("containerID", aps.nextContainer.Info.ID),
		)
		if err := aps.shutdownContainer(context.Background(), aps.config.AvsAddress, aps.nextContainer); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown next container: %w", err))
		}
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
func (aps *AvsPerformerServer) ListPerformers() []avsPerformer.PerformerInfo {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	var performers []avsPerformer.PerformerInfo

	// Add current container info if exists
	if aps.currentContainer != nil {
		performers = append(performers, convertPerformerContainer(aps.config.AvsAddress, aps.currentContainer))
	}

	// Add next container info if exists
	if aps.nextContainer != nil {
		performers = append(performers, convertPerformerContainer(aps.config.AvsAddress, aps.nextContainer))
	}

	return performers
}

// performPeriodicApplicationHealthChecks checks application health for containers that are Docker-healthy
func (aps *AvsPerformerServer) performPeriodicApplicationHealthChecks(ctx context.Context) {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	// Create a slice of containers to check
	var containersToCheck []*PerformerContainer

	if aps.currentContainer != nil && aps.currentContainer.PerformerHealth.ContainerHealth && aps.currentContainer.Client != nil {
		containersToCheck = append(containersToCheck, aps.currentContainer)
	}

	if aps.nextContainer != nil && aps.nextContainer.PerformerHealth.ContainerHealth && aps.nextContainer.Client != nil {
		containersToCheck = append(containersToCheck, aps.nextContainer)
	}

	// Process all containers with the same logic
	for _, container := range containersToCheck {
		containerID := container.Info.ID

		// Perform the application health check
		err := aps.checkApplicationHealth(ctx, container)
		container.PerformerHealth.LastHealthCheck = time.Now()

		if err != nil {
			// Health check failed
			container.PerformerHealth.ApplicationHealth = false
			container.PerformerHealth.ConsecutiveApplicationHealthFailures++

			aps.logger.Warn("Application health check failed",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("containerID", containerID),
				zap.Error(err),
				zap.Int("consecutiveFailures", container.PerformerHealth.ConsecutiveApplicationHealthFailures),
			)

			// Handle consecutive failures
			if container.PerformerHealth.ConsecutiveApplicationHealthFailures >= maxConsecutiveApplicationHealthFailures {
				consecutiveFailures := container.PerformerHealth.ConsecutiveApplicationHealthFailures
				container.PerformerHealth.ConsecutiveApplicationHealthFailures = 0

				aps.logger.Error("Container application health failed multiple times, triggering restart",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("containerID", containerID),
					zap.Int("consecutiveFailures", consecutiveFailures),
				)

				// Send unhealthy status event
				if container.StatusChan != nil {
					select {
					case container.StatusChan <- avsPerformer.PerformerStatusEvent{
						Status:      avsPerformer.PerformerUnhealthy,
						PerformerID: container.PerformerID,
						Message:     fmt.Sprintf("Container unhealthy after %d consecutive health check failures", consecutiveFailures),
						Timestamp:   time.Now(),
					}:
						aps.logger.Info("Sent unhealthy status event",
							zap.String("avsAddress", aps.config.AvsAddress),
							zap.String("performerID", container.PerformerID),
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
				return
			}
		} else {
			// Health check succeeded
			container.PerformerHealth.ApplicationHealth = true
			container.PerformerHealth.ConsecutiveApplicationHealthFailures = 0

			// Send healthy status event if this is the first time becoming healthy
			if container.StatusChan != nil {
				select {
				case container.StatusChan <- avsPerformer.PerformerStatusEvent{
					Status:      avsPerformer.PerformerHealthy,
					PerformerID: container.PerformerID,
					Message:     "Container is healthy and ready",
					Timestamp:   time.Now(),
				}:
					aps.logger.Info("Sent healthy status event",
						zap.String("avsAddress", aps.config.AvsAddress),
						zap.String("performerID", container.PerformerID),
					)
				default:
				}
			}
		}
	}
}

func convertPerformerContainer(avsAddress string, container *PerformerContainer) avsPerformer.PerformerInfo {
	return avsPerformer.PerformerInfo{
		PerformerID:        container.PerformerID,
		AvsAddress:         avsAddress,
		Status:             container.Status,
		ArtifactRegistry:   container.Image.Repository,
		ArtifactDigest:     container.Image.Tag,
		ContainerHealthy:   container.PerformerHealth.ContainerHealth,
		ApplicationHealthy: container.PerformerHealth.ApplicationHealth,
		LastHealthCheck:    container.PerformerHealth.LastHealthCheck,
		ContainerID:        container.Info.ID,
	}
}
