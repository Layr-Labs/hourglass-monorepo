package memory_test

import (
	"context"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInMemoryExecutorStore runs the standard storage test suite
func TestInMemoryExecutorStore(t *testing.T) {
	suite := &storage.TestSuite{
		NewStore: func() (storage.ExecutorStore, error) {
			return memory.NewInMemoryExecutorStore(), nil
		},
	}
	suite.Run(t)
}

// TestInMemorySpecific tests in-memory specific behavior
func TestInMemorySpecific(t *testing.T) {
	t.Run("MultipleInstances", func(t *testing.T) {
		// Test that multiple instances don't share state
		store1 := memory.NewInMemoryExecutorStore()
		store2 := memory.NewInMemoryExecutorStore()

		// Both should have independent state
		if store1 == store2 {
			t.Fatal("NewInMemoryExecutorStore should create independent instances")
		}
	})
}

// TestMemoryPassThroughBehavior verifies that memory storage doesn't track processed tasks
// to avoid unbounded memory growth
func TestMemoryPassThroughBehavior(t *testing.T) {
	store := memory.NewInMemoryExecutorStore()
	defer store.Close()

	ctx := context.Background()

	t.Run("AlwaysReturnsFalse", func(t *testing.T) {
		// Check various task IDs - all should return false
		taskIds := []string{"task-1", "task-2", "special-task", "0xdeadbeef"}

		for _, taskId := range taskIds {
			processed, err := store.IsTaskProcessed(ctx, taskId)
			require.NoError(t, err)
			assert.False(t, processed, "memory storage should always return false for task %s", taskId)
		}
	})

	t.Run("MarkingIsNoOp", func(t *testing.T) {
		taskId := "test-task"

		// Check before marking
		processed, err := store.IsTaskProcessed(ctx, taskId)
		require.NoError(t, err)
		assert.False(t, processed, "should return false before marking")

		// Mark as processed
		err = store.MarkTaskProcessed(ctx, taskId)
		require.NoError(t, err)

		// Check after marking - should still be false (pass-through)
		processed, err = store.IsTaskProcessed(ctx, taskId)
		require.NoError(t, err)
		assert.False(t, processed, "should still return false after marking (pass-through)")
	})

	t.Run("MultipleMarksRemainNoOp", func(t *testing.T) {
		taskId := "multi-mark-task"

		// Mark multiple times
		for i := 0; i < 5; i++ {
			err := store.MarkTaskProcessed(ctx, taskId)
			require.NoError(t, err)

			// Should always return false
			processed, err := store.IsTaskProcessed(ctx, taskId)
			require.NoError(t, err)
			assert.False(t, processed, "should return false after %d marks", i+1)
		}
	})

	t.Run("EmptyTaskIdValidation", func(t *testing.T) {
		// Empty task ID should still validate
		err := store.MarkTaskProcessed(ctx, "")
		assert.Error(t, err, "empty task ID should return error")
		assert.Contains(t, err.Error(), "task ID cannot be empty")
	})
}

// TestMemoryStoreAfterClose verifies proper error handling after store is closed
func TestMemoryStoreAfterClose(t *testing.T) {
	store := memory.NewInMemoryExecutorStore()
	ctx := context.Background()

	// Close the store
	err := store.Close()
	require.NoError(t, err)

	// Operations should return ErrStoreClosed
	err = store.MarkTaskProcessed(ctx, "task-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "store is closed")

	_, err = store.IsTaskProcessed(ctx, "task-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "store is closed")
}
