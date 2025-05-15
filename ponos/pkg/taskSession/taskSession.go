package taskSession

import (
	"context"
	"errors"
	"fmt"
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

	taskAggregator *aggregation.TaskResultAggregator
	thresholdMet   atomic.Bool
}

func NewTaskSession(
	ctx context.Context,
	cancel context.CancelFunc,
	task *types.Task,
	aggregatorAddress string,
	aggregatorUrl string,
	aggregatorSignature []byte,
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
		taskAggregator:      ta,
		thresholdMet:        atomic.Bool{},
	}
	ts.resultsCount.Store(0)
	ts.thresholdMet.Store(false)

	return ts, nil
}

func (ts *TaskSession) Process() (*aggregation.AggregatedCertificate, error) {
	ts.logger.Sugar().Infow("task session started",
		zap.String("taskId", ts.Task.TaskId),
	)

	certChan := make(chan *aggregation.AggregatedCertificate, 1)

	select {
	case cert := <-certChan:
		return cert, nil
	case <-ts.context.Done():
		if errors.Is(ts.context.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("task session context deadline exceeded: %w", ts.context.Err())
		}
		return nil, fmt.Errorf("task session context done: %w", ts.context.Err())
	}
}

func (ts *TaskSession) Broadcast() (*aggregation.AggregatedCertificate, error) {
	ts.logger.Sugar().Infow("task session broadcast started",
		zap.String("taskId", ts.Task.TaskId),
		zap.Any("recipientOperators", ts.Task.RecipientOperators),
	)
	taskSubmission := &executorV1.TaskSubmission{
		TaskId:            ts.Task.TaskId,
		AvsAddress:        ts.Task.AVSAddress,
		AggregatorAddress: ts.aggregatorAddress,
		Payload:           ts.Task.Payload,
		Signature:         ts.aggregatorSignature,
	}
	ts.logger.Sugar().Infow("broadcasting task session to operators",
		zap.Any("taskSubmission", taskSubmission),
	)

	resultsChan := make(chan *types.TaskResult)

	for _, peer := range ts.Task.RecipientOperators {
		go func(peer *peering.OperatorPeerInfo) {
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
			resultsChan <- types.TaskResultFromTaskResultProto(res)
		}(peer)
	}

	// iterate over results until we meet the signing threshold
	for taskResult := range resultsChan {
		if taskResult == nil {
			ts.logger.Sugar().Errorw("task result is nil",
				zap.String("taskId", ts.Task.TaskId),
			)
			continue
		}
		if ts.thresholdMet.Load() {
			ts.logger.Sugar().Infow("task completion threshold already met",
				zap.String("taskId", taskResult.TaskId),
				zap.String("operatorAddress", taskResult.OperatorAddress),
			)
			continue
		}
		if err := ts.taskAggregator.ProcessNewSignature(ts.context, taskResult.TaskId, taskResult); err != nil {
			ts.logger.Sugar().Errorw("Failed to process task result",
				zap.String("taskId", taskResult.TaskId),
				zap.String("operatorAddress", taskResult.OperatorAddress),
				zap.Error(err),
			)
		}

		if !ts.taskAggregator.SigningThresholdMet() {
			continue
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
			return nil, fmt.Errorf("failed to generate final certificate: %w", err)
		}
		return cert, nil
	}

	return nil, fmt.Errorf("failed to meet signing threshold")
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
