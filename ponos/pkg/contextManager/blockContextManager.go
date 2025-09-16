package contextManager

import (
	"context"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

// IBlockContextManager manages contexts for blocks with proper cancellation on reorgs
type IBlockContextManager interface {
	GetContext(blockNumber uint64, task *types.Task) context.Context
	CancelBlock(blockNumber uint64)
}

// blockContext holds both the context and its cancel function
type BlockContext struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}
