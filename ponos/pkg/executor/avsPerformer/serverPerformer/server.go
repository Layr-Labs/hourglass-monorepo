package serverPerformer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type AvsPerformerServer struct {
	config          *avsPerformer.AvsPerformerConfig
	logger          *zap.Logger
	containerId     string
	dockerClient    *client.Client
	performerClient performerV1.PerformerServiceClient
	// TODO(seanmcgary) make this an actual chan with a type
	taskBacklog chan *performerTask.PerformerTask

	peeringFetcher peering.IPeeringDataFetcher

	reportTaskResponse avsPerformer.ReceiveTaskResponse

	aggregatorPeers []*peering.OperatorPeerInfo
}

func NewAvsPerformerServer(
	config *avsPerformer.AvsPerformerConfig,
	peeringFetcher peering.IPeeringDataFetcher,
	reportTaskResponse avsPerformer.ReceiveTaskResponse,
	logger *zap.Logger,
) (*AvsPerformerServer, error) {
	return &AvsPerformerServer{
		config:             config,
		logger:             logger,
		taskBacklog:        make(chan *performerTask.PerformerTask, 50),
		reportTaskResponse: reportTaskResponse,
		peeringFetcher:     peeringFetcher,
	}, nil
}

const containerPort = 8080

// take a sha shash of the avs address and return the first 6 chars
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
	containerPortProto := nat.Port(fmt.Sprintf("%d/tcp", containerPort))

	aggregatorPeers, err := aps.fetchAggregatorPeerInfo(ctx)
	if err != nil {
		return err
	}
	aps.aggregatorPeers = aggregatorPeers
	aps.logger.Sugar().Infow("Fetched aggregator peers",
		zap.String("avsAddress", aps.config.AvsAddress),
		zap.Any("aggregatorPeers", aps.aggregatorPeers),
	)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to create Docker perfClient for performer",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		return err
	}
	dockerClient.NegotiateAPIVersion(ctx)
	aps.dockerClient = dockerClient

	hostname := fmt.Sprintf("avs-performer-%s", hashAvsAddress(aps.config.AvsAddress))

	aps.logger.Sugar().Infow("Using hostname",
		zap.String("hostname", hostname),
	)

	if aps.config.PerformerNetworkName != "" {
		if err := aps.createNetworkIfNotExists(ctx, dockerClient, aps.config.PerformerNetworkName); err != nil {
			aps.logger.Sugar().Errorw("Failed to create Docker network for performer",
				zap.String("avsAddress", aps.config.AvsAddress),
				zap.Error(err),
			)
			return err
		}
	}

	containerConfg := &container.Config{
		Hostname: hostname,
		Image:    fmt.Sprintf("%s:%s", aps.config.Image.Repository, aps.config.Image.Tag),
		ExposedPorts: nat.PortSet{
			containerPortProto: struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		PortBindings: nat.PortMap{
			containerPortProto: []nat.PortBinding{
				{
					HostIP: "0.0.0.0",

					// leave this blank to let Docker handle creating a random port
					HostPort: "",
				},
			},
		},
	}

	var netConfig *network.NetworkingConfig
	if aps.config.PerformerNetworkName != "" {
		netConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				aps.config.PerformerNetworkName: {},
			},
		}
	}

	res, err := dockerClient.ContainerCreate(
		ctx,
		containerConfg,
		hostConfig,
		netConfig,
		nil,
		hostname,
	)
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

	running, err := aps.waitForRunning(ctx, dockerClient, res.ID, containerPortProto)
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
	if portMap, ok := containerInfo.NetworkSettings.Ports[containerPortProto]; !ok {
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

	containerHost := "localhost"
	if aps.config.PerformerNetworkName != "" {
		containerHost = hostname
		exposedPort = fmt.Sprintf("%d", containerPort)
		aps.logger.Sugar().Infow("Custom network provided, using container hostname and container port",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.String("containerHost", containerHost),
			zap.String("exposedPort", exposedPort),
			zap.String("containerID", res.ID),
		)
	}

	perfClient, err := avsPerformerClient.NewAvsPerformerClient(fmt.Sprintf("%s:%s", containerHost, exposedPort), true)
	if err != nil {
		aps.logger.Sugar().Errorw("Failed to create performer perfClient",
			zap.String("avsAddress", aps.config.AvsAddress),
			zap.Error(err),
		)
		shutdownErr := aps.Shutdown()
		if shutdownErr != nil {
			err = errors.Wrap(err, "failed to shutdown Docker container")
		}
		return err
	}
	aps.performerClient = perfClient

	go aps.startHealthCheck(ctx)

	return nil
}

func (aps *AvsPerformerServer) ProcessTasks(ctx context.Context) error {
	var wg sync.WaitGroup
	for i := 0; i < aps.config.WorkerCount; i++ {
		wg.Add(1)
	}
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		aps.logger.Sugar().Infow("Waiting for tasks", zap.String("avs", aps.config.AvsAddress))
		for task := range aps.taskBacklog {
			res, err := aps.processTask(ctx, task)
			if err != nil {
				aps.logger.Sugar().Errorw("Failed to process task",
					zap.String("avsAddress", aps.config.AvsAddress),
					zap.Error(err),
				)
				continue
			}
			aps.reportTaskResponse(task, res, err)
		}

	}(&wg)
	return nil
}

func (aps *AvsPerformerServer) processTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	aps.logger.Sugar().Infow("Processing task", zap.Any("task", task))

	res, err := aps.performerClient.ExecuteTask(ctx, &performerV1.TaskRequest{
		TaskId:   []byte(task.TaskID),
		Metadata: task.Metadata,
		Payload:  task.Payload,
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
			zap.Error(err),
		)
		return fmt.Errorf("failed to verify signature")
	}

	return nil
}

func (aps *AvsPerformerServer) RunTask(ctx context.Context, task *performerTask.PerformerTask) error {
	select {
	case aps.taskBacklog <- task:
		aps.logger.Sugar().Infow("PerformerTask added to backlog")
	default:
		aps.logger.Sugar().Infow("PerformerTask backlog is full, dropping task")
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
