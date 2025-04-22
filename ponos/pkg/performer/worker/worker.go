package worker

import (
	performerV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/performer"
)

type IWorker interface {
	HandleTask(task *performerV1.Task) (*performerV1.TaskResult, error)
	ValidateTask(task *performerV1.Task) error
}
