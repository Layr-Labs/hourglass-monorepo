package server

import (
	"context"
	performerV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/performer"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (pp *PonosPerformer) ExecuteTask(ctx context.Context, task *performerV1.Task) (*performerV1.TaskResult, error) {

	if err := pp.taskWorker.ValidateTask(task); err != nil {
		pp.logger.Sugar().Errorw("Task is invalid",
			zap.String("taskId", task.TaskId),
			zap.String("avs", task.AvsAddress),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "Task is invalid: %s", err.Error())
	}

	res, err := pp.taskWorker.HandleTask(task)
	if err != nil {
		pp.logger.Sugar().Errorw("Failed to handle task",
			zap.String("taskId", task.TaskId),
			zap.String("avs", task.AvsAddress),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "Failed to handle task: %s", err.Error())
	}

	return &performerV1.TaskResult{
		TaskId:     task.TaskId,
		AvsAddress: task.AvsAddress,
		Result:     res.Result,
	}, nil
}

func (pp *PonosPerformer) Health(ctx context.Context, request *performerV1.HealthRequest) (*performerV1.HealthResponse, error) {
	return &performerV1.HealthResponse{
		Status: "running",
	}, nil
}
