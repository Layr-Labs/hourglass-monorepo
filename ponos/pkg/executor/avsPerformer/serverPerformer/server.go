package serverPerformer

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ContainerInstance holds a container manager and its associated info
type ContainerInstance struct {
	Manager   containerManager.ContainerManager
	Info      *containerManager.ContainerInfo
	Client    performerV1.PerformerServiceClient
	EventChan <-chan containerManager.ContainerEvent
}

type AvsPerformerServer struct {
	config         *avsPerformer.AvsPerformerConfig
	logger         *zap.Logger
	peeringFetcher peering.IPeeringDataFetcher

	// Container management
	currentContainer *ContainerInstance
	nextContainer    *ContainerInstance
	deploymentMu     sync.Mutex

	aggregatorPeers []*peering.OperatorPeerInfo
}

// NewAvsPerformerServer creates a new AvsPerformerServer with the provided container manager
func NewAvsPerformerServer(
	config *avsPerformer.AvsPerformerConfig,
	peeringFetcher peering.IPeeringDataFetcher,
	logger *zap.Logger,
	containerMgr containerManager.ContainerManager,
) *AvsPerformerServer {
	return &AvsPerformerServer{
		config:         config,
		logger:         logger,
		peeringFetcher: peeringFetcher,
		currentContainer: &ContainerInstance{
			Manager: containerMgr,
		},
	}
}

const containerPort = 8080

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

// createAndStartContainer handles the common container creation steps
// Returns the created ContainerInstance with populated Manager, Info, Client, and EventChan
func (aps *AvsPerformerServer) createAndStartContainer(
	ctx context.Context,
	containerMgr containerManager.ContainerManager,
	avsAddress string,
	imageRepo string,
	imageTag string,
	networkName string,
) (*ContainerInstance, error) {
	// Use the default factory method from containerManager
	result, err := containerManager.CreateAndStartDefaultContainer(
		ctx,
		containerMgr,
		avsAddress,
		imageRepo,
		imageTag,
		containerPort,
		networkName,
		aps.logger,
	)
	if err != nil {
		return nil, err
	}

	// Create performer client
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(result.Endpoint, true)
	if err != nil {
		// Clean up on failure
		if removeErr := containerMgr.Remove(ctx, result.Info.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container during cleanup",
				zap.String("containerID", result.Info.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to create performer client")
	}

	// Start liveness monitoring for this container
	livenessConfig := containerManager.NewDefaultAvsPerformerLivenessConfig()
	eventChan, err := containerMgr.StartLivenessMonitoring(ctx, result.Info.ID, livenessConfig)
	if err != nil {
		// Clean up on failure
		if removeErr := containerMgr.Remove(ctx, result.Info.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove container during monitoring setup failure",
				zap.String("containerID", result.Info.ID),
				zap.Error(removeErr),
			)
		}
		return nil, errors.Wrap(err, "failed to start liveness monitoring")
	}

	// Create the container instance with all components
	containerInstance := &ContainerInstance{
		Manager:   containerMgr,
		Info:      result.Info,
		Client:    perfClient,
		EventChan: eventChan,
	}

	aps.logger.Info("Container created and monitoring started",
		zap.String("avsAddress", avsAddress),
		zap.String("containerID", result.Info.ID),
		zap.String("endpoint", result.Endpoint),
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

	// Create and start container
	containerInstance, err := aps.createAndStartContainer(
		ctx,
		aps.currentContainer.Manager,
		aps.config.AvsAddress,
		aps.config.Image.Repository,
		aps.config.Image.Tag,
		aps.config.PerformerNetworkName,
	)
	if err != nil {
		return err
	}
	aps.currentContainer = containerInstance

	// Start application-level health checking
	aps.startApplicationHealthMonitor(ctx)

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
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", event.ContainerID),
		zap.String("eventType", string(event.Type)),
		zap.String("message", event.Message),
		zap.Int("restartCount", event.State.RestartCount),
	)

	switch event.Type {
	case containerManager.EventStarted:
		aps.logger.Info("Container started successfully",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
		)

	case containerManager.EventCrashed:
		aps.logger.Error("Container crashed",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.Int("exitCode", event.State.ExitCode),
			zap.Int("restartCount", event.State.RestartCount),
			zap.String("error", event.State.Error),
		)
		// Auto-restart is handled by containerManager based on RestartPolicy

	case containerManager.EventOOMKilled:
		aps.logger.Error("Container killed due to OOM",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.Int("restartCount", event.State.RestartCount),
		)
		// Auto-restart is handled by containerManager

	case containerManager.EventRestarted:
		aps.logger.Info("Container restarted successfully",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.Int("restartCount", event.State.RestartCount),
		)
		// Recreate performer client connection after restart
		go aps.recreatePerformerClient(ctx)

	case containerManager.EventRestartFailed:
		aps.logger.Error("Container restart failed",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.String("error", event.Message),
			zap.Int("restartCount", event.State.RestartCount),
		)
		// Could potentially signal the executor to take additional action

		// Check if restart failed because container doesn't exist (needs recreation)
		if strings.Contains(event.Message, "recreation needed") {
			aps.logger.Info("Container recreation needed, attempting to recreate",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("containerID", event.ContainerID),
			)
			go aps.recreateContainer(ctx)
		}

	case containerManager.EventHealthy:
		aps.logger.Debug("Container Docker healthy signal received",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
		)

		// Check if this healthy event is for a nextContainer that should be promoted
		aps.deploymentMu.Lock()
		if aps.nextContainer != nil && aps.nextContainer.Info != nil && aps.nextContainer.Info.ID == event.ContainerID {
			containerToPromote := aps.nextContainer
			aps.deploymentMu.Unlock()

			aps.logger.Info("Next container became Docker-healthy, verifying application health before promotion",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("nextContainerID", containerToPromote.Info.ID),
			)

			// Check application health before promoting
			if err := aps.checkContainerHealth(ctx, containerToPromote); err != nil {
				aps.logger.Warn("Next container not ready for promotion - application health check failed",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("containerID", containerToPromote.Info.ID),
					zap.Error(err),
				)
				// Don't promote yet, continuous health monitor will eventually promote when app is ready
				return
			}

			aps.logger.Info("Next container is ready for tasks",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("nextContainerID", containerToPromote.Info.ID),
			)

			if err := aps.promoteNextContainer(ctx); err != nil {
				aps.logger.Error("Failed to automatically promote healthy next container",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("containerID", containerToPromote.Info.ID),
					zap.Error(err),
				)
			} else {
				aps.logger.Info("Successfully promoted fully-healthy next container to current",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("containerID", containerToPromote.Info.ID),
				)
			}

		} else {
			aps.deploymentMu.Unlock()
		}

	case containerManager.EventUnhealthy:
		aps.logger.Warn("Container is unhealthy",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.String("reason", event.Message),
		)
		// The container manager will handle auto-restart based on policy
		// Application can decide to trigger manual restart if needed

	case containerManager.EventRestarting:
		aps.logger.Info("Container is being restarted",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
			zap.String("reason", event.Message),
		)
	}
}

// recreateContainer recreates a container that was killed/removed
func (aps *AvsPerformerServer) recreateContainer(ctx context.Context) {
	aps.logger.Info("Starting container recreation",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("previousContainerID", aps.currentContainer.Info.ID),
	)

	// Stop monitoring the old container
	if aps.currentContainer != nil && aps.currentContainer.Info != nil {
		aps.currentContainer.Manager.StopLivenessMonitoring(aps.currentContainer.Info.ID)
	}

	// Create and start new container
	containerInstance, err := aps.createAndStartContainer(
		ctx,
		aps.currentContainer.Manager,
		aps.config.AvsAddress,
		aps.config.Image.Repository,
		aps.config.Image.Tag,
		aps.config.PerformerNetworkName,
	)
	if err != nil {
		aps.logger.Error("Failed to recreate container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return
	}
	aps.currentContainer = containerInstance

	// Start monitoring events for the new container
	go aps.monitorContainerEvents(ctx, aps.currentContainer.EventChan)

	aps.logger.Info("Container recreation completed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("newContainerID", aps.currentContainer.Info.ID),
	)
}

// recreatePerformerClient recreates the performer client connection after container restart
func (aps *AvsPerformerServer) recreatePerformerClient(ctx context.Context) {
	// Wait a moment for the container to fully start
	time.Sleep(2 * time.Second)

	// Get updated container information
	updatedInfo, err := aps.currentContainer.Manager.Inspect(ctx, aps.currentContainer.Info.ID)
	if err != nil {
		aps.logger.Error("Failed to inspect container after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return
	}
	aps.currentContainer.Info = updatedInfo

	// Get the new container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, containerPort, aps.config.PerformerNetworkName)
	if err != nil {
		aps.logger.Error("Failed to get container endpoint after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return
	}

	// Create new performer client
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(endpoint, true)
	if err != nil {
		aps.logger.Error("Failed to recreate performer client after restart",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return
	}

	aps.currentContainer.Client = perfClient
	aps.logger.Info("Performer client recreated successfully after container restart",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("endpoint", endpoint),
	)
}

// TriggerContainerRestart allows the application to manually trigger a container restart
func (aps *AvsPerformerServer) TriggerContainerRestart(reason string) error {
	if aps.currentContainer == nil || aps.currentContainer.Manager == nil || aps.currentContainer.Info == nil {
		return fmt.Errorf("container manager or container info not available")
	}

	aps.logger.Info("Triggering manual container restart",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", aps.currentContainer.Info.ID),
		zap.String("reason", reason),
	)

	return aps.currentContainer.Manager.TriggerRestart(aps.currentContainer.Info.ID, reason)
}

// startApplicationHealthMonitor performs continuous application-level health checks
// for both currentContainer and nextContainer when they exist
func (aps *AvsPerformerServer) startApplicationHealthMonitor(ctx context.Context) {
	aps.logger.Info("Starting application health monitor",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	// Start the health monitor in a goroutine that runs for the lifetime of the server
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// Track failures separately for current and next containers
		currentContainerFailures := 0
		nextContainerFailures := 0
		const maxConsecutiveFailures = 3

		for {
			select {
			case <-ctx.Done():
				aps.logger.Info("Continuous application health monitor stopped",
					zap.String("avsAddress", aps.config.AvsAddress),
				)
				return
			case <-ticker.C:
				// Safely get current container references under mutex
				aps.deploymentMu.Lock()
				currentContainer := aps.currentContainer
				nextContainer := aps.nextContainer
				aps.deploymentMu.Unlock()

				// Now check health on local copies (thread-safe)
				if currentContainer != nil && currentContainer.Client != nil {
					if err := aps.checkContainerHealth(ctx, currentContainer); err != nil {
						currentContainerFailures++
						aps.logger.Warn("Current container health check failed",
							zap.String("avsAddress", aps.config.AvsAddress),
							zap.String("containerID", currentContainer.Info.ID),
							zap.Error(err),
							zap.Int("consecutiveFailures", currentContainerFailures),
						)

						// Trigger restart for current container after consecutive failures
						if currentContainerFailures >= maxConsecutiveFailures {
							aps.logger.Error("Current container health check failed multiple times, triggering restart",
								zap.String("avsAddress", aps.config.AvsAddress),
								zap.String("containerID", currentContainer.Info.ID),
								zap.Int("consecutiveFailures", currentContainerFailures),
							)

							if restartErr := aps.TriggerContainerRestart(fmt.Sprintf("application health check failed %d consecutive times", currentContainerFailures)); restartErr != nil {
								aps.logger.Error("Failed to trigger current container restart",
									zap.String("avsAddress", aps.config.AvsAddress),
									zap.Error(restartErr),
								)
							}
							currentContainerFailures = 0 // Reset after triggering restart
						}
					} else {
						// Reset failure counter on successful health check
						if currentContainerFailures > 0 {
							aps.logger.Info("Current container health recovered",
								zap.String("avsAddress", aps.config.AvsAddress),
								zap.String("containerID", currentContainer.Info.ID),
								zap.Int("previousFailures", currentContainerFailures),
							)
							currentContainerFailures = 0
						}
					}
				}

				// Check next container health (using local copy)
				if nextContainer != nil && nextContainer.Client != nil {
					if err := aps.checkContainerHealth(ctx, nextContainer); err != nil {
						nextContainerFailures++
						aps.logger.Warn("Next container health check failed",
							zap.String("avsAddress", aps.config.AvsAddress),
							zap.String("containerID", nextContainer.Info.ID),
							zap.Error(err),
							zap.Int("consecutiveFailures", nextContainerFailures),
						)

						// For next container, abort deployment after consecutive failures
						if nextContainerFailures >= maxConsecutiveFailures {
							// Still need mutex for modifying the actual fields
							aps.deploymentMu.Lock()
							if aps.nextContainer != nil && aps.nextContainer.Info.ID == nextContainer.Info.ID {
								// Double-check it's still the same container
								containerID := aps.nextContainer.Info.ID
								aps.logger.Info("Aborting deployment due to health check failures",
									zap.String("avsAddress", aps.config.AvsAddress),
									zap.String("nextContainerID", containerID),
								)

								// Clean up the next container
								if err := aps.shutdownContainer(ctx, aps.config.AvsAddress, aps.nextContainer); err != nil {
									aps.logger.Error("Failed to shutdown next container during deployment abort",
										zap.String("avsAddress", aps.config.AvsAddress),
										zap.String("containerID", containerID),
										zap.Error(err),
									)
								}

								// Clear the next container slot
								aps.nextContainer = nil
							}
							aps.deploymentMu.Unlock()
							nextContainerFailures = 0 // Reset after aborting deployment
						}
					} else {
						// Reset failure counter on successful health check
						if nextContainerFailures > 0 {
							aps.logger.Info("Next container health recovered",
								zap.String("avsAddress", aps.config.AvsAddress),
								zap.String("containerID", nextContainer.Info.ID),
								zap.Int("previousFailures", nextContainerFailures),
							)
							nextContainerFailures = 0
						}
					}
				} else {
					// Reset next container failure counter when no next container exists
					nextContainerFailures = 0
				}
			}
		}
	}()
}

// checkContainerHealth performs a single health check on the specified container
func (aps *AvsPerformerServer) checkContainerHealth(ctx context.Context, container *ContainerInstance) error {
	healthCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	res, err := container.Client.HealthCheck(healthCtx, &performerV1.HealthCheckRequest{})
	if err != nil {
		return err
	}

	aps.logger.Debug("Container health check successful",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", container.Info.ID),
		zap.String("status", res.Status.String()),
	)

	return nil
}

// DeployContainer deploys a new container as the nextContainer
// If deployment is successful, it can later be promoted to currentContainer
func (aps *AvsPerformerServer) DeployContainer(
	ctx context.Context,
	avsId string,
	config *avsPerformer.AvsPerformerConfig,
	containerMgr containerManager.ContainerManager,
) error {
	// Lock to prevent concurrent deployments
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	// Check if deployment is already in progress
	if aps.nextContainer != nil {
		return fmt.Errorf("deployment in progress, please wait for it to complete")
	}

	aps.logger.Info("Starting container deployment",
		zap.String("avsAddress", avsId),
		zap.String("imageRepository", config.Image.Repository),
		zap.String("imageTag", config.Image.Tag),
	)

	// Create the next container instance
	nextContainer, err := aps.createAndStartContainer(ctx, containerMgr, avsId, config.Image.Repository, config.Image.Tag, config.PerformerNetworkName)
	if err != nil {
		return errors.Wrap(err, "failed to create next container")
	}

	// Set the next container
	aps.nextContainer = nextContainer

	// Start monitoring events for the next container
	go aps.monitorContainerEvents(ctx, aps.nextContainer.EventChan)

	// Get endpoint for logging
	endpoint, _ := containerManager.GetContainerEndpoint(nextContainer.Info, containerPort, config.PerformerNetworkName)

	aps.logger.Info("Container deployment completed successfully",
		zap.String("avsAddress", avsId),
		zap.String("containerID", nextContainer.Info.ID),
		zap.String("endpoint", endpoint),
	)

	return nil
}

// promoteNextContainer promotes the nextContainer to currentContainer
// This should be called after verifying the nextContainer is working properly
func (aps *AvsPerformerServer) promoteNextContainer(ctx context.Context) error {
	aps.deploymentMu.Lock()
	defer aps.deploymentMu.Unlock()

	if aps.nextContainer == nil {
		return fmt.Errorf("no next container available to promote")
	}

	aps.logger.Info("Promoting next container to current container",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("newContainerID", aps.nextContainer.Info.ID),
		zap.String("oldContainerID", aps.currentContainer.Info.ID),
	)

	// Stop the old container gracefully
	oldContainer := aps.currentContainer
	if oldContainer != nil && oldContainer.Info != nil {
		// Stop liveness monitoring
		oldContainer.Manager.StopLivenessMonitoring(oldContainer.Info.ID)

		// Stop and remove old container
		if err := oldContainer.Manager.Stop(ctx, oldContainer.Info.ID, 10*time.Second); err != nil {
			aps.logger.Warn("Failed to stop old container",
				zap.String("containerID", oldContainer.Info.ID),
				zap.Error(err),
			)
		}
		if err := oldContainer.Manager.Remove(ctx, oldContainer.Info.ID, true); err != nil {
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

	verfied, err := sig.Verify(peer.PublicKey, t.Payload)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to verify signature",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("aggregatorAddress", t.AggregatorAddress),
			zap.Error(err),
		)
		return err
	}
	if !verfied {
		aps.logger.Sugar().Errorw("Failed to verify signature",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("publicKey", string(peer.PublicKey.Bytes())),
			zap.Error(err),
		)
		return fmt.Errorf("failed to verify signature")
	}

	return nil
}

func (aps *AvsPerformerServer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	aps.logger.Sugar().Infow("Processing task", zap.Any("task", task))

	res, err := aps.currentContainer.Client.ExecuteTask(ctx, &performerV1.TaskRequest{
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
func (aps *AvsPerformerServer) shutdownContainer(ctx context.Context, avsAddress string, container *ContainerInstance) error {
	if container == nil || container.Info == nil || container.Manager == nil {
		return nil
	}

	aps.logger.Info("Shutting down container",
		zap.String("avsAddress", avsAddress),
		zap.String("containerID", container.Info.ID),
	)

	// Stop liveness monitoring for this container
	container.Manager.StopLivenessMonitoring(container.Info.ID)

	// Stop the container
	if err := container.Manager.Stop(ctx, container.Info.ID, 10*time.Second); err != nil {
		aps.logger.Error("Failed to stop container",
			zap.String("avsAddress", avsAddress),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
	}

	// Remove the container
	if err := container.Manager.Remove(ctx, container.Info.ID, true); err != nil {
		aps.logger.Error("Failed to remove container",
			zap.String("avsAddress", avsAddress),
			zap.String("containerID", container.Info.ID),
			zap.Error(err),
		)
		return err
	}

	// Shutdown the container manager
	if err := container.Manager.Shutdown(ctx); err != nil {
		aps.logger.Error("Failed to shutdown container manager",
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

	// Shutdown both containers and collect any errors
	var errs []error

	if err := aps.shutdownContainer(ctx, aps.config.AvsAddress, aps.currentContainer); err != nil {
		errs = append(errs, fmt.Errorf("current container: %w", err))
	}

	if err := aps.shutdownContainer(ctx, aps.config.AvsAddress, aps.nextContainer); err != nil {
		errs = append(errs, fmt.Errorf("next container: %w", err))
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
