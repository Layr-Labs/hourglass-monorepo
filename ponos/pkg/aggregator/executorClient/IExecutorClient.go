package executorClient

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/tasks"
)

type IExecutorClient interface {
	SubmitTask(ctx context.Context, task *tasks.Task) error
}
