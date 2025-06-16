package taskSession

import (
	"context"
	"errors"
	"fmt"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
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

	operatorTableData *contractCaller.OperatorTableData

	taskAggregator *aggregation.TaskResultAggregator
	thresholdMet   atomic.Bool
}

func NewTaskSession(
	ctx context.Context,
	cancel context.CancelFunc,
	task *types.Task,
	aggregatorAddress string,
	aggregatorSignature []byte,
	operatorTableData *contractCaller.OperatorTableData,
	logger *zap.Logger,
) (*TaskSession, error) {
	operators := []*aggregation.Operator{}
	for _, peer := range task.RecipientOperators {
		opset, err := peer.GetOperatorSet(task.OperatorSetId)
		if err != nil {
			return nil, fmt.Errorf("failed to get operator set %d for peer %s: %w", task.OperatorSetId, peer.OperatorAddress, err)
		}
		operators = append(operators, &aggregation.Operator{
			Address:   peer.OperatorAddress,
			PublicKey: opset.PublicKey,
		})
	}

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
		aggregatorSignature: aggregatorSignature,
		results:             sync.Map{},
		context:             ctx,
		contextCancel:       cancel,
		logger:              logger,
		taskAggregator:      ta,
		operatorTableData:   operatorTableData,
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
	errChan := make(chan error, 1)

	go func() {
		cert, err := ts.Broadcast()
		if err != nil {
			ts.logger.Sugar().Errorw("task session broadcast failed",
				zap.String("taskId", ts.Task.TaskId),
				zap.Error(err),
			)
			errChan <- err
			return
		}
		ts.logger.Sugar().Infow("task session broadcast completed",
			zap.String("taskId", ts.Task.TaskId),
			zap.Any("cert", cert),
		)
		certChan <- cert
	}()

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
		zap.Any("operatorPeers", ts.Task.RecipientOperators),
	)

	resultsChan := make(chan *types.TaskResult)

	for _, peer := range ts.Task.RecipientOperators {
		go func(peer *peering.OperatorPeerInfo) {
			socket, err := peer.GetSocketForOperatorSet(ts.Task.OperatorSetId)
			if err != nil {
				ts.logger.Sugar().Errorw("Failed to get socket for operator set",
					zap.String("taskId", ts.Task.TaskId),
					zap.String("operatorAddress", peer.OperatorAddress),
					zap.Error(err),
				)
				return
			}
			c, err := executorClient.NewExecutorClient(socket, true)
			if err != nil {
				ts.logger.Sugar().Errorw("Failed to create executor client",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("taskId", ts.Task.TaskId),
					zap.Error(err),
				)
				return
			}

			ts.logger.Sugar().Infow("broadcasting task to operator",
				zap.String("taskId", ts.Task.TaskId),
				zap.String("operatorAddress", peer.OperatorAddress),
				zap.String("networkAddress", socket),
			)
			res, err := c.SubmitTask(ts.context, taskSubmission)
			if err != nil {
				ts.logger.Sugar().Errorw("Failed to submit task to executor",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("taskId", ts.Task.TaskId),
					zap.Error(err),
				)
				return
			}
			ts.logger.Sugar().Infow("received task result from executor",
				zap.String("taskId", ts.Task.TaskId),
				zap.String("operatorAddress", peer.OperatorAddress),
				zap.Any("result", res),
			)
			resultsChan <- types.TaskResultFromTaskResultProto(res)
		}(peer)
	}

	// iterate over results until we meet the signing threshold
	for taskResult := range resultsChan {
		ts.logger.Sugar().Infow("received task result on channel",
			zap.String("taskId", taskResult.TaskId),
			zap.String("operatorAddress", taskResult.OperatorAddress),
		)
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
			continue
		}
		ts.logger.Sugar().Infow("task result processed, checking signing threshold",
			zap.String("taskId", taskResult.TaskId),
			zap.String("operatorAddress", taskResult.OperatorAddress),
		)

		if !ts.taskAggregator.SigningThresholdMet() {
			ts.logger.Sugar().Infow("task completion threshold not met yet",
				zap.String("taskId", taskResult.TaskId),
				zap.String("operatorAddress", taskResult.OperatorAddress),
			)
			continue
		}
		ts.thresholdMet.Store(true)

		// threshold met, close the results channel to stop further processing
		close(resultsChan)

		ts.logger.Sugar().Infow("task completion threshold met, generating final certificate",
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
