package taskSession

import (
	"context"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
	"strings"
	"sync"
	"sync/atomic"
)

type TaskSession struct {
	task                *types.Task
	aggregatorSignature []byte
	recipientOperators  []*peering.OperatorPeerInfo
	context             context.Context
	contextCancel       context.CancelFunc
	logger              *zap.Logger
	results             sync.Map
	resultsCount        atomic.Uint32
	aggregatorAddress   string
	aggregatorUrl       string

	resultsQueue chan *TaskSession
}

func NewTaskSession(
	ctx context.Context,
	cancel context.CancelFunc,
	task *types.Task,
	aggregatorAddress string,
	aggregatorUrl string,
	aggregatorSignature []byte,
	recipientOperators []*peering.OperatorPeerInfo,
	resultsQueue chan *TaskSession,
	logger *zap.Logger,
) *TaskSession {
	ts := &TaskSession{
		task:                task,
		aggregatorAddress:   aggregatorAddress,
		aggregatorUrl:       aggregatorUrl,
		aggregatorSignature: aggregatorSignature,
		recipientOperators:  recipientOperators,
		results:             sync.Map{},
		context:             ctx,
		contextCancel:       cancel,
		logger:              logger,
		resultsQueue:        resultsQueue,
	}
	ts.resultsCount.Store(0)
	return ts
}

func (ts *TaskSession) Process() error {
	ts.logger.Sugar().Infow("task session started",
		zap.String("taskId", ts.task.TaskId),
	)
	go ts.Broadcast()

	<-ts.context.Done()
	ts.logger.Sugar().Infow("task session context done",
		zap.String("taskId", ts.task.TaskId),
	)
	return nil
}

func (ts *TaskSession) Broadcast() {
	ts.logger.Sugar().Infow("task session broadcast started",
		zap.String("taskId", ts.task.TaskId),
		zap.Any("recipientOperators", ts.recipientOperators),
	)
	taskSubmission := &executorV1.TaskSubmission{
		TaskId:            ts.task.TaskId,
		AvsAddress:        ts.task.AVSAddress,
		AggregatorAddress: ts.aggregatorAddress,
		Payload:           ts.task.Payload,
		AggregatorUrl:     ts.aggregatorUrl,
		Signature:         ts.aggregatorSignature,
	}
	ts.logger.Sugar().Infow("task session broadcast to operators",
		zap.Any("taskSubmission", taskSubmission),
	)

	var wg sync.WaitGroup
	for _, peer := range ts.recipientOperators {
		wg.Add(1)

		go func(wg *sync.WaitGroup, peer *peering.OperatorPeerInfo) {
			defer wg.Done()
			ts.logger.Sugar().Infow("task session broadcast to operator",
				zap.String("taskId", ts.task.TaskId),
				zap.String("operatorAddress", peer.OperatorAddress),
				zap.String("networkAddress", peer.NetworkAddress),
			)
			c, err := executorClient.NewExecutorClient(peer.NetworkAddress, true)
			if err != nil {
				ts.logger.Sugar().Errorw("Failed to create executor client",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("taskId", ts.task.TaskId),
					zap.Error(err),
				)
				return
			}

			res, err := c.SubmitTask(ts.context, taskSubmission)
			if err != nil {
				ts.logger.Sugar().Errorw("Failed to submit task to executor",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("taskId", ts.task.TaskId),
					zap.Error(err),
				)
				return
			}
			if !res.Success {
				ts.logger.Sugar().Errorw("task submission failed",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("taskId", ts.task.TaskId),
					zap.String("message", res.Message),
				)
				return
			}
			ts.logger.Sugar().Debugw("Successfully submitted task to executor",
				zap.String("executorAddress", peer.OperatorAddress),
				zap.String("taskId", ts.task.TaskId),
			)
		}(&wg, peer)
	}
	wg.Wait()
	ts.logger.Sugar().Infow("task submission completed",
		zap.String("taskId", ts.task.TaskId),
	)
}

func (ts *TaskSession) findOperatorByAddress(address string) *peering.OperatorPeerInfo {
	for _, peer := range ts.recipientOperators {
		if strings.EqualFold(peer.OperatorAddress, address) {
			return peer
		}
	}
	return nil
}

func (ts *TaskSession) RecordResult(taskResult *types.TaskResult) {
	peer := ts.findOperatorByAddress(taskResult.OperatorAddress)
	if peer == nil {
		ts.logger.Sugar().Errorw("Failed to find operator by address",
			"address", taskResult.OperatorAddress,
			"taskId", taskResult.TaskId,
		)
		return
	}

	if _, ok := ts.results.Load(peer); ok {
		ts.logger.Sugar().Errorw("Duplicate result for task",
			"taskId", taskResult.TaskId,
			"operatorAddress", taskResult.OperatorAddress,
		)
		return
	}
	ts.results.Store(peer, taskResult)
	ts.resultsCount.Add(1)

	if ts.resultsCount.Load() == uint32(len(ts.recipientOperators)) {
		ts.resultsQueue <- ts
		ts.logger.Sugar().Infow("Task result published to channel",
			"taskId", ts.task.TaskId,
		)
	}
}
