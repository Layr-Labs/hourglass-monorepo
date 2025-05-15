package taskSession

import (
	"context"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
)

type TaskSession struct {
	Task                *types.Task
	aggregatorSignature []byte
	context             context.Context
	contextCancel       context.CancelFunc
	logger              *zap.Logger
	results             sync.Map
	resultsCount        atomic.Uint32
	aggregatorAddress   string
	aggregatorUrl       string

	taskAggregator       *aggregation.TaskResultAggregator
	resultsQueue         chan *TaskSession
	thresholdMet         atomic.Bool
	AggregateCertificate *aggregation.AggregatedCertificate
}

func NewTaskSession(
	ctx context.Context,
	cancel context.CancelFunc,
	task *types.Task,
	aggregatorAddress string,
	aggregatorUrl string,
	aggregatorSignature []byte,
	resultsQueue chan *TaskSession,
	logger *zap.Logger,
) (*TaskSession, error) {
	operators := util.Map(task.RecipientOperators, func(peer *peering.OperatorPeerInfo, i uint64) *aggregation.Operator {
		return &aggregation.Operator{
			Address:   peer.OperatorAddress,
			PublicKey: peer.PublicKey,
		}
	})

	ta, err := aggregation.NewTaskResultAggregator(
		ctx,
		task.TaskId,
		task.BlockNumber,
		task.OperatorSetId,
		100,
		task.Payload,
		task.DeadlineUnixSeconds,
		operators,
	)
	if err != nil {
		return nil, err
	}
	ts := &TaskSession{
		Task:                task,
		aggregatorAddress:   aggregatorAddress,
		aggregatorUrl:       aggregatorUrl,
		aggregatorSignature: aggregatorSignature,
		results:             sync.Map{},
		context:             ctx,
		contextCancel:       cancel,
		logger:              logger,
		resultsQueue:        resultsQueue,
		taskAggregator:      ta,
		thresholdMet:        atomic.Bool{},
	}
	ts.resultsCount.Store(0)
	ts.thresholdMet.Store(false)

	return ts, nil
}

func (ts *TaskSession) Process() error {
	ts.logger.Sugar().Infow("task session started",
		zap.String("taskId", ts.Task.TaskId),
	)
	go ts.Broadcast()

	<-ts.context.Done()
	ts.logger.Sugar().Infow("task session context done",
		zap.String("taskId", ts.Task.TaskId),
	)
	return nil
}

func (ts *TaskSession) Broadcast() {
	ts.logger.Sugar().Infow("task session broadcast started",
		zap.String("taskId", ts.Task.TaskId),
		zap.Any("recipientOperators", ts.Task.RecipientOperators),
	)
	taskSubmission := &executorV1.TaskSubmission{
		TaskId:            ts.Task.TaskId,
		AvsAddress:        ts.Task.AVSAddress,
		AggregatorAddress: ts.aggregatorAddress,
		Payload:           ts.Task.Payload,
		AggregatorUrl:     ts.aggregatorUrl,
		Signature:         ts.aggregatorSignature,
	}
	ts.logger.Sugar().Infow("broadcasting task session to operators",
		zap.Any("taskSubmission", taskSubmission),
	)

	var wg sync.WaitGroup
	for _, peer := range ts.Task.RecipientOperators {
		wg.Add(1)

		go func(wg *sync.WaitGroup, peer *peering.OperatorPeerInfo) {
			defer wg.Done()
			ts.logger.Sugar().Infow("task session broadcast to operator",
				zap.String("taskId", ts.Task.TaskId),
				zap.String("operatorAddress", peer.OperatorAddress),
				zap.String("networkAddress", peer.NetworkAddress),
			)
			c, err := executorClient.NewExecutorClient(peer.NetworkAddress, true)
			if err != nil {
				ts.logger.Sugar().Errorw("Failed to create executor client",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("taskId", ts.Task.TaskId),
					zap.Error(err),
				)
				return
			}

			res, err := c.SubmitTask(ts.context, taskSubmission)
			if err != nil {
				ts.logger.Sugar().Errorw("Failed to submit task to executor",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("taskId", ts.Task.TaskId),
					zap.Error(err),
				)
				return
			}
			if !res.Success {
				ts.logger.Sugar().Errorw("task submission failed",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("taskId", ts.Task.TaskId),
					zap.String("message", res.Message),
				)
				return
			}
			ts.logger.Sugar().Debugw("Successfully submitted task to executor",
				zap.String("executorAddress", peer.OperatorAddress),
				zap.String("taskId", ts.Task.TaskId),
			)
		}(&wg, peer)
	}
	wg.Wait()
	ts.logger.Sugar().Infow("task submission completed",
		zap.String("taskId", ts.Task.TaskId),
	)
}

func (ts *TaskSession) RecordResult(taskResult *types.TaskResult) {
	if ts.thresholdMet.Load() {
		ts.logger.Sugar().Infow("task completion threshold already met",
			zap.String("taskId", taskResult.TaskId),
			zap.String("operatorAddress", taskResult.OperatorAddress),
		)
		return
	}
	if err := ts.taskAggregator.ProcessNewSignature(ts.context, taskResult.TaskId, taskResult); err != nil {
		ts.logger.Sugar().Errorw("Failed to process task result",
			zap.String("taskId", taskResult.TaskId),
			zap.String("operatorAddress", taskResult.OperatorAddress),
			zap.Error(err),
		)
	}

	if !ts.taskAggregator.SigningThresholdMet() {
		return
	}
	ts.thresholdMet.Store(true)
	ts.logger.Sugar().Infow("task completion threshold met",
		zap.String("taskId", taskResult.TaskId),
		zap.String("operatorAddress", taskResult.OperatorAddress),
	)

	cert, err := ts.taskAggregator.GenerateFinalCertificate()
	if err != nil {
		ts.logger.Sugar().Errorw("Failed to generate final certificate",
			zap.String("taskId", taskResult.TaskId),
			zap.String("operatorAddress", taskResult.OperatorAddress),
			zap.Error(err),
		)
		return
	}
	ts.AggregateCertificate = cert

	ts.resultsQueue <- ts
}

func (ts *TaskSession) GetOperatorOutputsMap() map[string][]byte {
	operatorOutputs := make(map[string][]byte)
	ts.results.Range(func(_, value any) bool {
		result := value.(*types.TaskResult)
		operatorOutputs[result.OperatorAddress] = result.Output
		return true
	})
	return operatorOutputs
}

func (ts *TaskSession) GetTaskResults() []*types.TaskResult {
	results := make([]*types.TaskResult, 0)
	ts.results.Range(func(_, value any) bool {
		result := value.(*types.TaskResult)
		results = append(results, result)
		return true
	})
	return results
}
