package executionManager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/common/v1"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/tasks"
	"go.uber.org/zap"
)

const (
	DefaultRefreshInterval = 30 * time.Second
)

type PonosExecutionManagerConfig struct {
	PeerRefreshInterval       time.Duration
	SecureConnection          bool
	AggregatorOperatorAddress string
	AggregatorUrl             string
}

func NewPonosExecutionManagerDefaultConfig() *PonosExecutionManagerConfig {
	return &PonosExecutionManagerConfig{
		PeerRefreshInterval: DefaultRefreshInterval,
		SecureConnection:    false,
	}
}

type PonosExecutionManager struct {
	rpcServer          *rpcServer.RpcServer
	taskQueue          chan *tasks.Task
	resultQueue        chan *tasks.TaskResult
	execClients        map[string]executorV1.ExecutorServiceClient
	peeringDataFetcher peering.IPeeringDataFetcher
	running            sync.Map
	config             *PonosExecutionManagerConfig
	signer             signer.Signer
	logger             *zap.Logger
}

func NewPonosExecutionManagerWithRpcServer(
	taskQueue chan *tasks.Task,
	resultQueue chan *tasks.TaskResult,
	peeringDataFetcher peering.IPeeringDataFetcher,
	config *PonosExecutionManagerConfig,
	rpcServerPort int,
	signer signer.Signer,
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
		signer,
		logger,
	), nil
}

func NewPonosExecutionManager(
	server *rpcServer.RpcServer,
	taskQueue chan *tasks.Task,
	resultQueue chan *tasks.TaskResult,
	peeringDataFetcher peering.IPeeringDataFetcher,
	config *PonosExecutionManagerConfig,
	signer signer.Signer,
	logger *zap.Logger,
) *PonosExecutionManager {
	manager := &PonosExecutionManager{
		rpcServer:          server,
		taskQueue:          taskQueue,
		resultQueue:        resultQueue,
		peeringDataFetcher: peeringDataFetcher,
		execClients:        map[string]executorV1.ExecutorServiceClient{},
		config:             config,
		signer:             signer,
		logger:             logger,
	}
	aggregatorV1.RegisterAggregatorServiceServer(server.GetGrpcServer(), manager)
	return manager
}

func (em *PonosExecutionManager) Start(ctx context.Context) error {
	em.logger.Sugar().Infow("Starting PonosExecutionManager",
		"secureConnection", em.config.SecureConnection,
		"refreshInterval", em.config.PeerRefreshInterval,
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
	ticker := time.NewTicker(em.config.PeerRefreshInterval)
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

func (em *PonosExecutionManager) processTask(ctx context.Context, task *tasks.Task) {
	sugar := em.logger.Sugar()
	sugar.Infow("Processing task", "taskId", task.TaskId)
	em.running.Store(task.TaskId, task)

	sig, err := em.signer.SignMessage(task.Payload)
	if err != nil {
		sugar.Errorw("Failed to sign task payload",
			zap.String("taskId", task.TaskId),
			zap.Error(err),
		)
		return
	}

	aggregatorUrl := fmt.Sprintf("localhost:%d", em.rpcServer.RpcConfig.GrpcPort)
	if em.config.AggregatorUrl != "" {
		sugar.Infow("Using custom aggregator URL",
			zap.String("aggregatorUrl", em.config.AggregatorUrl),
		)
		aggregatorUrl = em.config.AggregatorUrl
	}

	taskSubmission := &executorV1.TaskSubmission{
		TaskId:            task.TaskId,
		AvsAddress:        task.AVSAddress,
		AggregatorAddress: em.config.AggregatorOperatorAddress,
		Payload:           task.Payload,
		AggregatorUrl:     aggregatorUrl,
		Signature:         sig,
	}

	var wg sync.WaitGroup
	for addr, execClient := range em.execClients {
		wg.Add(1)

		go func(address string, client executorV1.ExecutorServiceClient, wg *sync.WaitGroup) {
			defer wg.Done()
			fmt.Printf("Submitting task: %+v\n", taskSubmission)
			res, err := client.SubmitTask(ctx, taskSubmission)
			if err != nil {
				sugar.Errorw("Failed to submit task to executor",
					zap.String("executorAddress", address),
					zap.String("taskId", task.TaskId),
					zap.Error(err),
				)
				return
			}
			if !res.Success {
				sugar.Errorw("Task submission failed",
					zap.String("executorAddress", address),
					zap.String("taskId", task.TaskId),
					zap.String("message", res.Message),
				)
				return
			}
			sugar.Debugw("Successfully submitted task to executor",
				zap.String("executorAddress", address),
				zap.String("taskId", task.TaskId),
			)

		}(addr, execClient, &wg)
	}
	wg.Wait()
	sugar.Infow("Task submission completed", zap.String("taskId", task.TaskId))
}

func (em *PonosExecutionManager) SubmitTaskResult(
	ctx context.Context,
	result *aggregatorV1.TaskResult,
) (*v1.SubmitAck, error) {
	taskID := result.TaskId

	value, ok := em.running.Load(taskID)
	if !ok {
		em.logger.Sugar().Warnw("Received result for unknown task", "task_id", taskID)
		return &v1.SubmitAck{Success: false, Message: "unknown task"}, nil
	}

	em.running.Delete(taskID)
	task := value.(*tasks.Task)

	taskResult := &tasks.TaskResult{
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
		em.logger.Sugar().Infow("Task result accepted", "task_id", taskID)
		return &v1.SubmitAck{Success: true, Message: "ok"}, nil
	case <-time.After(1 * time.Second):
		em.logger.Sugar().Errorw("Failed to enqueue task result (channel full or closed)", "task_id", taskID)
		return &v1.SubmitAck{Success: false, Message: "enqueue error"}, nil
	case <-ctx.Done():
		em.logger.Sugar().Warnw("Context cancelled while enqueueing result", "task_id", taskID)
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
			addr := fmt.Sprintf("%s:%d", peer.NetworkAddress, peer.Port)

			// TODO - SecureConnection should always be used unless the address contains 'localhost' or '127.0.0.1'
			client, err := executorClient.NewExecutorClient(addr, !em.config.SecureConnection)
			if err != nil {
				// TODO: emit metric
				sugar.Errorw("Failed to create executor client",
					zap.String("address", addr),
					zap.String("publicKey", peer.PublicKey),
					zap.String("operatorAddress", peer.OperatorAddress),
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
