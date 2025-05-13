package worker

import (
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
)

type IWorker interface {
	HandleTask(task *performerV1.TaskRequest) (*performerV1.TaskResponse, error)
	ValidateTask(task *performerV1.TaskRequest) error
}
