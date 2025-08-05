package server

import (
	"context"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	healthV1 "github.com/Layr-Labs/protocol-apis/gen/protos/grpc/health/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (pp *PonosPerformer) ExecuteTask(ctx context.Context, task *performerV1.TaskRequest) (*performerV1.TaskResponse, error) {

	if err := pp.taskWorker.ValidateTask(task); err != nil {
		pp.logger.Sugar().Errorw("task is invalid",
			zap.String("taskId", string(task.TaskId)),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "task is invalid: %s", err.Error())
	}

	res, err := pp.taskWorker.HandleTask(task)
	if err != nil {
		pp.logger.Sugar().Errorw("Failed to handle task",
			zap.String("taskId", string(task.TaskId)),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "Failed to handle task: %s", err.Error())
	}

	return &performerV1.TaskResponse{
		TaskId: task.TaskId,
		Result: res.Result,
	}, nil
}

func (pp *PonosPerformer) Check(ctx context.Context, request *healthV1.HealthCheckRequest) (*healthV1.HealthCheckResponse, error) {
	return &healthV1.HealthCheckResponse{
		Status: healthV1.HealthCheckResponse_SERVING,
	}, nil
}

func (pp *PonosPerformer) Watch(request *healthV1.HealthCheckRequest, g grpc.ServerStreamingServer[healthV1.HealthCheckResponse]) error {
	return status.Errorf(codes.Unimplemented, "Watch method is not implemented")
}

func (pp *PonosPerformer) StartSync(ctx context.Context, request *performerV1.StartSyncRequest) (*performerV1.StartSyncResponse, error) {
	return &performerV1.StartSyncResponse{}, nil
}
