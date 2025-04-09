package server

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/go-ponos/pkg/executor/avsPerformer"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type AvsPerformerServer struct {
	config       *avsPerformer.AvsPerformerConfig
	logger       *zap.Logger
	containerId  string
	dockerClient *client.Client
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
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
	}

	res, err := dockerClient.ContainerCreate(ctx, containerConfg, hostConfig, nil, nil, aps.config.AvsAddress)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to create Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}

	if err := dockerClient.ContainerStart(ctx, res.ID, container.StartOptions{}); err != nil {
		aps.logger.Sugar().Errorw("Failed to start Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}
	aps.containerId = res.ID
	aps.logger.Sugar().Infow("Started Docker container for performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", res.ID),
	)
	return nil
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
	if err := aps.dockerClient.ContainerStop(context.Background(), aps.containerId, container.StopOptions{}); err != nil {
		aps.logger.Sugar().Errorw("Failed to stop Docker container for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}
	aps.logger.Sugar().Infow("Stopped Docker container for performer",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.String("containerID", aps.containerId),
	)
	return nil
}
