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
	done            chan struct{}
}

func NewTaskBlockContextManager(parentCtx context.Context, logger *zap.Logger) *TaskBlockContextManager {
	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*contextManager.BlockContext),
		parentCtx:       parentCtx,
		logger:          logger.With(zap.String("component", "TaskBlockContextManager")),
		cleanupInterval: 5 * time.Minute,
		done:            make(chan struct{}),
	}

	go mgr.cleanupExpiredContexts()

	return mgr
}

func (bcm *TaskBlockContextManager) GetContext(blockNumber uint64, task *types.Task) context.Context {
	bcm.mu.RLock()
	if blockCtx, exists := bcm.blockContexts[blockNumber]; exists {
		bcm.mu.RUnlock()
		return blockCtx.Ctx
	}
	bcm.mu.RUnlock()

	// Need to create a new context
	bcm.mu.Lock()
	defer bcm.mu.Unlock()

	// Double-check in case another goroutine created it
	if blockCtx, exists := bcm.blockContexts[blockNumber]; exists {
		return blockCtx.Ctx
	}

	// Create context with deadline from task
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithDeadline(bcm.parentCtx, *task.DeadlineUnixSeconds)

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

// CancelBlock cancels the context for a given block number
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

// cleanupExpiredContexts runs periodically to remove expired or cancelled contexts from memory
func (bcm *TaskBlockContextManager) cleanupExpiredContexts() {
	ticker := time.NewTicker(bcm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bcm.doCleanup()
		case <-bcm.done:
			// Stop cleanup when manager is shut down
			bcm.logger.Debug("Stopping context cleanup goroutine")
			return
		case <-bcm.parentCtx.Done():
			// Stop if parent context is cancelled
			bcm.logger.Debug("Parent context cancelled, stopping cleanup")
			return
		}
	}
}

// doCleanup performs the actual cleanup of expired/cancelled contexts
func (bcm *TaskBlockContextManager) doCleanup() {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()

	blocksToRemove := make([]uint64, 0)

	for blockNum, blockCtx := range bcm.blockContexts {
		select {
		case <-blockCtx.Ctx.Done():
			// Context is cancelled or expired
			blocksToRemove = append(blocksToRemove, blockNum)
		default:
			// Context is still active
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

// Shutdown gracefully shuts down the context manager
func (bcm *TaskBlockContextManager) Shutdown() {
	close(bcm.done)

	// Cancel all remaining contexts
	bcm.mu.Lock()
	defer bcm.mu.Unlock()

	for _, blockCtx := range bcm.blockContexts {
		blockCtx.Cancel()
	}

	bcm.blockContexts = make(map[uint64]*contextManager.BlockContext)

	bcm.logger.Info("TaskBlockContextManager shut down")
}
