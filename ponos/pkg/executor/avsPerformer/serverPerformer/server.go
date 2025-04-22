package serverPerformer

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
	"sync"
	"time"
)

type AvsPerformerServer struct {
	config          *avsPerformer.AvsPerformerConfig
	logger          *zap.Logger
	containerId     string
	dockerClient    *client.Client
	performerClient *avsPerformerClient.AvsPerformerClient
	// TODO(seanmcgary) make this an actual chan with a type
	taskBacklog chan interface{}

	// map of tasks by taskID
	inflightTasks sync.Map

	reportTaskResponse avsPerformer.ReceiveTaskResponse
}

func NewAvsPerformerServer(
	config *avsPerformer.AvsPerformerConfig,
	reportTaskResponse avsPerformer.ReceiveTaskResponse,
	logger *zap.Logger,
) (*AvsPerformerServer, error) {
	return &AvsPerformerServer{
		config:             config,
		logger:             logger,
		taskBacklog:        make(chan interface{}, 50),
		reportTaskResponse: reportTaskResponse,
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

func (aps *AvsPerformerServer) ProcessTasks(ctx context.Context) error {
	var wg sync.WaitGroup
	for i := 0; i < aps.config.WorkerCount; i++ {
		wg.Add(1)
	}
	go func() {
		aps.logger.Sugar().Infow("Waiting for tasks", zap.String("avs", aps.config.AvsAddress))
		for {
			select {
			case task := <-aps.taskBacklog:
				res, err := aps.processTask(ctx, task)
				aps.reportTaskResponse(res, err)
			case <-ctx.Done():
				aps.logger.Sugar().Infow("Shutting down task processing")
				wg.Done()
				return
			}
		}
	}()
	return nil
}

func (aps *AvsPerformerServer) processTask(ctx context.Context, task interface{}) (interface{}, error) {
	// TODO(seanmcgary): implement task processing
	// This is a placeholder implementation
	aps.logger.Sugar().Infow("Processing task", zap.Any("task", task))
	return nil, nil
}

func (aps *AvsPerformerServer) RunTask(ctx context.Context, task interface{}) error {
	select {
	case aps.taskBacklog <- task:
		aps.logger.Sugar().Infow("Task added to backlog")
	default:
		aps.logger.Sugar().Infow("Task backlog is full, dropping task")
		return fmt.Errorf("task backlog is full for avs %s", aps.config.AvsAddress)
	}
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
