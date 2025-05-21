package serverPerformer

import (
	"context"
	"fmt"
	"time"

	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
)

func (aps *AvsPerformerServer) waitForRunning(
	ctx context.Context,
	dockerClient *client.Client,
	containerId string,
	containerPort nat.Port,
) (bool, error) {
	for attempts := 0; attempts < 10; attempts++ {
		// Check for context cancellation
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		info, err := dockerClient.ContainerInspect(ctx, containerId)
		if err != nil {
			return false, err
		}

		if info.State.Running {
			containerInfo, err := dockerClient.ContainerInspect(ctx, containerId)
			if err != nil {
				return false, err
			}
			portMap, ok := containerInfo.NetworkSettings.Ports[containerPort]
			if !ok {
				aps.logger.Sugar().Infow("PollerPort map not yet available", zap.String("containerId", containerId))
				continue
			}
			if len(portMap) == 0 {
				aps.logger.Sugar().Infow("PollerPort map is empty", zap.String("containerId", containerId))
				continue
			}
			aps.logger.Sugar().Infow("Container is running with port exposed",
				zap.String("containerId", containerId),
				zap.String("exposedPort", portMap[0].HostPort),
			)
			return true, nil
		}

		// Not ready yet, sleep and retry
		time.Sleep(1 * time.Second * time.Duration(attempts+1))
	}
	return false, fmt.Errorf("container %s is not running after 10 attempts", containerId)
}

func (aps *AvsPerformerServer) createNetworkIfNotExists(ctx context.Context, dockerClient *client.Client, networkName string) error {
	networks, err := dockerClient.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	var n *network.Summary
	for _, net := range networks {
		if net.Name == networkName {
			n = &net
			break
		}
	}

	// net already exists
	if n != nil {
		return nil
	}

	_, err = dockerClient.NetworkCreate(
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
		return fmt.Errorf("failed to create net: %w", err)
	}
	aps.logger.Sugar().Infow("Created net",
		zap.String("networkName", networkName),
	)
	return nil
}

func (aps *AvsPerformerServer) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	aps.logger.Sugar().Infow("Starting health check loop",
		zap.String("avsAddress", aps.config.AvsAddress))

	for {
		select {
		case <-ctx.Done():
			aps.logger.Sugar().Infow("Health check loop terminated due to context cancellation",
				zap.String("avsAddress", aps.config.AvsAddress))
			return
		case <-ticker.C:
			aps.checkHealth(ctx)
		}
	}
}

func (aps *AvsPerformerServer) checkHealth(ctx context.Context) {
	// Create a timeout context for the health check
	healthCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Check container health first via Docker API
	isHealthy := true
	var healthErr error

	// Check Docker container health first
	if aps.containerId != "" && aps.dockerClient != nil {
		// Check if container is running
		inspection, err := aps.dockerClient.ContainerInspect(healthCtx, aps.containerId)
		if err != nil {
			isHealthy = false
			healthErr = fmt.Errorf("failed to inspect container: %w", err)
			aps.logger.Sugar().Warnw("Failed to inspect container health",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("containerId", aps.containerId),
				zap.Error(err),
			)
		} else if !inspection.State.Running {
			isHealthy = false
			healthErr = fmt.Errorf("container is not running")
			aps.logger.Sugar().Warnw("Container is not running",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("containerId", aps.containerId),
				zap.String("state", inspection.State.Status),
			)
		} else if inspection.State.Health != nil && inspection.State.Health.Status != "healthy" {
			isHealthy = false
			healthErr = fmt.Errorf("container health check failed: %s", inspection.State.Health.Status)
			aps.logger.Sugar().Warnw("Container health check failed",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("containerId", aps.containerId),
				zap.String("health", inspection.State.Health.Status),
			)
		}
	}

	// Only proceed with gRPC health check if container is healthy
	if isHealthy && aps.performerClient != nil {
		res, err := aps.performerClient.HealthCheck(healthCtx, &performerV1.HealthCheckRequest{})
		if err != nil {
			isHealthy = false
			healthErr = err
			aps.logger.Sugar().Warnw("Failed to get health from performer via gRPC",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.Error(err),
			)
		} else if res.Status != performerV1.PerformerStatus_READY_FOR_TASK {
			isHealthy = false
			healthErr = fmt.Errorf("performer reported unhealthy status: %s", res.Status.String())
			aps.logger.Sugar().Warnw("Performer reported unhealthy status via gRPC",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("status", res.Status.String()),
			)
		} else {
			aps.logger.Sugar().Debugw("Performer is healthy",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.String("status", res.Status.String()),
			)
		}
	}

	// Update health status
	aps.healthMutex.Lock()
	aps.isHealthy = isHealthy
	aps.lastHealthErr = healthErr
	aps.lastHealthCheck = time.Now()
	aps.healthMutex.Unlock()
}
