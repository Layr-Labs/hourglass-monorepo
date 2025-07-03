package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Client struct {
	cli *client.Client
}

type ContainerConfig struct {
	Image        string
	Cmd          []string
	Env          []string
	CPULimit     int64
	MemoryLimit  int64
	DiskLimit    int64
	NetworkMode  string
	Labels       map[string]string
	AutoRemove   bool
}

type ContainerInfo struct {
	ID        string
	State     string
	StartedAt time.Time
	ExitCode  int
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

func (c *Client) PullImage(ctx context.Context, image string) error {
	reader, err := c.cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	return err
}

func (c *Client) CreateContainer(ctx context.Context, name string, config *ContainerConfig) (string, error) {
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			CPUQuota:  config.CPULimit * 100000,
			Memory:    config.MemoryLimit,
			DiskQuota: config.DiskLimit,
		},
		NetworkMode: container.NetworkMode(config.NetworkMode),
		AutoRemove:  config.AutoRemove,
	}

	containerConfig := &container.Config{
		Image:  config.Image,
		Cmd:    config.Cmd,
		Env:    config.Env,
		Labels: config.Labels,
	}

	resp, err := c.cli.ContainerCreate(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	if err := c.cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

func (c *Client) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	timeoutSec := int(timeout.Seconds())
	if err := c.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSec}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

func (c *Client) RemoveContainer(ctx context.Context, containerID string) error {
	if err := c.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

func (c *Client) InspectContainer(ctx context.Context, containerID string) (*ContainerInfo, error) {
	inspect, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	info := &ContainerInfo{
		ID:       inspect.ID,
		State:    inspect.State.Status,
		ExitCode: inspect.State.ExitCode,
	}

	if inspect.State.StartedAt != "" {
		startedAt, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
		if err == nil {
			info.StartedAt = startedAt
		}
	}

	return info, nil
}

func (c *Client) ListContainers(ctx context.Context, labels map[string]string) ([]types.Container, error) {
	filters := types.ContainerListOptions{
		All: true,
	}

	containers, err := c.cli.ContainerList(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var filtered []types.Container
	for _, container := range containers {
		match := true
		for k, v := range labels {
			if container.Labels[k] != v {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, container)
		}
	}

	return filtered, nil
}

func (c *Client) GetContainerLogs(ctx context.Context, containerID string, since time.Time) (io.ReadCloser, error) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      since.Format(time.RFC3339),
		Follow:     false,
	}

	logs, err := c.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return logs, nil
}

func (c *Client) GetContainerStats(ctx context.Context, containerID string) (*types.StatsJSON, error) {
	stats, err := c.cli.ContainerStatsOneShot(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	var statsJSON types.StatsJSON
	if err := statsJSON.Decode(stats.Body); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return &statsJSON, nil
}

func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

func (c *Client) Close() error {
	return c.cli.Close()
}