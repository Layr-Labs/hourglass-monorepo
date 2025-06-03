package serverPerformer

import (
	"context"
	"fmt"
	"strings"
	"time"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
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
	healthChan       <-chan bool

	peeringFetcher peering.IPeeringDataFetcher

	aggregatorPeers []*peering.OperatorPeerInfo
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

	// Start health checking
	healthChan, err := aps.containerManager.StartHealthCheck(ctx, containerInfo.ID, nil)
	if err != nil {
		aps.logger.Warn("Failed to start health check", zap.Error(err))
	} else {
		aps.healthChan = healthChan
		go aps.monitorHealth(ctx)
	}

	// Start application-level health checking
	go aps.startApplicationHealthCheck(ctx)

	return nil
}

// monitorHealth monitors the health check channel and logs health status changes
func (aps *AvsPerformerServer) monitorHealth(ctx context.Context) {
	if aps.healthChan == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case healthy, ok := <-aps.healthChan:
			if !ok {
				aps.logger.Info("Health check channel closed")
				return
			}
			if healthy {
				aps.logger.Debug("Container health check passed",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("containerID", aps.containerInfo.ID),
				)
			} else {
				aps.logger.Error("Container health check failed",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("containerID", aps.containerInfo.ID),
				)
			}
		}
	}
}

// startApplicationHealthCheck performs application-level health checks via gRPC
func (aps *AvsPerformerServer) startApplicationHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if aps.performerClient == nil {
				continue
			}

			res, err := aps.performerClient.HealthCheck(ctx, &performerV1.HealthCheckRequest{})
			if err != nil {
				aps.logger.Sugar().Errorw("Failed to get health from performer",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.Error(err),
				)
				continue
			}
			aps.logger.Sugar().Debugw("Got health response",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("status", res.Status.String()),
			)
		}
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
