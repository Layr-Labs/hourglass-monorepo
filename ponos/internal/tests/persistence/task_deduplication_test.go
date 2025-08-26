package persistence_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/badger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskDeduplicationWithBadgerDB tests that tasks are properly deduplicated when BadgerDB is enabled
func TestTaskDeduplicationWithBadgerDB(t *testing.T) {
	// Create a temporary directory for BadgerDB
	tempDir := t.TempDir()
	badgerDir := filepath.Join(tempDir, "badger")

	// Create BadgerDB store
	badgerStore, err := badger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
		Dir:      badgerDir,
		InMemory: false,
	})
	require.NoError(t, err)
	defer badgerStore.Close()

	ctx := context.Background()

	taskId := "test-task-123"

	// First check - task should not be processed
	processed, err := badgerStore.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.False(t, processed, "task should not be processed initially")

	// Mark task as processed
	err = badgerStore.MarkTaskProcessed(ctx, taskId)
	require.NoError(t, err)

	// Check that task is now marked as processed
	processed, err = badgerStore.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.True(t, processed, "task should be marked as processed")

	// Close and reopen the store to verify persistence
	badgerStore.Close()

	// Reopen the store
	badgerStore2, err := badger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
		Dir:      badgerDir,
		InMemory: false,
	})
	require.NoError(t, err)
	defer badgerStore2.Close()

	// Verify task is still marked as processed after restart
	processed, err = badgerStore2.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.True(t, processed, "task should still be marked as processed after restart")
}

// TestTaskDeduplicationWithMemoryStorage tests that memory storage is pass-through
func TestTaskDeduplicationWithMemoryStorage(t *testing.T) {
	ctx := context.Background()

	// Create memory store
	memStore := memory.NewInMemoryExecutorStore()
	defer memStore.Close()

	// Create a test task ID
	taskId := "test-task-456"

	// Initially not processed
	processed, err := memStore.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.False(t, processed)

	// Mark as processed (no-op for memory storage)
	err = memStore.MarkTaskProcessed(ctx, taskId)
	require.NoError(t, err)

	// Check it's still not tracked (pass-through behavior)
	processed, err = memStore.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.False(t, processed, "memory storage should not track processed tasks")

	// Try multiple marks - should all be no-ops
	for i := 0; i < 5; i++ {
		err = memStore.MarkTaskProcessed(ctx, fmt.Sprintf("task-%d", i))
		require.NoError(t, err)
	}

	// All should return false
	for i := 0; i < 5; i++ {
		processed, err = memStore.IsTaskProcessed(ctx, fmt.Sprintf("task-%d", i))
		require.NoError(t, err)
		assert.False(t, processed, "memory storage should not track any tasks")
	}
}

// TestPersistenceAcrossRestarts verifies that BadgerDB actually persists data to disk
func TestPersistenceAcrossRestarts(t *testing.T) {
	tempDir := t.TempDir()
	badgerDir := filepath.Join(tempDir, "badger_persistence_test")
	ctx := context.Background()

	taskIds := []string{"task-1", "task-2", "task-3"}

	// First session: Mark tasks as processed
	{
		store, err := badger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
			Dir:      badgerDir,
			InMemory: false,
		})
		require.NoError(t, err)

		for _, taskId := range taskIds {
			err = store.MarkTaskProcessed(ctx, taskId)
			require.NoError(t, err)
		}

		store.Close()
	}

	// Verify files were created on disk
	entries, err := os.ReadDir(badgerDir)
	require.NoError(t, err)
	assert.Greater(t, len(entries), 0, "BadgerDB should have created files on disk")

	// Second session: Verify tasks are still marked as processed
	{
		store, err := badger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
			Dir:      badgerDir,
			InMemory: false,
		})
		require.NoError(t, err)
		defer store.Close()

		for _, taskId := range taskIds {
			processed, err := store.IsTaskProcessed(ctx, taskId)
			require.NoError(t, err)
			assert.True(t, processed, "task %s should be persisted", taskId)
		}

		// Check that a new task is not processed
		processed, err := store.IsTaskProcessed(ctx, "new-task")
		require.NoError(t, err)
		assert.False(t, processed, "new task should not be processed")
	}
}

// TestConcurrentTaskProcessing tests thread safety of task deduplication
func TestConcurrentTaskProcessing(t *testing.T) {
	tempDir := t.TempDir()
	badgerDir := filepath.Join(tempDir, "badger_concurrent")

	store, err := badger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
		Dir:      badgerDir,
		InMemory: false,
	})
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	taskId := "concurrent-task"

	// Run multiple goroutines trying to mark the same task as processed
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = store.MarkTaskProcessed(ctx, taskId)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Task should be marked as processed
	processed, err := store.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.True(t, processed)
}
