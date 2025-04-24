package executorClient

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

type IExecutorClient interface {
	SubmitTask(ctx context.Context, task *types.Task) error
}
