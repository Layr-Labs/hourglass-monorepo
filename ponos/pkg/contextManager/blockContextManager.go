package contextManager

import (
	"context"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

type IBlockContextManager interface {
	GetContext(blockNumber uint64, task *types.Task) context.Context
	CancelBlock(blockNumber uint64)
}

type BlockContext struct {
	Ctx     context.Context
	Cancel  context.CancelFunc
	TaskIDs map[string]struct{}
}
