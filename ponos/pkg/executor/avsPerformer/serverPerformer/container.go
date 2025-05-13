package serverPerformer

import (
	"context"
	"fmt"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
	"time"
)

func (aps *AvsPerformerServer) waitForRunning(
	ctx context.Context,
	dockerClient *client.Client,
	containerId string,
	containerPort nat.Port,
) (bool, error) {
	for attempts := 0; attempts < 10; attempts++ {
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
	for {
		time.Sleep(5 * time.Second)
		res, err := aps.performerClient.HealthCheck(ctx, &performerV1.HealthCheckRequest{})
		if err != nil {
			aps.logger.Sugar().Errorw("Failed to get health from performer",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.Error(err),
			)
			continue
		}
		aps.logger.Sugar().Infow("Got health response",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("status", res.Status.String()),
		)
	}
}
