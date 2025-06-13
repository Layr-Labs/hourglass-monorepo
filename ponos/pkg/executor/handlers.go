package executor

import (
	"context"
	"fmt"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func (e *Executor) SubmitTask(ctx context.Context, req *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	res, err := e.handleReceivedTask(req)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to handle received task",
			"taskId", req.TaskId,
			"avsAddress", req.AvsAddress,
			"error", err,
		)
		return nil, fmt.Errorf("Failed to handle received task: %w", err)
	}
	return res, nil
}

func (e *Executor) handleReceivedTask(task *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	e.logger.Sugar().Infow("Received task from AVS avsPerformer",
		"taskId", task.TaskId,
		"avsAddress", task.AvsAddress,
	)
	avsAddress := strings.ToLower(task.GetAvsAddress())
	if avsAddress == "" {
		return nil, fmt.Errorf("AVS address is empty")
	}

	avsPerformer, ok := e.avsPerformers[task.AvsAddress]
	if !ok {
		return nil, fmt.Errorf("AVS avsPerformer not found for address %s", task.AvsAddress)
	}

	pt := performerTask.NewPerformerTaskFromTaskSubmissionProto(task)

	if err := avsPerformer.ValidateTaskSignature(pt); err != nil {
		return nil, fmt.Errorf("failed to validate task signature: %w", err)
	}
	e.inflightTasks.Store(task.TaskId, task)

	response, err := avsPerformer.RunTask(context.Background(), pt)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to run task",
			"taskId", task.TaskId,
			"avsAddress", task.AvsAddress,
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "Failed to run task %s", err.Error())
	}

	sig, err := e.signResult(response)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to sign result",
			zap.String("taskId", task.TaskId),
			zap.String("avsAddress", task.AvsAddress),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "Failed to sign result %s", err.Error())
	}

	e.logger.Sugar().Infow("returning task result to aggregator",
		zap.String("taskId", task.TaskId),
		zap.String("avsAddress", task.AvsAddress),
		zap.String("operatorAddress", e.config.Operator.Address),
		zap.String("signature", string(sig)),
	)

	e.inflightTasks.Delete(task.TaskId)
	return &executorV1.TaskResult{
		TaskId:          response.TaskID,
		OperatorAddress: e.config.Operator.Address,
		Output:          response.Result,
		Signature:       sig,
		AvsAddress:      task.AvsAddress,
	}, nil
}

func (e *Executor) signResult(result *performerTask.PerformerTaskResult) ([]byte, error) {
	// Generate a keccak256 hash of the result so that our signature is fixed in size.
	// This is for compatibility with the certificate verifier.
	digestBytes := util.GetKeccak256Digest(result.Result)

	return e.signer.SignMessageForSolidity(digestBytes)
}
