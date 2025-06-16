package serverPerformer

import (
	"context"
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

type AvsPerformerServer struct {
	config           *avsPerformer.AvsPerformerConfig
	logger           *zap.Logger
	containerManager containerManager.ContainerManager
	containerInfo    *containerManager.ContainerInfo
	performerClient  performerV1.PerformerServiceClient

	peeringFetcher peering.IPeeringDataFetcher

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
	}, nil
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

	// Create container configuration
	containerConfig := containerManager.CreateDefaultContainerConfig(
		aps.config.AvsAddress,
		aps.config.Image.Repository,
		aps.config.Image.Tag,
		containerPort,
		aps.config.PerformerNetworkName,
	)

	aps.logger.Sugar().Infow("Using container configuration",
		zap.String("hostname", containerConfig.Hostname),
		zap.String("image", containerConfig.Image),
		zap.String("networkName", containerConfig.NetworkName),
	)

	// Create the container
	containerInfo, err := aps.containerManager.Create(ctx, containerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create container")
	}
	aps.containerInfo = containerInfo

	// Start the container
	if err := aps.containerManager.Start(ctx, containerInfo.ID); err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown container after start failure")
		}
		return errors.Wrap(err, "failed to start container")
	}

	// Wait for the container to be running with ports exposed
	if err := aps.containerManager.WaitForRunning(ctx, containerInfo.ID, 30*time.Second); err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown container after wait failure")
		}
		return errors.Wrap(err, "failed to wait for container to be running")
	}

	// Get updated container information with port mappings
	updatedInfo, err := aps.containerManager.Inspect(ctx, containerInfo.ID)
	if err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown container after inspect failure")
		}
		return errors.Wrap(err, "failed to inspect container")
	}
	aps.containerInfo = updatedInfo

	// Get the container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, containerPort, aps.config.PerformerNetworkName)
	if err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown container after endpoint failure")
		}
		return errors.Wrap(err, "failed to get container endpoint")
	}

	aps.logger.Sugar().Infow("Container started successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", containerInfo.ID),
		zap.String("endpoint", endpoint),
	)

	// Create performer client
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(endpoint, true)
	if err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown container after client creation failure")
		}
		return errors.Wrap(err, "failed to create performer client")
	}
	aps.performerClient = perfClient

	// Start liveness monitoring with auto-restart capabilities
	livenessConfig := &containerManager.LivenessConfig{
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
			RestartOnUnhealthy: true, // Enable automatic restart on health check failures
		},
		ResourceThresholds: containerManager.ResourceThresholds{
			CPUThreshold:    90.0,
			MemoryThreshold: 90.0,
			RestartOnCPU:    false, // Log warnings but don't auto-restart
			RestartOnMemory: false, // Log warnings but don't auto-restart
		},
		ResourceMonitoring:    true,
		ResourceCheckInterval: 30 * time.Second,
	}

	eventChan, err := aps.containerManager.StartLivenessMonitoring(ctx, containerInfo.ID, livenessConfig)
	if err != nil {
		aps.logger.Warn("Failed to start liveness monitoring", zap.Error(err))
	} else {
		go aps.monitorContainerEvents(ctx, eventChan)
	}

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
		aps.logger.Debug("Container health recovered",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", event.ContainerID),
		)

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
		zap.String("previousContainerID", aps.containerInfo.ID),
	)

	// Stop monitoring the old container
	if aps.containerInfo != nil {
		aps.containerManager.StopLivenessMonitoring(aps.containerInfo.ID)
	}

	// Create new container configuration (same as in Initialize)
	containerConfig := containerManager.CreateDefaultContainerConfig(
		aps.config.AvsAddress,
		aps.config.Image.Repository,
		aps.config.Image.Tag,
		containerPort,
		aps.config.PerformerNetworkName,
	)

	aps.logger.Info("Recreating container with configuration",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("hostname", containerConfig.Hostname),
		zap.String("image", containerConfig.Image),
		zap.String("networkName", containerConfig.NetworkName),
	)

	// Create the new container
	containerInfo, err := aps.containerManager.Create(ctx, containerConfig)
	if err != nil {
		aps.logger.Error("Failed to create new container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return
	}

	// Start the new container
	if err := aps.containerManager.Start(ctx, containerInfo.ID); err != nil {
		aps.logger.Error("Failed to start new container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", containerInfo.ID),
			zap.Error(err),
		)
		// Clean up the failed container
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return
	}

	// Wait for the container to be running
	if err := aps.containerManager.WaitForRunning(ctx, containerInfo.ID, 30*time.Second); err != nil {
		aps.logger.Error("Failed to wait for new container to be running",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", containerInfo.ID),
			zap.Error(err),
		)
		// Clean up the failed container
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return
	}

	// Get updated container information
	updatedInfo, err := aps.containerManager.Inspect(ctx, containerInfo.ID)
	if err != nil {
		aps.logger.Error("Failed to inspect new container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", containerInfo.ID),
			zap.Error(err),
		)
		// Clean up the failed container
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return
	}

	// Update container info
	aps.containerInfo = updatedInfo

	// Get the container endpoint
	endpoint, err := containerManager.GetContainerEndpoint(updatedInfo, containerPort, aps.config.PerformerNetworkName)
	if err != nil {
		aps.logger.Error("Failed to get new container endpoint",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", containerInfo.ID),
			zap.Error(err),
		)
		// Clean up the failed container
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return
	}

	// Create new performer client
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(endpoint, true)
	if err != nil {
		aps.logger.Error("Failed to create performer client for new container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", containerInfo.ID),
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
		// Clean up the failed container
		if removeErr := aps.containerManager.Remove(ctx, containerInfo.ID, true); removeErr != nil {
			aps.logger.Error("Failed to remove failed container",
				zap.String("containerID", containerInfo.ID),
				zap.Error(removeErr),
			)
		}
		return
	}

	aps.performerClient = perfClient

	// Start liveness monitoring for the new container
	livenessConfig := &containerManager.LivenessConfig{
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

	eventChan, err := aps.containerManager.StartLivenessMonitoring(ctx, containerInfo.ID, livenessConfig)
	if err != nil {
		aps.logger.Error("Failed to start liveness monitoring for new container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", containerInfo.ID),
			zap.Error(err),
		)
	} else {
		go aps.monitorContainerEvents(ctx, eventChan)
	}

	// Start new application-level health checking for the recreated container
	aps.startApplicationHealthCheck(ctx)

	aps.logger.Info("Container recreation completed successfully",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("newContainerID", containerInfo.ID),
		zap.String("endpoint", endpoint),
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
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return
	}
	aps.containerInfo = updatedInfo

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

	aps.performerClient = perfClient
	aps.logger.Info("Performer client recreated successfully after container restart",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("endpoint", endpoint),
	)
}

// TriggerContainerRestart allows the application to manually trigger a container restart
func (aps *AvsPerformerServer) TriggerContainerRestart(reason string) error {
	if aps.containerManager == nil || aps.containerInfo == nil {
		return fmt.Errorf("container manager or container info not available")
	}

	aps.logger.Info("Triggering manual container restart",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", aps.containerInfo.ID),
		zap.String("reason", reason),
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

	aps.logger.Sugar().Infow("Starting application health check",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerId", aps.containerInfo.ID),
	)

	// Start the health check in a goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		consecutiveFailures := 0
		const maxConsecutiveFailures = 3

		for {
			select {
			case <-healthCtx.Done():
				aps.logger.Debug("Application health check cancelled",
					zap.String("avsAddress", aps.config.AvsAddress),
				)
				return
			case <-ticker.C:
				if aps.performerClient == nil {
					continue
				}

				res, err := aps.performerClient.HealthCheck(healthCtx, &performerV1.HealthCheckRequest{})
				if err != nil {
					consecutiveFailures++
					aps.logger.Sugar().Errorw("Failed to get health from performer",
						zap.String("avsAddress", aps.config.AvsAddress),
						zap.Error(err),
						zap.Int("consecutiveFailures", consecutiveFailures),
					)

					// Trigger container restart if we've had too many consecutive failures
					if consecutiveFailures >= maxConsecutiveFailures {
						aps.logger.Error("Application health check failed multiple times, triggering container restart",
							zap.String("avsAddress", aps.config.AvsAddress),
							zap.Int("consecutiveFailures", consecutiveFailures),
						)

						if err := aps.TriggerContainerRestart(fmt.Sprintf("application health check failed %d consecutive times", consecutiveFailures)); err != nil {
							aps.logger.Error("Failed to trigger container restart",
								zap.String("avsAddress", aps.config.AvsAddress),
								zap.Error(err),
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
						zap.String("avsAddress", aps.config.AvsAddress),
						zap.Int("previousFailures", consecutiveFailures),
					)
					consecutiveFailures = 0
				}

				aps.logger.Sugar().Debugw("Got health response",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("status", res.Status.String()),
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
			zap.String("avsAddress", aps.config.AvsAddress),
		)
	}
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

	res, err := aps.performerClient.ExecuteTask(ctx, &performerV1.TaskRequest{
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

func (aps *AvsPerformerServer) Shutdown() error {
	// Stop application health check first
	aps.stopApplicationHealthCheck()

	if aps.containerInfo == nil || aps.containerManager == nil {
		return nil
	}

	aps.logger.Sugar().Infow("Shutting down AVS performer server",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", aps.containerInfo.ID),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Stop the container
	if err := aps.containerManager.Stop(ctx, aps.containerInfo.ID, 10*time.Second); err != nil {
		aps.logger.Sugar().Errorw("Failed to stop container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", aps.containerInfo.ID),
			zap.Error(err),
		)
	}

	// Remove the container
	if err := aps.containerManager.Remove(ctx, aps.containerInfo.ID, true); err != nil {
		aps.logger.Sugar().Errorw("Failed to remove container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", aps.containerInfo.ID),
			zap.Error(err),
		)
		return err
	}

	// Shutdown the container manager
	if err := aps.containerManager.Shutdown(ctx); err != nil {
		aps.logger.Sugar().Errorw("Failed to shutdown container manager",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}

	aps.logger.Sugar().Infow("AVS performer server shutdown completed",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	return nil
}
