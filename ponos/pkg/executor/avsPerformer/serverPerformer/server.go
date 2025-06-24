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
	// Create container configuration
	containerConfig := containerManager.CreateDefaultContainerConfig(
		aps.config.AvsAddress,
		aps.config.Image.Repository,
		aps.config.Image.Tag,
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

	// Create and start container
	containerInfo, endpoint, err := aps.createAndStartContainer(ctx)
	if err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown after container creation failure")
		}
		return err
	}
	aps.containerInfo = containerInfo

	// Create performer client
	perfClient, err := aps.createPerformerClient(endpoint)
	if err != nil {
		if shutdownErr := aps.Shutdown(); shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown container after client creation failure")
		}
		return err
	}
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

		for {
			select {
			case <-healthCtx.Done():
				aps.logger.Debug("Application health check cancelled",
					aps.logFields()...,
				)
				return
			case <-ticker.C:
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
	aps.logger.Info("Processing task", append(aps.logFields(), zap.Any("task", task))...)

	res, err := aps.performerClient.ExecuteTask(ctx, &performerV1.TaskRequest{
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

	if aps.containerInfo == nil || aps.containerManager == nil {
		return nil
	}

	aps.logger.Info("Shutting down AVS performer server",
		aps.logFieldsWithContainer()...,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Stop the container
	if err := aps.containerManager.Stop(ctx, aps.containerInfo.ID, 10*time.Second); err != nil {
		aps.logger.Error("Failed to stop container",
			append(aps.logFieldsWithContainer(), zap.Error(err))...,
		)
	}

	// Remove the container
	if err := aps.containerManager.Remove(ctx, aps.containerInfo.ID, true); err != nil {
		aps.logger.Error("Failed to remove container",
			append(aps.logFieldsWithContainer(), zap.Error(err))...,
		)
		return err
	}

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
