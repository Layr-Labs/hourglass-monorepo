package serverPerformer

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"time"
)

func (aps *AvsPerformerServer) waitForRunning(ctx context.Context, dockerClient *client.Client, containerId string) (bool, error) {
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
				aps.logger.Sugar().Infow("Port map not yet available", zap.String("containerId", containerId))
				continue
			}
			if len(portMap) == 0 {
				aps.logger.Sugar().Infow("Port map is empty", zap.String("containerId", containerId))
				continue
			}
			aps.logger.Sugar().Infow("Container is running with port exposed",
				zap.String("containerId", containerId),
				zap.String("exposedPort", portMap[0].HostPort),
			)
			return true, nil
		}

		// Not ready yet, sleep and retry
		time.Sleep(100 * time.Millisecond * time.Duration(attempts+1))
	}
	return false, fmt.Errorf("container %s is not running after 10 attempts", containerId)
}

func (aps *AvsPerformerServer) startHealthCheck(ctx context.Context) {
	for {
		time.Sleep(5 * time.Second)
		res, err := aps.performerClient.GetHealth(ctx)
		if err != nil {
			aps.logger.Sugar().Errorw("Failed to get health from performer",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.Error(err),
			)
			continue
		}
		aps.logger.Sugar().Infow("Got health response",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("status", res.Status),
		)
	}
}
