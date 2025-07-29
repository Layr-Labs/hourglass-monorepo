package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StorageTestSuite defines a test suite that all storage implementations must pass
type StorageTestSuite struct {
	NewStore func() (storage.ExecutorStore, error)
}

// TestExecutorStore runs all storage interface compliance tests
func (s *StorageTestSuite) TestExecutorStore(t *testing.T) {
	t.Run("PerformerState", s.testPerformerState)
	t.Run("TaskTracking", s.testTaskTracking)
	t.Run("DeploymentTracking", s.testDeploymentTracking)
	t.Run("Lifecycle", s.testLifecycle)
	t.Run("ConcurrentAccess", s.testConcurrentAccess)
}

func (s *StorageTestSuite) testPerformerState(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	performerState := &storage.PerformerState{
		PerformerId:        "performer-123",
		AvsAddress:         "0xavs123",
		ContainerId:        "container-abc",
		Status:             "running",
		ArtifactRegistry:   "registry.io/avs",
		ArtifactDigest:     "sha256:abcdef",
		ArtifactTag:        "v1.0.0",
		DeploymentMode:     "docker",
		CreatedAt:          time.Now(),
		LastHealthCheck:    time.Now(),
		ContainerHealthy:   true,
		ApplicationHealthy: true,
	}

	// Test getting non-existent performer
	_, err = store.GetPerformerState(ctx, performerState.PerformerId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test saving and getting performer state
	err = store.SavePerformerState(ctx, performerState.PerformerId, performerState)
	require.NoError(t, err)

	retrieved, err := store.GetPerformerState(ctx, performerState.PerformerId)
	require.NoError(t, err)
	assert.Equal(t, performerState.PerformerId, retrieved.PerformerId)
	assert.Equal(t, performerState.AvsAddress, retrieved.AvsAddress)
	assert.Equal(t, performerState.ContainerId, retrieved.ContainerId)
	assert.Equal(t, performerState.Status, retrieved.Status)

	// Test listing performer states
	states, err := store.ListPerformerStates(ctx)
	require.NoError(t, err)
	assert.Len(t, states, 1)
	assert.Equal(t, performerState.PerformerId, states[0].PerformerId)

	// Test updating performer state
	performerState.Status = "stopped"
	performerState.ContainerHealthy = false
	err = store.SavePerformerState(ctx, performerState.PerformerId, performerState)
	require.NoError(t, err)

	retrieved, err = store.GetPerformerState(ctx, performerState.PerformerId)
	require.NoError(t, err)
	assert.Equal(t, "stopped", retrieved.Status)
	assert.False(t, retrieved.ContainerHealthy)

	// Test deleting performer state
	err = store.DeletePerformerState(ctx, performerState.PerformerId)
	require.NoError(t, err)

	_, err = store.GetPerformerState(ctx, performerState.PerformerId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test deleting non-existent performer
	err = store.DeletePerformerState(ctx, "non-existent")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func (s *StorageTestSuite) testTaskTracking(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	taskInfo := &storage.TaskInfo{
		TaskId:            "task-123",
		AvsAddress:        "0xavs123",
		OperatorAddress:   "0xoperator123",
		ReceivedAt:        time.Now(),
		Status:            "processing",
		AggregatorAddress: "0xaggregator123",
		OperatorSetId:     1,
	}

	// Test getting non-existent task
	_, err = store.GetInflightTask(ctx, taskInfo.TaskId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test saving and getting inflight task
	err = store.SaveInflightTask(ctx, taskInfo.TaskId, taskInfo)
	require.NoError(t, err)

	retrieved, err := store.GetInflightTask(ctx, taskInfo.TaskId)
	require.NoError(t, err)
	assert.Equal(t, taskInfo.TaskId, retrieved.TaskId)
	assert.Equal(t, taskInfo.AvsAddress, retrieved.AvsAddress)
	assert.Equal(t, taskInfo.Status, retrieved.Status)

	// Test listing inflight tasks
	tasks, err := store.ListInflightTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, taskInfo.TaskId, tasks[0].TaskId)

	// Test updating task
	taskInfo.Status = "completed"
	err = store.SaveInflightTask(ctx, taskInfo.TaskId, taskInfo)
	require.NoError(t, err)

	retrieved, err = store.GetInflightTask(ctx, taskInfo.TaskId)
	require.NoError(t, err)
	assert.Equal(t, "completed", retrieved.Status)

	// Test deleting task
	err = store.DeleteInflightTask(ctx, taskInfo.TaskId)
	require.NoError(t, err)

	_, err = store.GetInflightTask(ctx, taskInfo.TaskId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test deleting non-existent task
	err = store.DeleteInflightTask(ctx, "non-existent")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func (s *StorageTestSuite) testDeploymentTracking(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	deploymentInfo := &storage.DeploymentInfo{
		DeploymentId:     "deploy-123",
		AvsAddress:       "0xavs123",
		ArtifactRegistry: "registry.io/avs",
		ArtifactDigest:   "sha256:abcdef",
		Status:           storage.DeploymentStatusPending,
		StartedAt:        time.Now(),
		CompletedAt:      nil,
		Error:            "",
	}

	// Test getting non-existent deployment
	_, err = store.GetDeployment(ctx, deploymentInfo.DeploymentId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test saving and getting deployment
	err = store.SaveDeployment(ctx, deploymentInfo.DeploymentId, deploymentInfo)
	require.NoError(t, err)

	retrieved, err := store.GetDeployment(ctx, deploymentInfo.DeploymentId)
	require.NoError(t, err)
	assert.Equal(t, deploymentInfo.DeploymentId, retrieved.DeploymentId)
	assert.Equal(t, deploymentInfo.AvsAddress, retrieved.AvsAddress)
	assert.Equal(t, deploymentInfo.Status, retrieved.Status)

	// Test updating deployment status
	err = store.UpdateDeploymentStatus(ctx, deploymentInfo.DeploymentId, storage.DeploymentStatusDeploying)
	require.NoError(t, err)

	retrieved, err = store.GetDeployment(ctx, deploymentInfo.DeploymentId)
	require.NoError(t, err)
	assert.Equal(t, storage.DeploymentStatusDeploying, retrieved.Status)

	// Test updating to completed
	err = store.UpdateDeploymentStatus(ctx, deploymentInfo.DeploymentId, storage.DeploymentStatusRunning)
	require.NoError(t, err)

	retrieved, err = store.GetDeployment(ctx, deploymentInfo.DeploymentId)
	require.NoError(t, err)
	assert.Equal(t, storage.DeploymentStatusRunning, retrieved.Status)

	// Test updating non-existent deployment
	err = store.UpdateDeploymentStatus(ctx, "non-existent", storage.DeploymentStatusFailed)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func (s *StorageTestSuite) testLifecycle(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)

	ctx := context.Background()

	// Add some data
	performerState := &storage.PerformerState{
		PerformerId: "test-performer",
		AvsAddress:  "0xavs123",
		Status:      "running",
	}
	err = store.SavePerformerState(ctx, performerState.PerformerId, performerState)
	require.NoError(t, err)

	// Close the store
	err = store.Close()
	require.NoError(t, err)

	// Operations after close should fail
	err = store.SavePerformerState(ctx, "new-performer", performerState)
	assert.ErrorIs(t, err, storage.ErrStoreClosed)

	_, err = store.GetPerformerState(ctx, performerState.PerformerId)
	assert.ErrorIs(t, err, storage.ErrStoreClosed)
}

func (s *StorageTestSuite) testConcurrentAccess(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	done := make(chan bool)
	errors := make(chan error, 10)

	// Concurrent writes to different performers
	for i := 0; i < 5; i++ {
		go func(id int) {
			performerState := &storage.PerformerState{
				PerformerId: t.Name() + "-performer-" + string(rune(id)),
				AvsAddress:  "0xavs" + string(rune(id)),
				Status:      "running",
			}
			for j := 0; j < 10; j++ {
				performerState.Status = "status-" + string(rune(j))
				err := store.SavePerformerState(ctx, performerState.PerformerId, performerState)
				if err != nil {
					errors <- err
					return
				}
			}
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func(id int) {
			performerId := t.Name() + "-performer-" + string(rune(id))
			for j := 0; j < 10; j++ {
				_, err := store.GetPerformerState(ctx, performerId)
				if err != nil && err != storage.ErrNotFound {
					errors <- err
					return
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case err := <-errors:
			t.Fatalf("Concurrent access error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}