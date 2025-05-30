package serverPerformer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/artifactMonitor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/runtime"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"go.uber.org/zap"
)

type AvsPerformerServer struct {
	config          *avsPerformer.AvsPerformerConfig
	logger          *zap.Logger
	containerId     string
	performerClient performerV1.PerformerServiceClient

	// Container runtime controller
	containerController runtime.IContainerRuntimeController

	peeringFetcher peering.IPeeringDataFetcher

	aggregatorPeers []*peering.OperatorPeerInfo

	// Artifact monitoring
	artifactMonitor    *artifactMonitor.PerformerArtifactMonitor
	artifactContainers map[string]*ArtifactContainer
	artifactsMutex     sync.RWMutex

	// Context for health checks
	healthCheckCancel context.CancelFunc
}

// ArtifactContainer represents a running container for a specific artifact
type ArtifactContainer struct {
	Artifact        *artifactMonitor.PerformerArtifact
	ContainerID     string
	PerformerClient performerV1.PerformerServiceClient
	CreatedAt       time.Time
	CancelFunc      context.CancelFunc // To cancel individual health checks
}

func NewAvsPerformerServer(
	config *avsPerformer.AvsPerformerConfig,
	peeringFetcher peering.IPeeringDataFetcher,
	artifactMonitor *artifactMonitor.PerformerArtifactMonitor,
	containerController runtime.IContainerRuntimeController,
	logger *zap.Logger,
) (*AvsPerformerServer, error) {
	return &AvsPerformerServer{
		config:              config,
		logger:              logger,
		containerController: containerController,
		peeringFetcher:      peeringFetcher,
		artifactMonitor:     artifactMonitor,
		artifactContainers:  make(map[string]*ArtifactContainer),
	}, nil
}

const artifactCheckInterval = 30 * time.Second

// take a sha hash of the avs address and return the first 6 chars
func hashAvsAddress(avsAddress string) string {
	hasher := sha256.New()
	hasher.Write([]byte(avsAddress))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)[0:6]
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

func (aps *AvsPerformerServer) Initialize(ctx context.Context) error {
	// Fetch aggregator peers
	aggregatorPeers, err := aps.fetchAggregatorPeerInfo(ctx)
	if err != nil {
		return err
	}
	aps.aggregatorPeers = aggregatorPeers
	aps.logger.Sugar().Infow("Fetched aggregator peers",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.Any("aggregatorPeers", aps.aggregatorPeers),
	)

	// Create network if needed
	if aps.config.PerformerNetworkName != "" {
		if err := aps.containerController.CreateNetworkIfNotExists(ctx, aps.config.PerformerNetworkName); err != nil {
			aps.logger.Sugar().Errorw("Failed to create Docker network for performer",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.Error(err),
			)
			return err
		}
	}

	// Create default container
	hostname := fmt.Sprintf("avs-performer-%s", hashAvsAddress(aps.config.AvsAddress))
	containerConfig := &runtime.ContainerConfig{
		Name:        hostname,
		Image:       fmt.Sprintf("%s:%s", aps.config.Image.Repository, aps.config.Image.Tag),
		NetworkName: aps.config.PerformerNetworkName,
		Labels: map[string]string{
			"avs.address": aps.config.AvsAddress,
		},
		Logger: aps.logger,
	}

	result, err := aps.containerController.CreateAndStartContainer(ctx, containerConfig)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to create default performer container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}

	aps.containerId = result.ContainerID
	aps.performerClient = result.PerformerClient

	aps.logger.Sugar().Infow("Initialized default performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", result.ContainerID),
		zap.String("exposedPort", result.ExposedPort),
	)

	// Start health check for default container
	healthCtx, cancel := context.WithCancel(ctx)
	aps.healthCheckCancel = cancel

	go aps.containerController.StartHealthCheck(healthCtx, &runtime.HealthCheckConfig{
		Client:      aps.performerClient,
		Identifier:  "default",
		ContainerID: aps.config.AvsAddress,
	})

	return nil
}

// Start begins monitoring for new artifacts and creating containers
func (aps *AvsPerformerServer) Start(ctx context.Context) {
	aps.logger.Sugar().Infow("Starting artifact monitoring for AVS performer",
		zap.String("avsAddress", aps.config.AvsAddress),
	)

	go aps.monitorArtifacts(ctx)
}

// monitorArtifacts periodically checks for new artifacts and creates containers
func (aps *AvsPerformerServer) monitorArtifacts(ctx context.Context) {
	ticker := time.NewTicker(artifactCheckInterval)
	defer ticker.Stop()

	// Check immediately on startup
	aps.updateServerPool(ctx)

	for {
		select {
		case <-ctx.Done():
			aps.logger.Sugar().Infow("Stopping artifact monitoring",
				zap.String("avsAddress", aps.config.AvsAddress),
			)
			return
		case <-ticker.C:
			aps.updateServerPool(ctx)
		}
	}
}

// updateServerPool checks for new artifacts and creates containers for them
func (aps *AvsPerformerServer) updateServerPool(ctx context.Context) {
	artifacts, err := aps.artifactMonitor.GetArtifacts(aps.config.AvsAddress)
	if err != nil {
		aps.logger.Sugar().Debugw("No artifacts found for AVS",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return
	}

	aps.artifactsMutex.RLock()
	currentArtifacts := make(map[string]bool)
	for digest := range aps.artifactContainers {
		currentArtifacts[digest] = true
	}
	aps.artifactsMutex.RUnlock()

	// Find new artifacts
	for _, artifact := range artifacts {
		if _, exists := currentArtifacts[artifact.Digest]; !exists {
			aps.logger.Sugar().Infow("Found new artifact to deploy",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("digest", artifact.Digest),
				zap.String("operatorSetId", artifact.OperatorSetId),
				zap.String("registryUrl", artifact.RegistryUrl),
			)

			if err := aps.createArtifactContainer(ctx, artifact); err != nil {
				aps.logger.Sugar().Errorw("Failed to create container for artifact",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.String("digest", artifact.Digest),
					zap.Error(err),
				)
			}
		}
	}
}

// createArtifactContainer creates a new container for the given artifact
func (aps *AvsPerformerServer) createArtifactContainer(ctx context.Context, artifact *artifactMonitor.PerformerArtifact) error {
	// Generate unique container name based on AVS address and digest
	// TODO: add a uuid for this to be unique
	containerName := fmt.Sprintf("avs-performer-%s-%s",
		hashAvsAddress(aps.config.AvsAddress),
		artifact.Digest)

	// Determine image to use
	var image string
	if artifact.Digest != "" {
		// Use the registry URL from the artifact
		image = artifact.RegistryUrl
	} else {
		// Fall back to config image with digest as tag
		image = fmt.Sprintf("%s:%s", aps.config.Image.Repository, artifact.Digest)
	}

	aps.logger.Sugar().Infow("Creating container for artifact",
		zap.String("containerName", containerName),
		zap.String("image", image),
		zap.String("digest", artifact.Digest),
	)

	containerConfig := &runtime.ContainerConfig{
		Name:        containerName,
		Image:       image,
		NetworkName: aps.config.PerformerNetworkName,
		Labels: map[string]string{
			"avs.address":     aps.config.AvsAddress,
			"artifact.digest": artifact.Digest,
			"operator.set.id": artifact.OperatorSetId,
			"type":            "artifact",
		},
		Logger: aps.logger,
	}

	result, err := aps.containerController.CreateAndStartContainer(ctx, containerConfig)
	if err != nil {
		return fmt.Errorf("failed to create artifact container: %w", err)
	}

	// Create health check context for this container
	healthCtx, cancel := context.WithCancel(ctx)

	// Store the artifact container
	artifactContainer := &ArtifactContainer{
		Artifact:        artifact,
		ContainerID:     result.ContainerID,
		PerformerClient: result.PerformerClient,
		CreatedAt:       time.Now(),
		CancelFunc:      cancel,
	}

	aps.artifactsMutex.Lock()
	aps.artifactContainers[artifact.Digest] = artifactContainer
	aps.artifactsMutex.Unlock()

	aps.logger.Sugar().Infow("Successfully created container for artifact",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("digest", artifact.Digest),
		zap.String("containerID", result.ContainerID),
		zap.String("exposedPort", result.ExposedPort),
	)

	// Start health check for this container
	go aps.containerController.StartHealthCheck(healthCtx, &runtime.HealthCheckConfig{
		Client:      result.PerformerClient,
		Identifier:  artifact.Digest,
		ContainerID: result.ContainerID,
	})

	return nil
}

func (aps *AvsPerformerServer) Shutdown() error {
	// Cancel default container health check
	if aps.healthCheckCancel != nil {
		aps.healthCheckCancel()
	}

	// Shutdown all artifact containers
	aps.artifactsMutex.Lock()
	for digest, ac := range aps.artifactContainers {
		aps.logger.Sugar().Infow("Stopping artifact container",
			zap.String("digest", digest),
			zap.String("containerID", ac.ContainerID),
		)

		// Cancel health check
		if ac.CancelFunc != nil {
			ac.CancelFunc()
		}

		// Stop and remove container
		aps.containerController.StopAndRemoveContainer(ac.ContainerID)
	}
	aps.artifactContainers = make(map[string]*ArtifactContainer)
	aps.artifactsMutex.Unlock()

	// Shutdown the default container
	if len(aps.containerId) == 0 {
		return nil
	}

	aps.logger.Sugar().Infow("Stopping default Docker container",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", aps.containerId),
	)

	aps.containerController.StopAndRemoveContainer(aps.containerId)

	return nil
}

// GetPerformerClient returns the appropriate performer client based on the artifact digest
func (aps *AvsPerformerServer) GetPerformerClient(artifactDigest string) performerV1.PerformerServiceClient {
	if artifactDigest == "" {
		return aps.performerClient
	}

	aps.artifactsMutex.RLock()
	defer aps.artifactsMutex.RUnlock()

	if ac, exists := aps.artifactContainers[artifactDigest]; exists {
		return ac.PerformerClient
	}

	// Fall back to default performer if artifact not found
	aps.logger.Sugar().Warnw("Artifact container not found, using default performer",
		zap.String("artifactDigest", artifactDigest),
	)
	return aps.performerClient
}
