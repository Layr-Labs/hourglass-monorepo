package taskSession

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const maximumTaskResponseSize = 1.5 * 1024 * 1024

type TaskSession[SigT, CertT, PubKeyT any] struct {
	Task                *types.Task
	aggregatorSignature []byte
	context             context.Context
	contextCancel       context.CancelFunc
	logger              *zap.Logger
	results             sync.Map
	resultsCount        atomic.Uint32
	aggregatorAddress   string

	operatorPeersWeight *operatorManager.PeerWeight

	taskAggregator aggregation.ITaskResultAggregator[SigT, CertT, PubKeyT]
	thresholdMet   atomic.Bool
}

func NewBN254TaskSession(
	ctx context.Context,
	cancel context.CancelFunc,
	task *types.Task,
	aggregatorAddress string,
	aggregatorSignature []byte,
	operatorPeersWeight *operatorManager.PeerWeight,
	logger *zap.Logger,
) (*TaskSession[bn254.Signature, aggregation.AggregatedBN254Certificate, signing.PublicKey], error) {
	operators := []*aggregation.Operator[signing.PublicKey]{}
	for _, peer := range operatorPeersWeight.Operators {
		opset, err := peer.GetOperatorSet(task.OperatorSetId)
		if err != nil {
			return nil, fmt.Errorf("failed to get operator set %d for peer %s: %w", task.OperatorSetId, peer.OperatorAddress, err)
		}
		operators = append(operators, &aggregation.Operator[signing.PublicKey]{
			Address:   peer.OperatorAddress,
			PublicKey: opset.WrappedPublicKey.PublicKey,
		})
	}

	ta, err := aggregation.NewBN254TaskResultAggregator(
		ctx,
		task.TaskId,
		task.BlockNumber,
		task.OperatorSetId,
		task.ThresholdBips,
		task.Payload,
		task.DeadlineUnixSeconds,
		operators,
	)
	if err != nil {
		return nil, err
	}
	ts := &TaskSession[bn254.Signature, aggregation.AggregatedBN254Certificate, signing.PublicKey]{
		Task:                task,
		aggregatorAddress:   aggregatorAddress,
		aggregatorSignature: aggregatorSignature,
		results:             sync.Map{},
		context:             ctx,
		contextCancel:       cancel,
		logger:              logger,
		taskAggregator:      ta,
		operatorPeersWeight: operatorPeersWeight,
		thresholdMet:        atomic.Bool{},
	}
	ts.resultsCount.Store(0)
	ts.thresholdMet.Store(false)

	return ts, nil
}

func NewECDSATaskSession(
	ctx context.Context,
	cancel context.CancelFunc,
	task *types.Task,
	aggregatorAddress string,
	aggregatorSignature []byte,
	operatorPeersWeight *operatorManager.PeerWeight,
	logger *zap.Logger,
) (*TaskSession[ecdsa.Signature, aggregation.AggregatedECDSACertificate, common.Address], error) {
	operators := []*aggregation.Operator[common.Address]{}
	for _, peer := range operatorPeersWeight.Operators {
		opset, err := peer.GetOperatorSet(task.OperatorSetId)
		if err != nil {
			return nil, fmt.Errorf("failed to get operator set %d for peer %s: %w", task.OperatorSetId, peer.OperatorAddress, err)
		}
		operators = append(operators, &aggregation.Operator[common.Address]{
			Address:   peer.OperatorAddress,
			PublicKey: opset.WrappedPublicKey.ECDSAAddress,
		})
	}

	ta, err := aggregation.NewECDSATaskResultAggregator(
		ctx,
		task.TaskId,
		task.BlockNumber,
		task.OperatorSetId,
		task.ThresholdBips,
		task.Payload,
		task.DeadlineUnixSeconds,
		operators,
	)
	if err != nil {
		return nil, err
	}
	ts := &TaskSession[ecdsa.Signature, aggregation.AggregatedECDSACertificate, common.Address]{
		Task:                task,
		aggregatorAddress:   aggregatorAddress,
		aggregatorSignature: aggregatorSignature,
		results:             sync.Map{},
		context:             ctx,
		contextCancel:       cancel,
		logger:              logger,
		taskAggregator:      ta,
		operatorPeersWeight: operatorPeersWeight,
		thresholdMet:        atomic.Bool{},
	}
	ts.resultsCount.Store(0)
	ts.thresholdMet.Store(false)

	return ts, nil
}

func (ts *TaskSession[SigT, CertT, PubKeyT]) Process() (*CertT, error) {
	ts.logger.Sugar().Infow("task session started",
		zap.String("taskId", ts.Task.TaskId),
	)

	certChan := make(chan *CertT, 1)
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

func (ts *TaskSession[SigT, CertT, PubKeyT]) Broadcast() (*CertT, error) {
	ts.logger.Sugar().Infow("task session broadcast started",
		zap.String("taskId", ts.Task.TaskId),
		zap.Any("recipientOperators", ts.operatorPeersWeight.Operators),
	)
	taskSubmission := &executorV1.TaskSubmission{
		TaskId:             ts.Task.TaskId,
		AvsAddress:         ts.Task.AVSAddress,
		AggregatorAddress:  ts.aggregatorAddress,
		Payload:            ts.Task.Payload,
		Signature:          ts.aggregatorSignature,
		OperatorSetId:      ts.Task.OperatorSetId,
		ReferenceTimestamp: ts.operatorPeersWeight.RootReferenceTimestamp,
	}
	ts.logger.Sugar().Infow("broadcasting task session to operators",
		zap.Any("taskSubmission", taskSubmission),
		zap.Any("operatorPeers", ts.operatorPeersWeight.Operators),
	)

	resultsChan := make(chan *types.TaskResult)
	doneChan := make(chan struct{})

	for _, peer := range ts.operatorPeersWeight.Operators {
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
					zap.String("networkAddress", socket),
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
			tr := types.TaskResultFromTaskResultProto(res)
			outputSize := len(tr.Output)
			if outputSize >= maximumTaskResponseSize {
				ts.logger.Sugar().Errorw("dropping response exceeding maximum output size",
					zap.String("taskId", ts.Task.TaskId),
					zap.Int("size", outputSize),
					zap.Int("maximum", maximumTaskResponseSize),
				)
				return
			}

			// Check if done before sending to prevent race condition
			select {
			case resultsChan <- tr:
			case <-doneChan:
				ts.logger.Sugar().Infow("task threshold already met, discarding result",
					zap.String("taskId", ts.Task.TaskId),
					zap.String("operatorAddress", peer.OperatorAddress),
				)
				return
			}
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
		if ts.Task.TaskId != taskResult.TaskId {
			ts.logger.Sugar().Errorw("task ID mismatch: expected",
				zap.String("expected", ts.Task.TaskId),
				zap.String("received", taskResult.TaskId),
			)
			continue
		}
		if err := ts.taskAggregator.ProcessNewSignature(ts.context, ts.Task.TaskId, taskResult); err != nil {
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

		// Signal producers to stop before closing results channel
		close(doneChan)
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
