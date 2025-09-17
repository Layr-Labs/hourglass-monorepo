package taskBlockContextManager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contextManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Test basic context creation and retrieval
func TestTaskBlockContextManager_BasicContextOperations(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	// Create a task with deadline
	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		AVSAddress:          "0xavs",
		DeadlineUnixSeconds: &deadline,
	}

	// Get context for block 100
	ctx := mgr.GetContext(100, task)
	assert.NotNil(t, ctx)

	// Verify context has the correct deadline
	ctxDeadline, hasDeadline := ctx.Deadline()
	assert.True(t, hasDeadline)
	assert.WithinDuration(t, deadline, ctxDeadline, 10*time.Millisecond)
}

// Test that getting context for same block returns same context
func TestTaskBlockContextManager_ContextCaching(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	deadline := time.Now().Add(1 * time.Hour)
	task1 := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}
	task2 := &types.Task{
		TaskId:              "task-2",
		DeadlineUnixSeconds: &deadline,
	}

	// Get context for block 100 with different tasks
	ctx1 := mgr.GetContext(100, task1)
	ctx2 := mgr.GetContext(100, task2)

	// Should return the same context
	assert.Equal(t, ctx1, ctx2, "Same block should return same context")
}

// Test cancelling specific blocks
func TestTaskBlockContextManager_CancelSpecificBlock(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	// Create contexts for multiple blocks
	ctx100 := mgr.GetContext(100, task)
	ctx101 := mgr.GetContext(101, task)
	ctx102 := mgr.GetContext(102, task)

	// Cancel only block 101
	mgr.CancelBlock(101)

	// Verify only ctx101 is cancelled
	select {
	case <-ctx100.Done():
		t.Fatal("Context for block 100 should not be cancelled")
	default:
		// Expected
	}

	select {
	case <-ctx101.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context for block 101 should be cancelled")
	}

	select {
	case <-ctx102.Done():
		t.Fatal("Context for block 102 should not be cancelled")
	default:
		// Expected
	}
}

// Test cancelling non-existent block
func TestTaskBlockContextManager_CancelNonExistentBlock(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	// Cancel block that doesn't exist - should not panic
	assert.NotPanics(t, func() {
		mgr.CancelBlock(999)
	})
}

// Test that CancelBlock deletes tasks from store
func TestTaskBlockContextManager_CancelBlock_DeletesTasksFromStore(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(ctx, store, logger)

	deadline := time.Now().Add(1 * time.Hour)

	// Create multiple tasks for the same block
	task1 := &types.Task{
		TaskId:              "task-1",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &deadline,
		SourceBlockNumber:   100,
	}
	task2 := &types.Task{
		TaskId:              "task-2",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &deadline,
		SourceBlockNumber:   100,
	}
	task3 := &types.Task{
		TaskId:              "task-3",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &deadline,
		SourceBlockNumber:   100,
	}

	// Create task for different block (should not be deleted)
	task4 := &types.Task{
		TaskId:              "task-4",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &deadline,
		SourceBlockNumber:   101,
	}

	// Save all tasks to store
	require.NoError(t, store.SavePendingTask(ctx, task1))
	require.NoError(t, store.SavePendingTask(ctx, task2))
	require.NoError(t, store.SavePendingTask(ctx, task3))
	require.NoError(t, store.SavePendingTask(ctx, task4))

	// Get contexts for the tasks (this registers them with the block)
	_ = mgr.GetContext(100, task1)
	_ = mgr.GetContext(100, task2)
	_ = mgr.GetContext(100, task3)
	_ = mgr.GetContext(101, task4)

	// Verify all tasks exist in store before cancellation
	storedTask1, err := store.GetTask(ctx, "task-1")
	assert.NoError(t, err)
	assert.NotNil(t, storedTask1)

	storedTask2, err := store.GetTask(ctx, "task-2")
	assert.NoError(t, err)
	assert.NotNil(t, storedTask2)

	storedTask3, err := store.GetTask(ctx, "task-3")
	assert.NoError(t, err)
	assert.NotNil(t, storedTask3)

	storedTask4, err := store.GetTask(ctx, "task-4")
	assert.NoError(t, err)
	assert.NotNil(t, storedTask4)

	// Cancel block 100 (should delete task-1, task-2, and task-3)
	mgr.CancelBlock(100)

	// Verify tasks from block 100 are deleted
	_, err = store.GetTask(ctx, "task-1")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrNotFound), "task-1 should be deleted")

	_, err = store.GetTask(ctx, "task-2")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrNotFound), "task-2 should be deleted")

	_, err = store.GetTask(ctx, "task-3")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrNotFound), "task-3 should be deleted")

	// Verify task from block 101 still exists
	storedTask4After, err := store.GetTask(ctx, "task-4")
	assert.NoError(t, err)
	assert.NotNil(t, storedTask4After)
	assert.Equal(t, "task-4", storedTask4After.TaskId)

	// Verify the block context is removed
	mgr.mu.RLock()
	_, exists := mgr.blockContexts[100]
	mgr.mu.RUnlock()
	assert.False(t, exists, "Block 100 context should be removed after cancellation")

	// Verify block 101 context still exists
	mgr.mu.RLock()
	_, exists = mgr.blockContexts[101]
	mgr.mu.RUnlock()
	assert.True(t, exists, "Block 101 context should still exist")
}

// Test cleanup of expired contexts
func TestTaskBlockContextManager_AutoCleanupExpiredContexts(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()

	// Create manager with very short cleanup interval for testing
	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*contextManager.BlockContext),
		parentCtx:       parentCtx,
		store:           store,
		logger:          logger,
		cleanupInterval: 100 * time.Millisecond, // Fast cleanup for testing
	}
	go mgr.cleanupExpiredContexts()

	// Create context with very short deadline
	shortDeadline := time.Now().Add(50 * time.Millisecond)
	task1 := &types.Task{
		TaskId:              "task-short",
		DeadlineUnixSeconds: &shortDeadline,
	}

	// Create context with long deadline
	longDeadline := time.Now().Add(10 * time.Hour)
	task2 := &types.Task{
		TaskId:              "task-long",
		DeadlineUnixSeconds: &longDeadline,
	}

	ctx1 := mgr.GetContext(100, task1)
	_ = mgr.GetContext(101, task2)

	// Verify both exist initially
	mgr.mu.RLock()
	assert.Equal(t, 2, len(mgr.blockContexts))
	mgr.mu.RUnlock()

	// Wait for short context to expire and cleanup to run
	<-ctx1.Done()                      // Wait for context to expire
	time.Sleep(200 * time.Millisecond) // Wait for cleanup cycle

	// Verify expired context was cleaned up
	mgr.mu.RLock()
	assert.Equal(t, 1, len(mgr.blockContexts), "Expired context should be cleaned up")
	_, exists100 := mgr.blockContexts[100]
	assert.False(t, exists100, "Block 100 context should be removed")
	_, exists101 := mgr.blockContexts[101]
	assert.True(t, exists101, "Block 101 context should remain")
	mgr.mu.RUnlock()
}

// Test cleanup of cancelled contexts
func TestTaskBlockContextManager_AutoCleanupCancelledContexts(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()

	// Create manager with short cleanup interval
	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*contextManager.BlockContext),
		parentCtx:       parentCtx,
		store:           store,
		logger:          logger,
		cleanupInterval: 100 * time.Millisecond,
	}
	go mgr.cleanupExpiredContexts()

	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	// Create contexts
	_ = mgr.GetContext(100, task)
	_ = mgr.GetContext(101, task)
	_ = mgr.GetContext(102, task)

	// Cancel one block
	mgr.CancelBlock(101)

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify cancelled context stays removed and others remain
	mgr.mu.RLock()
	assert.Equal(t, 2, len(mgr.blockContexts), "Should have 2 contexts remaining")
	_, exists101 := mgr.blockContexts[101]
	assert.False(t, exists101, "Cancelled block 101 should not exist")
	mgr.mu.RUnlock()
}

// Test that parent context cancellation propagates
func TestTaskBlockContextManager_ParentContextCancellation(t *testing.T) {
	parentCtx, parentCancel := context.WithCancel(context.Background())
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	// Create child contexts
	ctx1 := mgr.GetContext(100, task)
	ctx2 := mgr.GetContext(101, task)

	// Cancel parent
	parentCancel()

	// All child contexts should be cancelled
	for i, ctx := range []context.Context{ctx1, ctx2} {
		select {
		case <-ctx.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Child context %d should be cancelled when parent is cancelled", i)
		}
	}
}

// Test multiple cancellations of same block
func TestTaskBlockContextManager_MultipleCancellations(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	ctx := mgr.GetContext(100, task)

	// Cancel the block multiple times - should not panic
	mgr.CancelBlock(100)
	mgr.CancelBlock(100)
	mgr.CancelBlock(100)

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be cancelled")
	}

	// Getting context for same block should create new one
	newCtx := mgr.GetContext(100, task)
	assert.NotEqual(t, ctx, newCtx, "Should get new context after cancellation")

	// New context should not be cancelled
	select {
	case <-newCtx.Done():
		t.Fatal("New context should not be cancelled")
	default:
		// Expected
	}
}

// Test cleanup doesn't interfere with active contexts
func TestTaskBlockContextManager_CleanupDoesNotAffectActive(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	// Create contexts with various deadlines
	now := time.Now()
	tasks := []*types.Task{
		{
			TaskId:              "task-expired",
			DeadlineUnixSeconds: func() *time.Time { t := now.Add(-1 * time.Hour); return &t }(),
		},
		{
			TaskId:              "task-active",
			DeadlineUnixSeconds: func() *time.Time { t := now.Add(1 * time.Hour); return &t }(),
		},
	}

	// Create contexts
	expiredCtx := mgr.GetContext(100, tasks[0])
	activeCtx := mgr.GetContext(101, tasks[1])

	// Expired context should be done
	select {
	case <-expiredCtx.Done():
		// Expected
	default:
		t.Fatal("Expired context should be done")
	}

	// Run cleanup manually
	mgr.cleanupContexts()

	// Active context should still be there
	mgr.mu.RLock()
	_, exists := mgr.blockContexts[101]
	mgr.mu.RUnlock()
	assert.True(t, exists, "Active context should remain after cleanup")

	// Active context should not be cancelled
	select {
	case <-activeCtx.Done():
		t.Fatal("Active context should not be cancelled by cleanup")
	default:
		// Expected
	}
}

// Benchmark context creation
func BenchmarkTaskBlockContextManager_GetContext(b *testing.B) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mgr.GetContext(uint64(i), task)
	}
}

// Benchmark concurrent context operations
func BenchmarkTaskBlockContextManager_ConcurrentOps(b *testing.B) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			blockNum := uint64(i % 1000)
			if i%3 == 0 {
				mgr.CancelBlock(blockNum)
			} else {
				_ = mgr.GetContext(blockNum, task)
			}
			i++
		}
	})
}

// Test that cleanup interval is respected
func TestTaskBlockContextManager_CleanupInterval(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()

	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*contextManager.BlockContext),
		parentCtx:       parentCtx,
		store:           store,
		logger:          logger,
		cleanupInterval: 50 * time.Millisecond, // Very short interval for testing
	}

	// Create a context that will expire soon
	deadline := time.Now().Add(25 * time.Millisecond)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	// Create the context
	_ = mgr.GetContext(100, task)

	// Start the cleanup goroutine
	go mgr.cleanupExpiredContexts()

	// Wait for the context to expire
	time.Sleep(30 * time.Millisecond)

	// Verify context exists before cleanup
	mgr.mu.RLock()
	_, exists := mgr.blockContexts[100]
	mgr.mu.RUnlock()
	assert.True(t, exists, "Context should exist before cleanup")

	// Wait for cleanup cycle to run
	time.Sleep(60 * time.Millisecond)

	// Verify context was cleaned up
	mgr.mu.RLock()
	_, exists = mgr.blockContexts[100]
	mgr.mu.RUnlock()
	assert.False(t, exists, "Context should be cleaned up after interval")
}

// Test empty manager state
func TestTaskBlockContextManager_EmptyState(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := NewTaskBlockContextManager(parentCtx, store, logger)

	// Cleanup on empty manager should not panic
	assert.NotPanics(t, func() {
		mgr.cleanupContexts()
	})

	// Cancel on empty manager should not panic
	assert.NotPanics(t, func() {
		mgr.CancelBlock(123)
	})

	// Verify manager is still empty
	mgr.mu.RLock()
	assert.Equal(t, 0, len(mgr.blockContexts))
	mgr.mu.RUnlock()
}
