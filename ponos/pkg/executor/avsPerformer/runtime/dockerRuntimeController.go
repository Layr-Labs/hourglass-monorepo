package runtime

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
	"time"
)

// DockerRuntimeController implements IContainerRuntimeController using Docker
type DockerRuntimeController struct {
	dockerClient *client.Client
	logger       *zap.Logger
}

// NewDockerRuntimeController creates a new DockerRuntimeController instance
func NewDockerRuntimeController(logger *zap.Logger, dockerClient *client.Client) IContainerRuntimeController {
	return &DockerRuntimeController{
		dockerClient: dockerClient,
		logger:       logger,
	}
}

// CreateAndStartContainer creates and starts a Docker container with the given configuration
func (drc *DockerRuntimeController) CreateAndStartContainer(
	ctx context.Context,
	config *ContainerConfig,
) (*ContainerResult, error) {
	containerPortProto := nat.Port(fmt.Sprintf("%d/tcp", containerPort))

	// Container configuration
	containerConfig := &container.Config{
		Hostname: config.Name,
		Image:    config.Image,
		ExposedPorts: nat.PortSet{
			containerPortProto: struct{}{},
		},
		Labels: config.Labels,
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		AutoRemove: true,
		PortBindings: nat.PortMap{
			containerPortProto: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "", // Random port
				},
			},
		},
	}

	// Network configuration
	var netConfig *network.NetworkingConfig
	if config.NetworkName != "" {
		netConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				config.NetworkName: {},
			},
		}
	}

	// Create container
	res, err := drc.dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		netConfig,
		nil,
		config.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := drc.dockerClient.ContainerStart(ctx, res.ID, container.StartOptions{}); err != nil {
		drc.removeContainer(res.ID)
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	config.Logger.Sugar().Infow("Started Docker container",
		zap.String("containerID", res.ID),
		zap.String("name", config.Name),
	)

	// Wait for container to be running
	running, err := drc.waitForRunning(ctx, res.ID, containerPortProto)
	if err != nil || !running {
		drc.removeContainer(res.ID)
		return nil, fmt.Errorf("container failed to reach running state: %w", err)
	}

	// Get container info to find exposed port
	containerInfo, err := drc.dockerClient.ContainerInspect(ctx, res.ID)
	if err != nil {
		drc.removeContainer(res.ID)
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Extract exposed port
	exposedPort, err := drc.getExposedPort(containerInfo, containerPortProto)
	if err != nil {
		drc.removeContainer(res.ID)
		return nil, err
	}

	// Determine container host
	containerHost := "localhost"
	if config.NetworkName != "" {
		containerHost = config.Name
		exposedPort = fmt.Sprintf("%d", containerPort)
		config.Logger.Sugar().Infow("Custom network provided, using container hostname",
			zap.String("containerHost", containerHost),
			zap.String("exposedPort", exposedPort),
		)
	}

	// Create performer client
	perfClient, err := avsPerformerClient.NewAvsPerformerClient(fmt.Sprintf("%s:%s", containerHost, exposedPort), true)
	if err != nil {
		drc.removeContainer(res.ID)
		return nil, fmt.Errorf("failed to create performer client: %w", err)
	}

	return &ContainerResult{
		ContainerID:     res.ID,
		ExposedPort:     exposedPort,
		ContainerHost:   containerHost,
		PerformerClient: perfClient,
	}, nil
}

// StopAndRemoveContainer stops and removes a container
func (drc *DockerRuntimeController) StopAndRemoveContainer(containerID string) {
	// Stop container
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := drc.dockerClient.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		drc.logger.Sugar().Warnw("Failed to stop container",
			zap.String("containerID", containerID),
			zap.Error(err),
		)
	}
	cancel()

	// Remove container
	if err := drc.dockerClient.ContainerRemove(context.Background(), containerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		drc.logger.Sugar().Warnw("Failed to remove container",
			zap.String("containerID", containerID),
			zap.Error(err),
		)
	} else {
		drc.logger.Sugar().Infow("Removed container",
			zap.String("containerID", containerID),
		)
	}
}

// CreateNetworkIfNotExists creates a Docker network if it doesn't already exist
func (drc *DockerRuntimeController) CreateNetworkIfNotExists(ctx context.Context, networkName string) error {
	networks, err := drc.dockerClient.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	// Check if network already exists
	for _, net := range networks {
		if net.Name == networkName {
			return nil
		}
	}

	// Create network
	_, err = drc.dockerClient.NetworkCreate(
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
		return fmt.Errorf("failed to create network: %w", err)
	}
	drc.logger.Sugar().Infow("Created network",
		zap.String("networkName", networkName),
	)
	return nil
}

// StartHealthCheck monitors health of a performer
func (drc *DockerRuntimeController) StartHealthCheck(ctx context.Context, config *HealthCheckConfig) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			drc.CheckHealth(ctx, config.Client, config.Identifier, config.ContainerID)
		}
	}
}

// CheckHealth performs a health check on a performer client
func (drc *DockerRuntimeController) CheckHealth(ctx context.Context, client performerV1.PerformerServiceClient, identifier string, containerID string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := client.HealthCheck(ctx, &performerV1.HealthCheckRequest{})
	if err != nil {
		drc.logger.Sugar().Warnw("Health check failed",
			zap.String("identifier", identifier),
			zap.String("containerID", containerID),
			zap.Error(err),
		)
		return
	}

	drc.logger.Sugar().Debugw("Health check successful",
		zap.String("identifier", identifier),
		zap.String("containerID", containerID),
		zap.String("status", res.Status.String()),
	)
}

// getExposedPort extracts the exposed port from container info
func (drc *DockerRuntimeController) getExposedPort(containerInfo container.InspectResponse, containerPortProto nat.Port) (string, error) {
	portMap, ok := containerInfo.NetworkSettings.Ports[containerPortProto]
	if !ok {
		return "", fmt.Errorf("port map not found for %s", containerPortProto)
	}
	if len(portMap) == 0 {
		return "", fmt.Errorf("no exposed ports found in container")
	}
	return portMap[0].HostPort, nil
}

// removeContainer removes a container by ID (best effort)
func (drc *DockerRuntimeController) removeContainer(containerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := drc.dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		drc.logger.Sugar().Warnw("Failed to remove container",
			zap.String("containerID", containerID),
			zap.Error(err),
		)
	}
}

// waitForRunning waits for a container to be in running state with exposed ports
func (drc *DockerRuntimeController) waitForRunning(
	ctx context.Context,
	containerId string,
	containerPort nat.Port,
) (bool, error) {
	for attempts := 0; attempts < 10; attempts++ {
		info, err := drc.dockerClient.ContainerInspect(ctx, containerId)
		if err != nil {
			return false, err
		}

		if info.State.Running {
			containerInfo, err := drc.dockerClient.ContainerInspect(ctx, containerId)
			if err != nil {
				return false, err
			}
			portMap, ok := containerInfo.NetworkSettings.Ports[containerPort]
			if !ok {
				drc.logger.Sugar().Infow("Port map not yet available", zap.String("containerId", containerId))
				time.Sleep(time.Second * time.Duration(attempts+1))
				continue
			}
			if len(portMap) == 0 {
				drc.logger.Sugar().Infow("Port map is empty", zap.String("containerId", containerId))
				time.Sleep(time.Second * time.Duration(attempts+1))
				continue
			}
			drc.logger.Sugar().Infow("Container is running with port exposed",
				zap.String("containerId", containerId),
				zap.String("exposedPort", portMap[0].HostPort),
			)
			return true, nil
		}

		// Not ready yet, sleep and retry
		time.Sleep(time.Second * time.Duration(attempts+1))
	}
	return false, fmt.Errorf("container %s is not running after 10 attempts", containerId)
}
