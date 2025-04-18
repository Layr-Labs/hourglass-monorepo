package executionManager

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

type IExecutionManager interface {
	ExecuteTask(task *types.Task) error
	LoadResults() []*types.TaskResult
}
