package worker

import "github.com/Layr-Labs/hourglass-monorepo/ponos-performer/go/pkg/task"

type IWorker interface {
	HandleTask(task *task.Task) (*task.TaskResult, error)
}
