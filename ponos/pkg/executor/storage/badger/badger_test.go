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

		// Save an inflight task
		task := &storage.TaskInfo{
			TaskId:            "task-123",
			AvsAddress:        "0xAVS1",
			OperatorAddress:   "0xOperator1",
			ReceivedAt:        time.Now(),
			Status:            "processing",
			AggregatorAddress: "0xAggregator1",
			OperatorSetId:     1,
		}
		err = store.SaveInflightTask(ctx, "task-123", task)
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

		// Verify task exists
		retrievedTask, err := store.GetInflightTask(ctx, "task-123")
		require.NoError(t, err)
		assert.Equal(t, "task-123", retrievedTask.TaskId)
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

func TestBadgerExecutorStore_DeploymentStatusTransitions(t *testing.T) {
	// Test deployment status transitions
	tmpDir, err := os.MkdirTemp("", "badger-executor-deploy-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := &executorConfig.BadgerConfig{
		Dir: tmpDir,
	}

	store, err := NewBadgerExecutorStore(cfg)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	deploymentId := "deploy-123"

	// Save initial deployment
	deployment := &storage.DeploymentInfo{
		DeploymentId:     deploymentId,
		AvsAddress:       "0xAVS1",
		ArtifactRegistry: "registry.example.com",
		ArtifactDigest:   "sha256:abcdef",
		Status:           storage.DeploymentStatusPending,
		StartedAt:        time.Now(),
	}
	err = store.SaveDeployment(ctx, deploymentId, deployment)
	require.NoError(t, err)

	// Valid transition: pending -> deploying
	err = store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusDeploying)
	require.NoError(t, err)

	// Valid transition: deploying -> running
	err = store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusRunning)
	require.NoError(t, err)

	// Invalid transition: running -> pending
	err = store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusPending)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")

	// Valid transition: running -> failed
	err = store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusFailed)
	require.NoError(t, err)

	// Invalid transition: failed -> running (terminal state)
	err = store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusRunning)
	assert.Error(t, err)
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
