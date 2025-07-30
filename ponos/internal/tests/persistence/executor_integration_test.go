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

	// Save inflight tasks
	task1 := &storage.TaskInfo{
		TaskId:            "task-1",
		AvsAddress:        "0xavs1",
		OperatorAddress:   "0xoperator1",
		ReceivedAt:        time.Now().Add(-2 * time.Minute),
		Status:            "processing",
		AggregatorAddress: "0xaggregator1",
		OperatorSetId:     1,
	}
	require.NoError(t, store.SaveInflightTask(ctx, task1.TaskId, task1))

	// Save deployment info
	deployment := &storage.DeploymentInfo{
		DeploymentId:     "deployment-1",
		AvsAddress:       "0xavs1",
		ArtifactRegistry: "registry.io/avs1",
		ArtifactDigest:   "sha256:abc123",
		Status:           storage.DeploymentStatusRunning,
		StartedAt:        time.Now().Add(-30 * time.Minute),
		CompletedAt:      nil,
		Error:            "",
	}
	require.NoError(t, store.SaveDeployment(ctx, deployment.DeploymentId, deployment))

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

	recoveredTasks, err := store.ListInflightTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, recoveredTasks, 1)
	assert.Equal(t, "task-1", recoveredTasks[0].TaskId)

	recoveredDeployment, err := store.GetDeployment(ctx, "deployment-1")
	require.NoError(t, err)
	assert.Equal(t, storage.DeploymentStatusRunning, recoveredDeployment.Status)
}

func TestExecutorDeploymentLifecycle(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryExecutorStore()

	deploymentId := "deployment-lifecycle-1"
	deployment := &storage.DeploymentInfo{
		DeploymentId:     deploymentId,
		AvsAddress:       "0xavs3",
		ArtifactRegistry: "registry.io/avs3",
		ArtifactDigest:   "sha256:xyz789",
		Status:           storage.DeploymentStatusPending,
		StartedAt:        time.Now(),
		CompletedAt:      nil,
		Error:            "",
	}

	// Save initial deployment
	require.NoError(t, store.SaveDeployment(ctx, deploymentId, deployment))

	// Update to deploying
	require.NoError(t, store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusDeploying))

	// Verify status update
	updated, err := store.GetDeployment(ctx, deploymentId)
	require.NoError(t, err)
	assert.Equal(t, storage.DeploymentStatusDeploying, updated.Status)

	// Update to running
	require.NoError(t, store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusRunning))

	// Verify final status
	final, err := store.GetDeployment(ctx, deploymentId)
	require.NoError(t, err)
	assert.Equal(t, storage.DeploymentStatusRunning, final.Status)

	// Test invalid status transition
	err = store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusPending)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid deployment status")
}

func TestExecutorConcurrentTaskOperations(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryExecutorStore()

	// Run concurrent operations
	done := make(chan bool)
	errors := make(chan error, 100)

	// Writer goroutines for tasks
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				task := &storage.TaskInfo{
					TaskId:            fmt.Sprintf("concurrent-task-%d-%d", id, j),
					AvsAddress:        fmt.Sprintf("0xavs%d", id),
					OperatorAddress:   "0xoperator1",
					ReceivedAt:        time.Now(),
					Status:            "processing",
					AggregatorAddress: "0xaggregator1",
					OperatorSetId:     uint32(id),
				}
				if err := store.SaveInflightTask(ctx, task.TaskId, task); err != nil {
					errors <- err
				}
			}
			done <- true
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = store.ListInflightTasks(ctx)
				time.Sleep(time.Microsecond * 10)
			}
			done <- true
		}()
	}

	// Deleter goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			time.Sleep(time.Millisecond * 10) // Let some tasks be created first
			for j := 0; j < 20; j++ {
				taskId := fmt.Sprintf("concurrent-task-%d-%d", id, j)
				_ = store.DeleteInflightTask(ctx, taskId)
			}
			done <- true
		}(i)
	}

	// Wait for completion
	for i := 0; i < 20; i++ {
		<-done
	}

	close(errors)

	// Check for errors
	for err := range errors {
		t.Fatalf("Concurrent operation failed: %v", err)
	}

	// Verify remaining tasks
	remainingTasks, err := store.ListInflightTasks(ctx)
	require.NoError(t, err)
	t.Logf("Remaining tasks after concurrent operations: %d", len(remainingTasks))
	assert.True(t, len(remainingTasks) > 0 && len(remainingTasks) < 500)
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