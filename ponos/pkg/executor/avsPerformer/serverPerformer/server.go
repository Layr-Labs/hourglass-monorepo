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
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	containerPort                           = 8080
	maxConsecutiveApplicationHealthFailures = 3
	defaultApplicationHealthCheckInterval   = 15 * time.Second
)

// PerformerHealth tracks the health state of a container
type PerformerHealth struct {
	ContainerHealth                      bool
	ApplicationHealth                    bool
	ConsecutiveApplicationHealthFailures int
	LastHealthCheck                      time.Time
}

// PerformerContainer holds all information about a container
type PerformerContainer struct {
	Info            *containerManager.ContainerInfo
	Client          performerV1.PerformerServiceClient
	EventChan       <-chan containerManager.ContainerEvent
	PerformerHealth *PerformerHealth
	StatusChan      chan<- avsPerformer.PerformerStatusEvent
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

	// Create the container instance with all components
	containerInstance := &PerformerContainer{
		Info:            updatedInfo,
		Client:          perfClient,
		EventChan:       eventChan,
		PerformerHealth: &PerformerHealth{},
	}

	aps.logger.Info("Container created and monitoring started",
		zap.String("avsAddress", avsAddress),
		zap.String("containerID", updatedInfo.ID),
		zap.String("endpoint", endpoint),
	)

	return containerInstance, nil
}

func (aps *AvsPerformerServer) Initialize(ctx context.Context) error {
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

	// Check if we should create an initial container
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
	aps.deploymentMu.Lock()
	aps.currentContainer = containerInstance
	aps.deploymentMu.Unlock()

	// Start monitoring events for the new container
	go aps.monitorContainerEvents(ctx, aps.currentContainer.EventChan)

	return nil
}

// monitorContainerEvents monitors container lifecycle events and performs periodic application health checks
func (aps *AvsPerformerServer) monitorContainerEvents(ctx context.Context, eventChan <-chan containerManager.ContainerEvent) {
	// Create a ticker for periodic application health checks
	appHealthCheckTicker := time.NewTicker(aps.applicationHealthCheckInterval)
	defer appHealthCheckTicker.Stop()

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
		case <-appHealthCheckTicker.C:
			// Perform periodic application health checks for containers that are Docker-healthy
			aps.performPeriodicApplicationHealthChecks(ctx)
		}
	}
}

// handleContainerEvent processes individual container events
func (aps *AvsPerformerServer) handleContainerEvent(ctx context.Context, event containerManager.ContainerEvent) {
	aps.logger.Info("Container event received",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", event.ContainerID),
		zap.String("eventType", string(event.Type)),
		zap.String("message", event.Message),
		zap.Int("restartCount", event.State.RestartCount),
	)

	// Determine which container this event is for and handle the event
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	var targetContainer *PerformerContainer

	if aps.currentContainer != nil && aps.currentContainer.Info != nil && aps.currentContainer.Info.ID == event.ContainerID {
		targetContainer = aps.currentContainer
	} else if aps.nextContainer != nil && aps.nextContainer.Info != nil && aps.nextContainer.Info.ID == event.ContainerID {
		targetContainer = aps.nextContainer
	} else {
		aps.logger.Warn("Container event received but not relevant container")
		return
	}

	if targetContainer == nil {
		aps.logger.Warn("Received event for unknown container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.String("eventType", string(event.Type)),
		)
		return
	}

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
		// Recreate performer client connection for the specific container that restarted
		go aps.recreatePerformerClientForContainer(ctx, targetContainer)

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
			go aps.recreateContainer(ctx)
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
	select {
	case targetContainer.StatusChan <- avsPerformer.PerformerStatusEvent{
		Status:      avsPerformer.PerformerUnhealthy,
		ContainerID: targetContainer.Info.ID,
		Message:     "Container is unhealthy",
		Timestamp:   time.Now(),
	}:
		aps.logger.Info("Sent unhealthy status event",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", targetContainer.Info.ID),
		)
	default:
	}
}

// recreateContainer recreates a container that was killed/removed
func (aps *AvsPerformerServer) recreateContainer(ctx context.Context) {
	aps.deploymentMu.Lock()
	prevContainerID := ""
	if aps.currentContainer != nil && aps.currentContainer.Info != nil {
		prevContainerID = aps.currentContainer.Info.ID
		aps.containerManager.StopLivenessMonitoring(aps.currentContainer.Info.ID)
	}
	aps.deploymentMu.Unlock()

	aps.logger.Info("Starting container recreation",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("previousContainerID", prevContainerID),
	)

	// Create and start new container
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
		aps.logger.Error("Failed to recreate container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return
	}

	aps.deploymentMu.Lock()
	aps.currentContainer = containerInstance
	aps.deploymentMu.Unlock()

	// Start monitoring events for the new container
	go aps.monitorContainerEvents(ctx, containerInstance.EventChan)

	aps.logger.Info("Container recreation completed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("newContainerID", containerInstance.Info.ID),
	)
}

// recreatePerformerClientForContainer recreates the performer client connection after container restart
func (aps *AvsPerformerServer) recreatePerformerClientForContainer(ctx context.Context, container *PerformerContainer) {
	// Wait a moment for the container to fully start
	time.Sleep(2 * time.Second)

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

	// Lock before modifying shared container fields
	aps.deploymentMu.Lock()
	container.Info = updatedInfo
	container.Client = perfClient
	aps.deploymentMu.Unlock()

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

// DeployContainer deploys a new container and returns a status channel
// If no currentContainer exists, it deploys to currentContainer slot
// Otherwise, it deploys to nextContainer slot (overwriting if necessary)
func (aps *AvsPerformerServer) DeployContainer(
	ctx context.Context,
	avsId string,
	image avsPerformer.PerformerImage,
) (<-chan avsPerformer.PerformerStatusEvent, error) {
	// Lock to prevent concurrent deployments
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	aps.logger.Info("Starting container deployment",
		zap.String("avsAddress", avsId),
		zap.String("imageRepository", image.Repository),
		zap.String("imageTag", image.Tag),
	)

	// Create the new container instance
	newContainer, err := aps.createAndStartContainer(
		ctx,
		avsId,
		containerManager.CreateDefaultContainerConfig(
			avsId,
			image.Repository,
			image.Tag,
			containerPort,
			aps.config.PerformerNetworkName,
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	// Create status channel for deployment monitoring
	statusChan := make(chan avsPerformer.PerformerStatusEvent, 10)
	newContainer.StatusChan = statusChan

	// Determine where to place the container
	if aps.currentContainer == nil {
		// No current container, this becomes the current container
		aps.currentContainer = newContainer
		aps.logger.Info("Deployed as current container",
			zap.String("avsAddress", avsId),
			zap.String("containerID", newContainer.Info.ID),
		)
	} else {
		// Current container exists, deploy as next container
		// If there's an existing nextContainer, clean it up first
		if aps.nextContainer != nil {
			aps.logger.Info("Cleaning up existing next container",
				zap.String("avsAddress", avsId),
				zap.String("containerID", aps.nextContainer.Info.ID),
			)

			// Close its status channel if it exists
			if aps.nextContainer.StatusChan != nil {
				close(aps.nextContainer.StatusChan)
			}

			// Schedule cleanup in background to avoid blocking
			containerToCleanup := aps.nextContainer
			go func() {
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := aps.shutdownContainer(cleanupCtx, avsId, containerToCleanup); err != nil {
					aps.logger.Error("Failed to cleanup previous next container",
						zap.String("avsAddress", avsId),
						zap.String("containerID", containerToCleanup.Info.ID),
						zap.Error(err),
					)
				}
			}()
		}

		aps.nextContainer = newContainer
		aps.logger.Info("Deployed as next container",
			zap.String("avsAddress", avsId),
			zap.String("containerID", newContainer.Info.ID),
		)
	}

	// Start monitoring events for the new container
	go aps.monitorContainerEvents(ctx, newContainer.EventChan)

	// Get endpoint for logging
	endpoint, _ := containerManager.GetContainerEndpoint(newContainer.Info, containerPort, aps.config.PerformerNetworkName)

	aps.logger.Info("Container deployment started, monitoring health status",
		zap.String("avsAddress", avsId),
		zap.String("containerID", newContainer.Info.ID),
		zap.String("endpoint", endpoint),
	)

	return statusChan, nil
}

// CancelDeployment cancels an in-progress deployment by cleaning up the nextContainer
func (aps *AvsPerformerServer) CancelDeployment(ctx context.Context) error {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	if aps.nextContainer == nil {
		// No deployment in progress
		return nil
	}

	aps.logger.Info("Canceling deployment, cleaning up next container",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", aps.nextContainer.Info.ID),
	)

	// Close the status channel to signal cancellation
	if aps.nextContainer.StatusChan != nil {
		close(aps.nextContainer.StatusChan)
	}

	// Get reference to container for cleanup
	containerToCleanup := aps.nextContainer
	aps.nextContainer = nil

	// Shutdown the container (this will stop monitoring and remove it)
	if err := aps.shutdownContainer(ctx, aps.config.AvsAddress, containerToCleanup); err != nil {
		aps.logger.Error("Failed to shutdown container during deployment cancellation",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", containerToCleanup.Info.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to cleanup canceled deployment: %w", err)
	}

	aps.logger.Info("Deployment canceled successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	return nil
}

// PromoteContainer promotes the nextContainer to currentContainer after verifying it's healthy
// If there's no nextContainer but currentContainer is the first deployment, it's a no-op
func (aps *AvsPerformerServer) PromoteContainer(ctx context.Context) error {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	// If there's no nextContainer, check if this is first deployment scenario
	if aps.nextContainer == nil {
		if aps.currentContainer != nil && aps.currentContainer.PerformerHealth.ApplicationHealth {
			// First deployment went directly to currentContainer, nothing to promote
			aps.logger.Info("Container already in current slot (first deployment), nothing to promote",
				zap.String("avsAddress", aps.config.AvsAddress),
			)
			return nil
		}
		return fmt.Errorf("no next container available to promote")
	}

	// Verify nextContainer is healthy
	if !aps.nextContainer.PerformerHealth.ContainerHealth || !aps.nextContainer.PerformerHealth.ApplicationHealth {
		return fmt.Errorf("cannot promote unhealthy container (container health: %v, application health: %v)",
			aps.nextContainer.PerformerHealth.ContainerHealth,
			aps.nextContainer.PerformerHealth.ApplicationHealth)
	}

	aps.logger.Info("Promoting next container to current container",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("newContainerID", aps.nextContainer.Info.ID),
	)

	// Stop the old container gracefully
	oldContainer := aps.currentContainer
	if oldContainer != nil && oldContainer.Info != nil {
		// Stop liveness monitoring
		aps.containerManager.StopLivenessMonitoring(oldContainer.Info.ID)

		// Stop and remove old container
		if err := aps.containerManager.Stop(ctx, oldContainer.Info.ID, 10*time.Second); err != nil {
			aps.logger.Warn("Failed to stop old container",
				zap.String("containerID", oldContainer.Info.ID),
				zap.Error(err),
			)
		}
		if err := aps.containerManager.Remove(ctx, oldContainer.Info.ID, true); err != nil {
			aps.logger.Warn("Failed to remove old container",
				zap.String("containerID", oldContainer.Info.ID),
				zap.Error(err),
			)
		}
	}

	// Promote next to current
	aps.currentContainer = aps.nextContainer
	aps.nextContainer = nil

	aps.logger.Info("Container promotion completed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("currentContainerID", aps.currentContainer.Info.ID),
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
		zap.String("containerID", container.Info.ID),
	)

	// Stop liveness monitoring for this container
	aps.containerManager.StopLivenessMonitoring(container.Info.ID)

	// Stop the container
	if err := aps.containerManager.Stop(ctx, container.Info.ID, 10*time.Second); err != nil {
		aps.logger.Error("Failed to stop container",
			zap.String("avsAddress", avsAddress),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
	}

	// Remove the container
	if err := aps.containerManager.Remove(ctx, container.Info.ID, true); err != nil {
		aps.logger.Error("Failed to remove container",
			zap.String("avsAddress", avsAddress),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
		return err
	}

	aps.logger.Info("Container shutdown completed",
		zap.String("avsAddress", avsAddress),
		zap.String("containerID", container.Info.ID),
	)

	return nil
}

func (aps *AvsPerformerServer) Shutdown() error {
	aps.logger.Sugar().Infow("Shutting down AVS performer server",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Safely get container references under mutex
	aps.deploymentMu.Lock()
	current := aps.currentContainer
	next := aps.nextContainer
	aps.deploymentMu.Unlock()

	// Shutdown both containers and collect any errors
	var errs []error

	if err := aps.shutdownContainer(ctx, aps.config.AvsAddress, current); err != nil {
		errs = append(errs, fmt.Errorf("current container: %w", err))
	}

	if err := aps.shutdownContainer(ctx, aps.config.AvsAddress, next); err != nil {
		errs = append(errs, fmt.Errorf("next container: %w", err))
	}

	// Shutdown the container manager
	if err := aps.containerManager.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("container manager: %w", err))
	}

	// Use errors.Join to combine all errors into one
	if combinedErr := stderrors.Join(errs...); combinedErr != nil {
		aps.logger.Error("Shutdown failed", zap.Error(combinedErr))
		return combinedErr
	}

	aps.logger.Sugar().Infow("AVS performer server shutdown completed",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	return nil
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
						ContainerID: containerID,
						Message:     fmt.Sprintf("Container unhealthy after %d consecutive health check failures", consecutiveFailures),
						Timestamp:   time.Now(),
					}:
						aps.logger.Info("Sent unhealthy status event",
							zap.String("avsAddress", aps.config.AvsAddress),
							zap.String("containerID", containerID),
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
					ContainerID: containerID,
					Message:     "Container is healthy and ready",
					Timestamp:   time.Now(),
				}:
					aps.logger.Info("Sent healthy status event",
						zap.String("avsAddress", aps.config.AvsAddress),
						zap.String("containerID", containerID),
					)
				default:
				}
			}
		}
	}
}
