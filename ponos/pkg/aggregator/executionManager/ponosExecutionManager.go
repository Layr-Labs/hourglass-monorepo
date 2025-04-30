package executionManager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/common/v1"
	aggregatorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/fauxSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

const (
	DefaultRefreshInterval = 30 * time.Second
)

type PonosExecutionManagerConfig struct {
	RefreshInterval  time.Duration
	SecureConnection bool
}

func NewPonosExecutionManagerDefaultConfig() *PonosExecutionManagerConfig {
	return &PonosExecutionManagerConfig{
		RefreshInterval:  DefaultRefreshInterval,
		SecureConnection: false,
	}
}

type PonosExecutionManager struct {
	rpcServer          *rpcServer.RpcServer
	taskQueue          chan *types.Task
	resultQueue        chan *types.TaskResult
	execClients        map[string]executorClient.IExecutorClient
	peeringDataFetcher peering.IPeeringDataFetcher
	running            sync.Map
	config             *PonosExecutionManagerConfig
	logger             *zap.Logger
}

func NewPonosExecutionManagerWithRpcServer(
	taskQueue chan *types.Task,
	resultQueue chan *types.TaskResult,
	peeringDataFetcher peering.IPeeringDataFetcher,
	config *PonosExecutionManagerConfig,
	rpcServerPort int,
	logger *zap.Logger,
) (*PonosExecutionManager, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: rpcServerPort,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %w", err)
	}

	return NewPonosExecutionManager(
		rpc,
		taskQueue,
		resultQueue,
		peeringDataFetcher,
		config,
		logger,
	), nil
}

func NewPonosExecutionManager(
	server *rpcServer.RpcServer,
	taskQueue chan *types.Task,
	resultQueue chan *types.TaskResult,
	peeringDataFetcher peering.IPeeringDataFetcher,
	config *PonosExecutionManagerConfig,
	logger *zap.Logger,
) *PonosExecutionManager {
	manager := &PonosExecutionManager{
		rpcServer:          server,
		taskQueue:          taskQueue,
		resultQueue:        resultQueue,
		peeringDataFetcher: peeringDataFetcher,
		execClients:        map[string]executorClient.IExecutorClient{},
		config:             config,
		logger:             logger,
	}
	aggregatorpb.RegisterAggregatorServiceServer(server.GetGrpcServer(), manager)
	return manager
}

func (em *PonosExecutionManager) Start(ctx context.Context) error {
	em.logger.Sugar().Infow("Starting PonosExecutionManager",
		"secureConnection", em.config.SecureConnection,
		"refreshInterval", em.config.RefreshInterval,
	)

	err := em.rpcServer.Start(ctx)
	if err != nil {
		return err
	}

	go em.processTaskQueue(ctx)
	go em.refreshExecutorClientsLoop(ctx)

	return nil
}

func (em *PonosExecutionManager) refreshExecutorClientsLoop(ctx context.Context) {
	ticker := time.NewTicker(em.config.RefreshInterval)
	defer ticker.Stop()

	sugar := em.logger.Sugar()
	sugar.Info("Starting executor client refresh loop")

	em.refreshExecutorClients()

	for {
		select {
		case <-ctx.Done():
			sugar.Info("Stopping executor client refresh loop")
			return
		case <-ticker.C:
			em.refreshExecutorClients()
		}
	}
}

func (em *PonosExecutionManager) processTaskQueue(ctx context.Context) {
	sugar := em.logger.Sugar()
	sugar.Info("Starting task processing loop")

	for {
		select {
		case <-ctx.Done():
			sugar.Info("Stopping task processing loop")
			return
		case task, ok := <-em.taskQueue:
			if !ok {
				sugar.Warn("Task queue channel closed, exiting")
				return
			}

			go em.processTask(ctx, task)
		}
	}
}

func (em *PonosExecutionManager) processTask(ctx context.Context, task *types.Task) {
	sugar := em.logger.Sugar()
	sugar.Infow("Processing task", "taskId", task.TaskId)

	em.running.Store(task.TaskId, task)
	clientCount := 0
	for addr, execClient := range em.execClients {
		clientCount++

		go func(address string, client executorClient.IExecutorClient) {
			fmt.Printf("TASK: %+v\n", task)
			err := client.SubmitTask(ctx, task)
			if err != nil {
				sugar.Errorw("Failed to submit task to executor",
					"executor_address", address,
					"task_id", task.TaskId,
					"error", err,
				)
			} else {
				sugar.Debugw("Successfully submitted task to executor",
					"executor_address", address,
					"task_id", task.TaskId,
				)
			}
		}(addr, execClient)
	}
}

func (em *PonosExecutionManager) SubmitTaskResult(
	ctx context.Context,
	result *aggregatorpb.TaskResult,
) (*v1.SubmitAck, error) {
	sugar := em.logger.Sugar()
	taskID := result.TaskId

	value, ok := em.running.Load(taskID)
	if !ok {
		sugar.Warnw("Received result for unknown task", "task_id", taskID)
		return &v1.SubmitAck{Success: false, Message: "unknown task"}, nil
	}

	em.running.Delete(taskID)
	task := value.(*types.Task)

	taskResult := &types.TaskResult{
		TaskId:        task.TaskId,
		AvsAddress:    task.AVSAddress,
		CallbackAddr:  task.CallbackAddr,
		OperatorSetId: task.OperatorSetId,
		Output:        result.Output,
		ChainId:       task.ChainId,
		BlockNumber:   task.BlockNumber,
		BlockHash:     task.BlockHash,
	}

	select {
	case em.resultQueue <- taskResult:
		sugar.Infow("Task result accepted", "task_id", taskID)
		return &v1.SubmitAck{Success: true, Message: "ok"}, nil
	case <-time.After(1 * time.Second):
		sugar.Errorw("Failed to enqueue task result (channel full or closed)", "task_id", taskID)
		return &v1.SubmitAck{Success: false, Message: "enqueue error"}, nil
	case <-ctx.Done():
		sugar.Warnw("Context cancelled while enqueueing result", "task_id", taskID)
		return &v1.SubmitAck{Success: false, Message: "context cancelled"}, nil
	}
}

func (em *PonosExecutionManager) refreshExecutorClients() {
	sugar := em.logger.Sugar()

	peers, err := em.peeringDataFetcher.ListExecutorOperators()
	if err != nil {
		sugar.Errorw("Failed to list executor peers", "error", err)
		return
	}

	newClientCount := 0

	for _, peer := range peers {
		if _, exists := em.execClients[peer.PublicKey]; !exists {
			client, err := em.loadExecutorClient(peer, em.config.SecureConnection)
			if err != nil {
				// TODO: emit metric
				sugar.Errorw("Failed to create executor client",
					"public_key", peer.PublicKey,
					"error", err,
				)
				continue
			}
			em.execClients[peer.PublicKey] = client
			newClientCount++
			// TODO: emit metric
			sugar.Infow("Registered new executor client", "public_key", peer.PublicKey)
		}
	}

	if newClientCount > 0 {
		sugar.Infow("Refreshed executor clients", "newClients", newClientCount, "totalClients", len(em.execClients))
	}
}

func (em *PonosExecutionManager) loadExecutorClient(
	peer *peering.OperatorPeerInfo,
	secureConnection bool,
) (executorClient.IExecutorClient, error) {
	conn, err := clients.NewGrpcClient(
		fmt.Sprintf("%s:%d", peer.NetworkAddress, peer.Port),
		secureConnection,
	)
	if err != nil {
		return nil, err
	}
	client := executorpb.NewExecutorServiceClient(conn)
	// TODO: replace with a real signer.
	return executorClient.NewPonosExecutorClient(client, fauxSigner.NewFauxSigner()), nil
}
