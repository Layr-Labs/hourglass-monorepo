package taskBlockContextManager

import (
	"context"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contextManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

type TaskBlockContextManager struct {
	mu              sync.RWMutex
	blockContexts   map[uint64]*contextManager.BlockContext
	parentCtx       context.Context
	logger          *zap.Logger
	cleanupInterval time.Duration
}

func NewTaskBlockContextManager(parentCtx context.Context, logger *zap.Logger) *TaskBlockContextManager {
	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*contextManager.BlockContext),
		parentCtx:       parentCtx,
		logger:          logger.With(zap.String("component", "TaskBlockContextManager")),
		cleanupInterval: 5 * time.Minute,
	}

	go mgr.cleanupExpiredContexts()

	return mgr
}

func (bcm *TaskBlockContextManager) GetContext(blockNumber uint64, task *types.Task) context.Context {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()

	if blockCtx, exists := bcm.blockContexts[blockNumber]; exists {
		return blockCtx.Ctx
	}

	ctx, cancel := context.WithDeadline(bcm.parentCtx, *task.DeadlineUnixSeconds)

	bcm.blockContexts[blockNumber] = &contextManager.BlockContext{
		Ctx:    ctx,
		Cancel: cancel,
	}

	bcm.logger.Debug("Created new context for block",
		zap.Uint64("blockNumber", blockNumber),
		zap.String("taskId", task.TaskId),
		zap.Time("deadline", *task.DeadlineUnixSeconds),
	)

	return ctx
}

func (bcm *TaskBlockContextManager) CancelBlock(blockNumber uint64) {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()

	if blockCtx, exists := bcm.blockContexts[blockNumber]; exists {
		blockCtx.Cancel()
		delete(bcm.blockContexts, blockNumber)

		bcm.logger.Info("Cancelled context for reorged block",
			zap.Uint64("blockNumber", blockNumber),
		)
	}
}

func (bcm *TaskBlockContextManager) cleanupExpiredContexts() {
	ticker := time.NewTicker(bcm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bcm.cleanupContexts()
		case <-bcm.parentCtx.Done():
			bcm.logger.Info("Parent context cancelled, stopping cleanup")
			return
		}
	}
}

func (bcm *TaskBlockContextManager) cleanupContexts() {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()

	blocksToRemove := make([]uint64, 0)

	for blockNum, blockCtx := range bcm.blockContexts {
		select {
		case <-blockCtx.Ctx.Done():
			blocksToRemove = append(blocksToRemove, blockNum)
		default:
		}
	}

	for _, blockNum := range blocksToRemove {
		delete(bcm.blockContexts, blockNum)
	}

	if len(blocksToRemove) > 0 {
		bcm.logger.Debug("Cleaned up expired/cancelled block contexts",
			zap.Int("removedCount", len(blocksToRemove)),
			zap.Int("remainingCount", len(bcm.blockContexts)),
		)
	}
}
