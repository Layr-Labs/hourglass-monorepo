package executionManager

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/common/v1"
	aggregatorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/workQueue"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/fauxSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
)

type PonosExecutionManager struct {
	aggregatorServer   *rpcServer.RpcServer
	taskQueue          workQueue.IOutputQueue[types.Task]
	resultQueue        workQueue.IInputQueue[types.TaskResult]
	peeringDataFetcher *peering.LocalPeeringDataFetcher
	execClients        map[string]executorClient.IExecutorClient
	running            sync.Map
	wg                 sync.WaitGroup

	cancel context.CancelFunc
	logger *zap.Logger
}

func NewPonosExecutionManager(
	server *rpcServer.RpcServer,
	taskQueue workQueue.IOutputQueue[types.Task],
	resultQueue workQueue.IInputQueue[types.TaskResult],
	peeringDataFetcher *peering.LocalPeeringDataFetcher,
	logger *zap.Logger,
) *PonosExecutionManager {
	manager := &PonosExecutionManager{
		aggregatorServer:   server,
		taskQueue:          taskQueue,
		resultQueue:        resultQueue,
		peeringDataFetcher: peeringDataFetcher,
		execClients:        map[string]executorClient.IExecutorClient{},
		logger:             logger,
	}
	aggregatorpb.RegisterAggregatorServiceServer(server.GetGrpcServer(), manager)
	return manager
}

func (em *PonosExecutionManager) Start(ctx context.Context) error {
	ctx, em.cancel = context.WithCancel(ctx)
	err := em.aggregatorServer.Start(ctx)
	if err != nil {
		return err
	}
	go em.run(ctx)
	return nil
}

func (em *PonosExecutionManager) Close() error {
	if em.cancel != nil {
		em.cancel()
	}
	em.logger.Info("Waiting for all execution results to arrive...")
	em.wg.Wait()
	em.logger.Info("Shutdown complete")
	return nil
}

func (em *PonosExecutionManager) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			em.logger.Info("ExecutionManager shutting down")
			return
		default:
			em.refreshExecutorClients()

			task := em.taskQueue.Dequeue()
			em.wg.Add(1)
			go func() {
				em.running.Store(task.TaskId, task)
				for addr, execClient := range em.execClients {
					err := execClient.SubmitTask(context.Background(), task)
					if err != nil {
						em.logger.Error("Failed to submit task to executor", zap.String("executor_address", addr), zap.String("task_id", task.TaskId), zap.Error(err))
						em.running.Delete(task.TaskId)
						em.wg.Done()
					}
				}
			}()

		}
	}
}

func (em *PonosExecutionManager) SubmitTaskResult(
	_ context.Context,
	result *aggregatorpb.TaskResult,
) (*v1.SubmitAck, error) {
	taskID := result.TaskId
	value, ok := em.running.Load(taskID)
	if !ok {
		em.logger.Warn("Received result for unknown task", zap.String("task_id", taskID))
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

	if err := em.resultQueue.Enqueue(taskResult); err != nil {
		em.logger.Error("Failed to enqueue task result", zap.String("task_id", taskID), zap.Error(err))
		return &v1.SubmitAck{Success: false, Message: "enqueue error"}, nil
	}

	em.logger.Info("Task result accepted", zap.String("task_id", taskID))
	em.wg.Done()
	return &v1.SubmitAck{Success: true, Message: "ok"}, nil
}

func (em *PonosExecutionManager) refreshExecutorClients() {
	peers, err := em.peeringDataFetcher.ListExecutorOperators()
	if err != nil {
		em.logger.Error("Failed to list executor peers", zap.Error(err))
		return
	}
	for _, peer := range peers {
		if _, exists := em.execClients[peer.PublicKey]; !exists {
			client, err := em.loadExecutorClient(peer)
			if err != nil {
				// TODO: emit metric
				em.logger.Error("Failed to create executor client", zap.String("public_key", peer.PublicKey), zap.Error(err))
				continue
			}
			em.execClients[peer.PublicKey] = client
			// TODO: emit metric
			em.logger.Info("Registered new executor client", zap.String("public_key", peer.PublicKey))
		}
	}
}

func (em *PonosExecutionManager) loadExecutorClient(peer *peering.ExecutorOperatorPeerInfo) (executorClient.IExecutorClient, error) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", peer.NetworkAddress, peer.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	client := executorpb.NewExecutorServiceClient(conn)
	// TODO: replace with a real signer.
	return executorClient.NewPonosExecutorClient(client, fauxSigner.NewFauxSigner()), nil
}
