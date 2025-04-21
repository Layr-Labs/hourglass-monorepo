package executorClient

import "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"

type IExecutorClient interface {
	SubmitTask(task *types.Task) error
}
