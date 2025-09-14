package badger

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadgerExecutorStore(t *testing.T) {
	// Create a temporary directory for BadgerDB
	tmpDir, err := os.MkdirTemp("", "badger-executor-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Run the reusable test suite
	suite := &storage.TestSuite{
		NewStore: func() (storage.ExecutorStore, error) {
			cfg := &executorConfig.BadgerConfig{
				Dir: tmpDir,
			}
			return NewBadgerExecutorStore(cfg)
		},
	}
	suite.Run(t)
}

func TestBadgerExecutorStore_Persistence(t *testing.T) {
	// Test that data persists across store restarts
	tmpDir, err := os.MkdirTemp("", "badger-executor-persist-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := &executorConfig.BadgerConfig{
		Dir: tmpDir,
	}

	ctx := context.Background()
	performerId := "test-performer-1"

	// Create store, save data, and close
	{
		store, err := NewBadgerExecutorStore(cfg)
		require.NoError(t, err)

		// Save a performer state
		state := &storage.PerformerState{
			PerformerId:        performerId,
			AvsAddress:         "0xAVS1",
			ContainerId:        "container-123",
			Status:             "running",
			ArtifactRegistry:   "registry.example.com",
			ArtifactDigest:     "sha256:abcdef",
			ArtifactTag:        "latest",
			DeploymentMode:     "docker",
			CreatedAt:          time.Now(),
			LastHealthCheck:    time.Now(),
			ContainerHealthy:   true,
			ApplicationHealthy: true,
		}
		err = store.SavePerformerState(ctx, performerId, state)
		require.NoError(t, err)

		// Close store
		err = store.Close()
		require.NoError(t, err)
	}

	// Reopen store and verify data persists
	{
		store, err := NewBadgerExecutorStore(cfg)
		require.NoError(t, err)
		defer store.Close()

		// Verify performer state exists
		retrievedState, err := store.GetPerformerState(ctx, performerId)
		require.NoError(t, err)
		assert.Equal(t, performerId, retrievedState.PerformerId)
		assert.Equal(t, "container-123", retrievedState.ContainerId)
	}
}

func TestBadgerExecutorStore_InMemory(t *testing.T) {
	// Test in-memory mode
	cfg := &executorConfig.BadgerConfig{
		Dir:      "",
		InMemory: true,
	}

	store, err := NewBadgerExecutorStore(cfg)
	require.NoError(t, err)
	defer store.Close()

	// Run basic operations
	ctx := context.Background()
	state := &storage.PerformerState{
		PerformerId:        "performer-1",
		AvsAddress:         "0xAVS1",
		ContainerId:        "container-1",
		Status:             "running",
		ArtifactRegistry:   "registry.example.com",
		ArtifactDigest:     "sha256:abcdef",
		ArtifactTag:        "latest",
		DeploymentMode:     "docker",
		CreatedAt:          time.Now(),
		LastHealthCheck:    time.Now(),
		ContainerHealthy:   true,
		ApplicationHealthy: true,
	}
	err = store.SavePerformerState(ctx, "performer-1", state)
	require.NoError(t, err)

	retrieved, err := store.GetPerformerState(ctx, "performer-1")
	require.NoError(t, err)
	assert.Equal(t, state.PerformerId, retrieved.PerformerId)
}

func BenchmarkBadgerExecutorStore(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "badger-executor-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	cfg := &executorConfig.BadgerConfig{
		Dir: tmpDir,
	}

	store, err := NewBadgerExecutorStore(cfg)
	require.NoError(b, err)
	defer store.Close()

	ctx := context.Background()

	b.Run("SavePerformerState", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			state := &storage.PerformerState{
				PerformerId:        fmt.Sprintf("performer-%d", i),
				AvsAddress:         "0xAVS1",
				ContainerId:        fmt.Sprintf("container-%d", i),
				Status:             "running",
				ArtifactRegistry:   "registry.example.com",
				ArtifactDigest:     "sha256:abcdef",
				ArtifactTag:        "latest",
				DeploymentMode:     "docker",
				CreatedAt:          time.Now(),
				LastHealthCheck:    time.Now(),
				ContainerHealthy:   true,
				ApplicationHealthy: true,
			}
			_ = store.SavePerformerState(ctx, state.PerformerId, state)
		}
	})

	b.Run("GetPerformerState", func(b *testing.B) {
		// Pre-populate some performers
		for i := 0; i < 100; i++ {
			state := &storage.PerformerState{
				PerformerId:        fmt.Sprintf("performer-get-%d", i),
				AvsAddress:         "0xAVS1",
				ContainerId:        fmt.Sprintf("container-%d", i),
				Status:             "running",
				ArtifactRegistry:   "registry.example.com",
				ArtifactDigest:     "sha256:abcdef",
				ArtifactTag:        "latest",
				DeploymentMode:     "docker",
				CreatedAt:          time.Now(),
				LastHealthCheck:    time.Now(),
				ContainerHealthy:   true,
				ApplicationHealthy: true,
			}
			_ = store.SavePerformerState(ctx, state.PerformerId, state)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetPerformerState(ctx, fmt.Sprintf("performer-get-%d", i%100))
		}
	})

	b.Run("ListPerformerStates", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.ListPerformerStates(ctx)
		}
	})
}

// TestBadgerTaskDeduplication verifies that BadgerDB properly persists and deduplicates tasks
func TestBadgerTaskDeduplication(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "badger-dedup-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewBadgerExecutorStore(&executorConfig.BadgerConfig{
		Dir:      tmpDir,
		InMemory: false,
	})
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	t.Run("UnprocessedTaskReturnsFalse", func(t *testing.T) {
		processed, err := store.IsTaskProcessed(ctx, "unprocessed-task")
		require.NoError(t, err)
		assert.False(t, processed, "unprocessed task should return false")
	})

	t.Run("ProcessedTaskReturnsTrue", func(t *testing.T) {
		taskId := "test-task-123"

		// Mark task as processed
		err := store.MarkTaskProcessed(ctx, taskId)
		require.NoError(t, err)

		// Verify it's marked as processed
		processed, err := store.IsTaskProcessed(ctx, taskId)
		require.NoError(t, err)
		assert.True(t, processed, "processed task should return true")
	})

	t.Run("IdempotentMarking", func(t *testing.T) {
		taskId := "idempotent-task"

		// Mark task multiple times
		for i := 0; i < 3; i++ {
			err := store.MarkTaskProcessed(ctx, taskId)
			require.NoError(t, err)
		}

		// Should still be marked as processed
		processed, err := store.IsTaskProcessed(ctx, taskId)
		require.NoError(t, err)
		assert.True(t, processed, "task should remain processed after multiple marks")
	})

	t.Run("MultipleTasksTracked", func(t *testing.T) {
		taskIds := []string{"task-a", "task-b", "task-c", "task-d"}

		// Mark all tasks as processed
		for _, taskId := range taskIds {
			err := store.MarkTaskProcessed(ctx, taskId)
			require.NoError(t, err)
		}

		// All should be marked as processed
		for _, taskId := range taskIds {
			processed, err := store.IsTaskProcessed(ctx, taskId)
			require.NoError(t, err)
			assert.True(t, processed, "task %s should be marked as processed", taskId)
		}

		// Unprocessed task should still return false
		processed, err := store.IsTaskProcessed(ctx, "never-processed")
		require.NoError(t, err)
		assert.False(t, processed, "unprocessed task should return false")
	})

	t.Run("EmptyTaskIdValidation", func(t *testing.T) {
		err := store.MarkTaskProcessed(ctx, "")
		assert.Error(t, err, "empty task ID should return error")
		assert.Contains(t, err.Error(), "task ID cannot be empty")
	})
}

// TestBadgerTaskPersistence verifies tasks persist across store restarts
func TestBadgerTaskPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "badger-task-persist-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	taskIds := []string{"persistent-1", "persistent-2", "persistent-3"}

	// First session: Mark tasks as processed
	{
		store, err := NewBadgerExecutorStore(&executorConfig.BadgerConfig{
			Dir:      tmpDir,
			InMemory: false,
		})
		require.NoError(t, err)

		for _, taskId := range taskIds {
			err = store.MarkTaskProcessed(ctx, taskId)
			require.NoError(t, err)
		}

		// Verify before closing
		for _, taskId := range taskIds {
			processed, err := store.IsTaskProcessed(ctx, taskId)
			require.NoError(t, err)
			assert.True(t, processed, "task %s should be processed in first session", taskId)
		}

		store.Close()
	}

	// Second session: Verify persistence
	{
		store, err := NewBadgerExecutorStore(&executorConfig.BadgerConfig{
			Dir:      tmpDir,
			InMemory: false,
		})
		require.NoError(t, err)
		defer store.Close()

		// Previously processed tasks should still be processed
		for _, taskId := range taskIds {
			processed, err := store.IsTaskProcessed(ctx, taskId)
			require.NoError(t, err)
			assert.True(t, processed, "task %s should persist after restart", taskId)
		}

		// New task should not be processed
		processed, err := store.IsTaskProcessed(ctx, "new-task")
		require.NoError(t, err)
		assert.False(t, processed, "new task should not be processed")
	}
}

// TestBadgerConcurrentTaskMarking tests thread safety of task deduplication
func TestBadgerConcurrentTaskMarking(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "badger-concurrent-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewBadgerExecutorStore(&executorConfig.BadgerConfig{
		Dir:      tmpDir,
		InMemory: false,
	})
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	taskId := "concurrent-task"

	// Run multiple goroutines trying to mark the same task
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			err := store.MarkTaskProcessed(ctx, taskId)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case err := <-errors:
			t.Fatalf("Concurrent marking failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Task should be marked as processed
	processed, err := store.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.True(t, processed, "task should be marked as processed after concurrent marking")
}

// TestBadgerInMemoryModeDeduplication tests that in-memory mode still tracks tasks (but doesn't persist)
func TestBadgerInMemoryModeDeduplication(t *testing.T) {
	store, err := NewBadgerExecutorStore(&executorConfig.BadgerConfig{
		InMemory: true,
	})
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	taskId := "in-memory-task"

	// Mark task as processed
	err = store.MarkTaskProcessed(ctx, taskId)
	require.NoError(t, err)

	// Should be marked as processed
	processed, err := store.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.True(t, processed, "in-memory BadgerDB should track processed tasks")

	// Note: We can't test persistence for in-memory mode since it doesn't persist
}
