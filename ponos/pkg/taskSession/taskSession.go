package taskSession

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const maximumTaskResponseSize = 1.5 * 1024 * 1024

type TaskSession[SigT, CertT, PubKeyT any] struct {
	Task                *types.Task
	signer              signer.ISigner
	context             context.Context
	logger              *zap.Logger
	operatorPeersWeight *operatorManager.PeerWeight
	taskAggregator      aggregation.ITaskResultAggregator[SigT, CertT, PubKeyT]
	aggregatorAddress   string
	tlsEnabled          bool
}

func NewBN254TaskSession(
	ctx context.Context,
	task *types.Task,
	l1ContractCaller contractCaller.IContractCaller,
	aggregatorAddress string,
	signer signer.ISigner,
	operatorPeersWeight *operatorManager.PeerWeight,
	tlsEnabled bool,
	logger *zap.Logger,
) (*TaskSession[bn254.Signature, aggregation.AggregatedBN254Certificate, signing.PublicKey], error) {
	operators := make([]*aggregation.Operator[signing.PublicKey], 0)

	for _, peer := range operatorPeersWeight.Operators {
		opset, err := peer.GetOperatorSet(task.OperatorSetId)
		if err != nil {
			return nil, fmt.Errorf("failed to get operator set %d for peer %s: %w", task.OperatorSetId, peer.OperatorAddress, err)
		}

		// Retrieve weights from the PeerWeight structure
		weights := operatorPeersWeight.Weights[peer.OperatorAddress]

		operators = append(operators, &aggregation.Operator[signing.PublicKey]{
			Address:       peer.OperatorAddress,
			PublicKey:     opset.WrappedPublicKey.PublicKey,
			OperatorIndex: opset.OperatorIndex,
			Weights:       weights,
		})
	}

	ta, err := aggregation.NewBN254TaskResultAggregator(
		ctx,
		task.TaskId,
		task.ReferenceTimestamp,
		task.OperatorSetId,
		task.ThresholdBips,
		l1ContractCaller,
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
		signer:              signer,
		context:             ctx,
		logger:              logger,
		taskAggregator:      ta,
		operatorPeersWeight: operatorPeersWeight,
		tlsEnabled:          tlsEnabled,
	}

	return ts, nil
}

func NewECDSATaskSession(
	ctx context.Context,
	task *types.Task,
	l1ContractCaller contractCaller.IContractCaller,
	aggregatorAddress string,
	signer signer.ISigner,
	operatorPeersWeight *operatorManager.PeerWeight,
	tlsEnabled bool,
	logger *zap.Logger,
) (*TaskSession[ecdsa.Signature, aggregation.AggregatedECDSACertificate, common.Address], error) {
	operators := make([]*aggregation.Operator[common.Address], 0)
	for _, peer := range operatorPeersWeight.Operators {
		opset, err := peer.GetOperatorSet(task.OperatorSetId)
		if err != nil {
			return nil, fmt.Errorf("failed to get operator set %d for peer %s: %w", task.OperatorSetId, peer.OperatorAddress, err)
		}

		// Retrieve weights from the PeerWeight structure
		weights := operatorPeersWeight.Weights[peer.OperatorAddress]

		operators = append(operators, &aggregation.Operator[common.Address]{
			Address:       peer.OperatorAddress,
			PublicKey:     opset.WrappedPublicKey.ECDSAAddress,
			OperatorIndex: opset.OperatorIndex,
			Weights:       weights,
		})
	}

	ta, err := aggregation.NewECDSATaskResultAggregator(
		ctx,
		task.TaskId,
		task.ReferenceTimestamp,
		task.OperatorSetId,
		task.ThresholdBips,
		l1ContractCaller,
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
		signer:              signer,
		context:             ctx,
		logger:              logger,
		taskAggregator:      ta,
		operatorPeersWeight: operatorPeersWeight,
		tlsEnabled:          tlsEnabled,
	}

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

	resultsChan := make(chan *types.TaskResult, len(ts.operatorPeersWeight.Operators))
	submissionContext, cancelSubmissions := context.WithCancel(ts.context)
	defer cancelSubmissions()

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
			c, err := executorClient.NewExecutorClient(socket, ts.tlsEnabled)
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

			signature, err := ts.generateSignatureForExecutor(peer.OperatorAddress)
			if err != nil {
				ts.logger.Sugar().Errorw("Failed to generate executor signature",
					zap.String("taskId", ts.Task.TaskId),
					zap.String("executorAddress", peer.OperatorAddress),
					zap.Error(err),
				)
				return
			}

			taskSubmission := &executorV1.TaskSubmission{
				TaskId:             ts.Task.TaskId,
				AggregatorAddress:  ts.aggregatorAddress,
				AvsAddress:         ts.Task.AVSAddress,
				Payload:            ts.Task.Payload,
				Signature:          signature,
				OperatorSetId:      ts.Task.OperatorSetId,
				ReferenceTimestamp: ts.Task.ReferenceTimestamp,
				ExecutorAddress:    peer.OperatorAddress,
				TaskBlockNumber:    ts.Task.L1ReferenceBlockNumber,
				Version:            ts.Task.Version,
			}
			ts.logger.Sugar().Infow("broadcasting task session to operators",
				zap.Any("taskSubmission", taskSubmission),
				zap.Any("operatorPeers", ts.operatorPeersWeight.Operators),
			)

			res, err := c.SubmitTask(submissionContext, taskSubmission)
			if err != nil {

				if err.Error() == context.Canceled.Error() {
					ts.logger.Sugar().Infow("task session submission cancelled")
					return
				}

				ts.logger.Sugar().Errorw("Failed to submit task to executor",
					zap.String("executorAddress", peer.OperatorAddress),
					zap.String("networkAddress", socket),
					zap.String("taskId", ts.Task.TaskId),
					zap.Error(err),
				)
				return
			}

			if !strings.EqualFold(res.OperatorAddress, peer.OperatorAddress) {
				ts.logger.Sugar().Errorw("Operator address mismatch in response",
					zap.String("taskId", ts.Task.TaskId),
					zap.String("expected", peer.OperatorAddress),
					zap.String("claimed", res.OperatorAddress),
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

			resultsChan <- tr
		}(peer)
	}

	for {
		select {
		case taskResult := <-resultsChan:
			ts.logger.Sugar().Infow("received task result on channel",
				zap.String("taskId", taskResult.TaskId),
				zap.String("operatorAddress", taskResult.OperatorAddress),
			)
			if ts.Task.TaskId != taskResult.TaskId {
				ts.logger.Sugar().Errorw("task ID mismatch: expected",
					zap.String("expected", ts.Task.TaskId),
					zap.String("received", taskResult.TaskId),
				)
				continue
			}
			if err := ts.taskAggregator.ProcessNewSignature(ts.context, taskResult); err != nil {
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
			cancelSubmissions()

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
		case <-ts.context.Done():
			ts.logger.Sugar().Errorw("task session context cancelled while waiting for results",
				zap.String("taskId", ts.Task.TaskId),
				zap.Error(ts.context.Err()),
			)
			return nil, fmt.Errorf("task session context done while waiting for results: %w", ts.context.Err())
		}
	}
}

func (ts *TaskSession[SigT, CertT, PubKeyT]) generateSignatureForExecutor(executorAddress string) ([]byte, error) {
	encodedMessage, err := util.EncodeTaskSubmissionMessageVersioned(
		ts.Task.TaskId,
		ts.Task.AVSAddress,
		executorAddress,
		ts.Task.ReferenceTimestamp,
		ts.Task.L1ReferenceBlockNumber,
		ts.Task.OperatorSetId,
		ts.Task.Payload,
		ts.Task.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encode task submission message: %w", err)
	}

	return ts.signer.SignMessage(encodedMessage)
}
