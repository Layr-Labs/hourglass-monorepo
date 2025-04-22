package executor

import (
	"context"
	"fmt"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/common/v1"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/tasks"
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

	performerTask := tasks.NewTaskFromTaskSubmissionProto(task)

	err := avsPerformer.RunTask(context.Background(), performerTask)
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
