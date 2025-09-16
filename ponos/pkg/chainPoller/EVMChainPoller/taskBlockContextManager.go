package EVMChainPoller

import (
	"context"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

// IBlockContextManager manages contexts for blocks with proper cancellation on reorgs
type IBlockContextManager interface {
	GetContext(blockNumber uint64, task *types.Task) context.Context
	CancelBlock(blockNumber uint64)
}

// blockContext holds both the context and its cancel function
type blockContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// TaskBlockContextManager provides per-block context management with automatic cleanup
type TaskBlockContextManager struct {
	mu              sync.RWMutex
	blockContexts   map[uint64]*blockContext
	parentCtx       context.Context
	logger          *zap.Logger
	cleanupInterval time.Duration
	done            chan struct{}
}

// NewTaskBlockContextManager creates a new TaskBlockContextManager instance
func NewTaskBlockContextManager(parentCtx context.Context, logger *zap.Logger) *TaskBlockContextManager {
	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*blockContext),
		parentCtx:       parentCtx,
		logger:          logger.With(zap.String("component", "TaskBlockContextManager")),
		cleanupInterval: 5 * time.Minute,
		done:            make(chan struct{}),
	}

	go mgr.cleanupExpiredContexts()

	return mgr
}

// GetContext returns an existing context for the block or creates a new one with the task deadline
func (bcm *TaskBlockContextManager) GetContext(blockNumber uint64, task *types.Task) context.Context {
	bcm.mu.RLock()
	if blockCtx, exists := bcm.blockContexts[blockNumber]; exists {
		bcm.mu.RUnlock()
		return blockCtx.ctx
	}
	bcm.mu.RUnlock()

	// Need to create a new context
	bcm.mu.Lock()
	defer bcm.mu.Unlock()

	// Double-check in case another goroutine created it
	if blockCtx, exists := bcm.blockContexts[blockNumber]; exists {
		return blockCtx.ctx
	}

	// Create context with deadline from task
	var ctx context.Context
	var cancel context.CancelFunc

	if task.DeadlineUnixSeconds != nil {
		ctx, cancel = context.WithDeadline(bcm.parentCtx, *task.DeadlineUnixSeconds)
	} else {
		// No deadline specified, use parent context with cancel
		ctx, cancel = context.WithCancel(bcm.parentCtx)
	}

	bcm.blockContexts[blockNumber] = &blockContext{
		ctx:    ctx,
		cancel: cancel,
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
		blockCtx.cancel()
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
		case <-blockCtx.ctx.Done():
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
		blockCtx.cancel()
	}

	bcm.blockContexts = make(map[uint64]*blockContext)

	bcm.logger.Info("TaskBlockContextManager shut down")
}