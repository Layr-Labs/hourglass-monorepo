package persistence_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutorCrashRecovery(t *testing.T) {
	ctx := context.Background()

	// Create in-memory store
	store := memory.NewInMemoryExecutorStore()

	// Simulate some state before "crash"
	// Save performer states
	performer1 := &storage.PerformerState{
		PerformerId:        "performer-1",
		AvsAddress:         "0xavs1",
		ContainerId:        "container-1",
		Status:             "running",
		ArtifactRegistry:   "registry.io/avs1",
		ArtifactTag:        "v1.0.0",
		ArtifactDigest:     "sha256:abc123",
		DeploymentMode:     "docker",
		CreatedAt:          time.Now().Add(-1 * time.Hour),
		LastHealthCheck:    time.Now().Add(-5 * time.Minute),
		ContainerHealthy:   true,
		ApplicationHealthy: true,
	}
	require.NoError(t, store.SavePerformerState(ctx, performer1.PerformerId, performer1))

	performer2 := &storage.PerformerState{
		PerformerId:        "performer-2",
		AvsAddress:         "0xavs2",
		ContainerId:        "container-2",
		Status:             "failed",
		ArtifactRegistry:   "registry.io/avs2",
		ArtifactTag:        "v2.0.0",
		ArtifactDigest:     "sha256:def456",
		DeploymentMode:     "kubernetes",
		CreatedAt:          time.Now().Add(-2 * time.Hour),
		LastHealthCheck:    time.Now().Add(-10 * time.Minute),
		ContainerHealthy:   false,
		ApplicationHealthy: false,
	}
	require.NoError(t, store.SavePerformerState(ctx, performer2.PerformerId, performer2))

	// Simulate "crash" by creating new executor with same store
	// Note: In real usage, executor would be created with all dependencies
	// Here we're just verifying that storage preserved the state correctly

	// Verify state was preserved
	recoveredPerformers, err := store.ListPerformerStates(ctx)
	require.NoError(t, err)
	assert.Len(t, recoveredPerformers, 2)

	recoveredPerformer1, err := store.GetPerformerState(ctx, "performer-1")
	require.NoError(t, err)
	assert.Equal(t, "performer-1", recoveredPerformer1.PerformerId)
	assert.Equal(t, "running", recoveredPerformer1.Status)
	assert.True(t, recoveredPerformer1.ContainerHealthy)
}

func TestPerformerStateManagement(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryExecutorStore()

	// Test multiple performers for same AVS
	avsAddress := "0xavs-multi"
	for i := 0; i < 5; i++ {
		performer := &storage.PerformerState{
			PerformerId:        fmt.Sprintf("performer-%d", i),
			AvsAddress:         avsAddress,
			ContainerId:        fmt.Sprintf("container-%d", i),
			Status:             "running",
			ArtifactRegistry:   "registry.io/avs",
			ArtifactTag:        fmt.Sprintf("v%d.0.0", i),
			ArtifactDigest:     fmt.Sprintf("sha256:abc%d", i),
			DeploymentMode:     "docker",
			CreatedAt:          time.Now().Add(time.Duration(-i) * time.Hour),
			LastHealthCheck:    time.Now(),
			ContainerHealthy:   true,
			ApplicationHealthy: true,
		}
		require.NoError(t, store.SavePerformerState(ctx, performer.PerformerId, performer))
	}

	// List all performers
	allPerformers, err := store.ListPerformerStates(ctx)
	require.NoError(t, err)
	assert.Len(t, allPerformers, 5)

	// Update a performer
	updatedPerformer := &storage.PerformerState{
		PerformerId:        "performer-0",
		AvsAddress:         avsAddress,
		ContainerId:        "container-0-new",
		Status:             "failed",
		ArtifactRegistry:   "registry.io/avs",
		ArtifactTag:        "v0.1.0",
		ArtifactDigest:     "sha256:def0",
		DeploymentMode:     "docker",
		CreatedAt:          allPerformers[0].CreatedAt,
		LastHealthCheck:    time.Now(),
		ContainerHealthy:   false,
		ApplicationHealthy: false,
	}
	require.NoError(t, store.SavePerformerState(ctx, "performer-0", updatedPerformer))

	// Verify update
	retrieved, err := store.GetPerformerState(ctx, "performer-0")
	require.NoError(t, err)
	assert.Equal(t, "failed", retrieved.Status)
	assert.Equal(t, "container-0-new", retrieved.ContainerId)
	assert.False(t, retrieved.ContainerHealthy)

	// Delete a performer
	require.NoError(t, store.DeletePerformerState(ctx, "performer-0"))

	// Verify deletion
	_, err = store.GetPerformerState(ctx, "performer-0")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// List should now have 4
	remainingPerformers, err := store.ListPerformerStates(ctx)
	require.NoError(t, err)
	assert.Len(t, remainingPerformers, 4)
}

func TestExecutorStorageIsolation(t *testing.T) {
	ctx := context.Background()

	// Create two separate stores to simulate two executor instances
	store1 := memory.NewInMemoryExecutorStore()
	store2 := memory.NewInMemoryExecutorStore()

	// Add data to store1
	performer1 := &storage.PerformerState{
		PerformerId:        "isolated-performer-1",
		AvsAddress:         "0xavs1",
		ContainerId:        "container-1",
		Status:             "running",
		ArtifactRegistry:   "registry.io/avs1",
		ArtifactTag:        "v1.0.0",
		ArtifactDigest:     "sha256:isolated1",
		DeploymentMode:     "docker",
		CreatedAt:          time.Now(),
		LastHealthCheck:    time.Now(),
		ContainerHealthy:   true,
		ApplicationHealthy: true,
	}
	require.NoError(t, store1.SavePerformerState(ctx, performer1.PerformerId, performer1))

	// Add data to store2
	performer2 := &storage.PerformerState{
		PerformerId:        "isolated-performer-2",
		AvsAddress:         "0xavs2",
		ContainerId:        "container-2",
		Status:             "running",
		ArtifactRegistry:   "registry.io/avs2",
		ArtifactTag:        "v2.0.0",
		ArtifactDigest:     "sha256:isolated2",
		DeploymentMode:     "kubernetes",
		CreatedAt:          time.Now(),
		LastHealthCheck:    time.Now(),
		ContainerHealthy:   true,
		ApplicationHealthy: true,
	}
	require.NoError(t, store2.SavePerformerState(ctx, performer2.PerformerId, performer2))

	// Verify isolation - store1 shouldn't have store2's data
	_, err := store1.GetPerformerState(ctx, "isolated-performer-2")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Verify isolation - store2 shouldn't have store1's data
	_, err = store2.GetPerformerState(ctx, "isolated-performer-1")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Each store should only have its own data
	list1, err := store1.ListPerformerStates(ctx)
	require.NoError(t, err)
	assert.Len(t, list1, 1)
	assert.Equal(t, "isolated-performer-1", list1[0].PerformerId)

	list2, err := store2.ListPerformerStates(ctx)
	require.NoError(t, err)
	assert.Len(t, list2, 1)
	assert.Equal(t, "isolated-performer-2", list2[0].PerformerId)
}

func TestTaskDeduplicationMemoryPassthrough(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryExecutorStore()

	// Test that memory storage doesn't track processed tasks (pass-through behavior)
	taskId := "dedup-task-1"

	// Initially not processed
	processed, err := store.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.False(t, processed, "should always return false for memory storage")

	// Mark as processed (no-op for memory storage)
	err = store.MarkTaskProcessed(ctx, taskId)
	require.NoError(t, err)

	// Should still return false (pass-through)
	processed, err = store.IsTaskProcessed(ctx, taskId)
	require.NoError(t, err)
	assert.False(t, processed, "memory storage should not track processed tasks")

	// Mark again (idempotent no-op)
	err = store.MarkTaskProcessed(ctx, taskId)
	require.NoError(t, err)

	// Test multiple tasks - all should return false
	tasks := []string{"task-a", "task-b", "task-c"}
	for _, tid := range tasks {
		err = store.MarkTaskProcessed(ctx, tid)
		require.NoError(t, err)
	}

	// All should return false (not tracked)
	for _, tid := range tasks {
		processed, err = store.IsTaskProcessed(ctx, tid)
		require.NoError(t, err)
		assert.False(t, processed, "memory storage should not track task %s", tid)
	}

	// Non-processed task should also return false
	processed, err = store.IsTaskProcessed(ctx, "never-processed")
	require.NoError(t, err)
	assert.False(t, processed)
}
