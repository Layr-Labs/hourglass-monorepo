package executionManager

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

// TODO: change LoadResults to ProcessResult and make PonosExecutionManager store result in shared result queue.
type IExecutionManager interface {
	ExecuteTask(task *types.Task) error
	LoadResults() []*types.TaskResult
}
