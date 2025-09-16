package taskBlockContextManager

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contextManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Test basic context creation and retrieval
func TestTaskBlockContextManager_BasicContextOperations(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

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
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

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
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

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
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

	// Cancel block that doesn't exist - should not panic
	assert.NotPanics(t, func() {
		mgr.CancelBlock(999)
	})
}

// Test cleanup of expired contexts
func TestTaskBlockContextManager_AutoCleanupExpiredContexts(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()

	// Create manager with very short cleanup interval for testing
	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*contextManager.BlockContext),
		parentCtx:       parentCtx,
		logger:          logger,
		cleanupInterval: 100 * time.Millisecond, // Fast cleanup for testing
		done:            make(chan struct{}),
	}
	go mgr.cleanupExpiredContexts()
	defer mgr.Shutdown()

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

	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*contextManager.BlockContext),
		parentCtx:       parentCtx,
		logger:          logger,
		cleanupInterval: 100 * time.Millisecond,
		done:            make(chan struct{}),
	}
	go mgr.cleanupExpiredContexts()
	defer mgr.Shutdown()

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

// Test shutdown cancels all contexts
func TestTaskBlockContextManager_ShutdownCancelsAll(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	mgr := NewTaskBlockContextManager(parentCtx, logger)

	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	// Create multiple contexts
	contexts := make([]context.Context, 5)
	for i := 0; i < 5; i++ {
		contexts[i] = mgr.GetContext(uint64(100+i), task)
	}

	// Shutdown manager
	mgr.Shutdown()

	// All contexts should be cancelled
	for i, ctx := range contexts {
		select {
		case <-ctx.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Context %d should be cancelled after shutdown", i)
		}
	}

	// Map should be empty
	mgr.mu.RLock()
	assert.Equal(t, 0, len(mgr.blockContexts), "All contexts should be removed after shutdown")
	mgr.mu.RUnlock()
}

// Test that parent context cancellation propagates
func TestTaskBlockContextManager_ParentContextCancellation(t *testing.T) {
	parentCtx, parentCancel := context.WithCancel(context.Background())
	logger := zap.NewNop()
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

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
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

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
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

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
	mgr.doCleanup()

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
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

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
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

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

	mgr := &TaskBlockContextManager{
		blockContexts:   make(map[uint64]*contextManager.BlockContext),
		parentCtx:       parentCtx,
		logger:          logger,
		cleanupInterval: 50 * time.Millisecond, // Very short interval for testing
		done:            make(chan struct{}),
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

	mgr.Shutdown()
}

// Test empty manager state
func TestTaskBlockContextManager_EmptyState(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	mgr := NewTaskBlockContextManager(parentCtx, logger)
	defer mgr.Shutdown()

	// Cleanup on empty manager should not panic
	assert.NotPanics(t, func() {
		mgr.doCleanup()
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
