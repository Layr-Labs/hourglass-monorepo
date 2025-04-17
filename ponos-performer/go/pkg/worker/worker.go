package worker

import "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer"

type IWorker interface {
	HandleTask(task *performer.Task) (*performer.TaskResult, error)
	ValidateTask(task *performer.Task) error
}
