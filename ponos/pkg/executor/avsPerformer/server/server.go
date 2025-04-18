package server

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type AvsPerformerServer struct {
	config          *avsPerformer.AvsPerformerConfig
	logger          *zap.Logger
	containerId     string
	dockerClient    *client.Client
	performerClient *avsPerformerClient.AvsPerformerClient
}

func NewAvsPerformerServer(
	config *avsPerformer.AvsPerformerConfig,
	logger *zap.Logger,
) (*AvsPerformerServer, error) {
	return &AvsPerformerServer{
		config: config,
		logger: logger,
	}, nil
}

const containerPort = "8080/tcp"

func (aps *AvsPerformerServer) Initialize(ctx context.Context) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to create Docker client for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}
	dockerClient.NegotiateAPIVersion(ctx)
	aps.dockerClient = dockerClient

	containerConfg := &container.Config{
		Image: fmt.Sprintf("%s:%s", aps.config.Image.Repository, aps.config.Image.Tag),
		ExposedPorts: nat.PortSet{
			containerPort: struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{
				{
					HostIP: "127.0.0.1",

					// leave this blank to let Docker handle creating a random port
					HostPort: "",
				},
			},
		},
	}

	res, err := dockerClient.ContainerCreate(ctx, containerConfg, hostConfig, nil, nil, aps.config.AvsAddress)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to create Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}
	aps.containerId = res.ID

	if err := dockerClient.ContainerStart(ctx, res.ID, container.StartOptions{}); err != nil {
		aps.logger.Sugar().Errorw("Failed to start Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		shutdownErr := aps.Shutdown()
		if shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown Docker container")
		}
		return err
	}
	aps.logger.Sugar().Infow("Started Docker container for performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", res.ID),
	)

	running, err := aps.waitForRunning(ctx, dockerClient, res.ID)
	if err != nil || !running {
		aps.logger.Sugar().Errorw("Failed to wait for Docker container to be running",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		shutdownErr := aps.Shutdown()
		if shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown Docker container")
		}
		return err
	}

	containerInfo, err := dockerClient.ContainerInspect(ctx, res.ID)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to inspect Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		shutdownErr := aps.Shutdown()
		if shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown Docker container")
		}
		return err
	}
	var exposedPort string
	if portMap, ok := containerInfo.NetworkSettings.Ports[containerPort]; !ok {
		aps.logger.Sugar().Errorw("Failed to get exposed port from Docker container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		shutdownErr := aps.Shutdown()
		if shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown Docker container")
		}
		return err
	} else if len(portMap) == 0 {
		aps.logger.Sugar().Errorw("No exposed ports found in Docker container",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
	} else {
		exposedPort = portMap[0].HostPort
	}

	aps.performerClient = avsPerformerClient.NewAvsPerformerClient(fmt.Sprintf("http://localhost:%s", exposedPort), nil)

	go aps.startHealthCheck(ctx)

	return nil
}

func (aps *AvsPerformerServer) waitForRunning(ctx context.Context, dockerClient *client.Client, containerId string) (bool, error) {
	for attempts := 0; attempts < 10; attempts++ {
		info, err := dockerClient.ContainerInspect(ctx, containerId)
		if err != nil {
			return false, err
		}

		if info.State.Running {
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

func (aps *AvsPerformerServer) RunTask(ctx context.Context) error {
	// Implement the logic to run the task
	return nil
}

func (aps *AvsPerformerServer) Shutdown() error {
	if len(aps.containerId) == 0 {
		return nil
	}
	if aps.dockerClient == nil {
		return nil
	}

	aps.logger.Sugar().Infow("Stopping Docker container for performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", aps.containerId),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := aps.dockerClient.ContainerStop(ctx, aps.containerId, container.StopOptions{}); err != nil {
		aps.logger.Sugar().Errorw("Failed to stop Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
	} else {
		aps.logger.Sugar().Infow("Stopped Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerID", aps.containerId),
		)
	}
	aps.logger.Sugar().Infow("Removing Docker container for performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", aps.containerId),
	)
	if err := aps.dockerClient.ContainerRemove(context.Background(), aps.containerId, container.RemoveOptions{
		Force: true,
	}); err != nil {
		aps.logger.Sugar().Errorw("Failed to remove Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}
	return nil
}
