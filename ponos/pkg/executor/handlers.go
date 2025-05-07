package executor

import (
	"context"
	"fmt"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/common/v1"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/aggregatorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func (e *Executor) SubmitTask(ctx context.Context, req *executorV1.TaskSubmission) (*commonV1.SubmitAck, error) {
	err := e.handleReceivedTask(req)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to handle received task",
			"taskId", req.TaskId,
			"avsAddress", req.AvsAddress,
			"error", err,
		)
		return &commonV1.SubmitAck{Message: err.Error(), Success: false}, nil
	}
	return &commonV1.SubmitAck{Message: "Scheduled task", Success: true}, nil
}

func (e *Executor) handleReceivedTask(task *executorV1.TaskSubmission) error {
	e.logger.Sugar().Infow("Received task from AVS avsPerformer",
		"taskId", task.TaskId,
		"avsAddress", task.AvsAddress,
	)
	avsAddress := strings.ToLower(task.GetAvsAddress())
	if avsAddress == "" {
		return fmt.Errorf("AVS address is empty")
	}

	avsPerformer, ok := e.avsPerformers[task.AvsAddress]
	if !ok {
		return fmt.Errorf("AVS avsPerformer not found for address %s", task.AvsAddress)
	}

	pt := performerTask.NewPerformerTaskFromTaskSubmissionProto(task)

	if err := avsPerformer.ValidateTaskSignature(pt); err != nil {
		return fmt.Errorf("failed to validate task signature: %w", err)
	}

	e.inflightTasks.Store(task.TaskId, task)

	err := avsPerformer.RunTask(context.Background(), pt)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to run task",
			"taskId", task.TaskId,
			"avsAddress", task.AvsAddress,
			"error", err,
		)
		return status.Errorf(codes.Internal, "Failed to run task %s", err.Error())
	}
	return nil
}

func (e *Executor) receiveTaskResponse(originalTask *performerTask.PerformerTask, response *performerTask.PerformerTaskResult, err error) {
	if err != nil {
		e.logger.Sugar().Errorw("Encountered error while receiving task response",
			zap.String("taskId", originalTask.TaskID),
			zap.String("avsAddress", originalTask.Avs),
			zap.Error(err),
		)
		return
	}
	e.logger.Sugar().Infow("Received task response",
		zap.Any("response", response),
	)

	storedTask, ok := e.inflightTasks.Load(response.TaskID)
	if !ok {
		e.logger.Sugar().Errorw("PerformerTask not found in inflight tasks",
			zap.String("taskId", response.TaskID),
			zap.Error(err),
		)
		return
	}
	task := storedTask.(*executorV1.TaskSubmission)

	// TODO(seanmcgary): should probably assume secure unless localhost or something...
	aggClient, err := aggregatorClient.NewAggregatorClient(task.AggregatorUrl, true)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to create aggregator client",
			zap.String("taskId", task.TaskId),
			zap.String("avsAddress", task.AvsAddress),
			zap.Error(err),
		)
		return
	}

	sig, err := e.signResult(response)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to sign result",
			zap.String("taskId", task.TaskId),
			zap.String("avsAddress", task.AvsAddress),
			zap.Error(err),
		)
		return
	}

	e.logger.Sugar().Infow("Submitting task result to aggregator",
		zap.String("taskId", task.TaskId),
		zap.String("avsAddress", task.AvsAddress),
		zap.String("aggregatorUrl", task.AggregatorUrl),
		zap.String("operatorAddress", e.config.Operator.Address),
		zap.String("signature", string(sig)),
	)

	// TODO(seanmcgary): add a retry wrapper around this call to handle cases where the aggregator is unreachable
	_, err = aggClient.SubmitTaskResult(context.Background(), &aggregatorV1.TaskResult{
		TaskId:          response.TaskID,
		OperatorAddress: e.config.Operator.Address,
		Output:          response.Result,
		Signature:       sig,
		AvsAddress:      task.AvsAddress,
	})
	if err != nil {
		e.logger.Sugar().Errorw("Failed to submit task result",
			zap.String("taskId", task.TaskId),
			zap.String("avsAddress", task.AvsAddress),
			zap.Error(err),
		)
		return
	}
	e.inflightTasks.Delete(task.TaskId)
}

func (e *Executor) signResult(result *performerTask.PerformerTaskResult) ([]byte, error) {
	// Generate a keccak256 hash of the result so that our signature is fixed in size.
	// This is for compatibility with the certificate verifier.
	digestBytes := util.GetKeccak256Digest(result.Result)

	return e.signer.SignMessage(digestBytes[:])
}
