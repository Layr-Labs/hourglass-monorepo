package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSuite defines a test suite that all storage implementations must pass
type TestSuite struct {
	NewStore func() (ExecutorStore, error)
}

// Run executes all storage interface compliance tests
func (s *TestSuite) Run(t *testing.T) {
	t.Run("PerformerState", s.testPerformerState)
	t.Run("Lifecycle", s.testLifecycle)
	t.Run("ProcessedTasks", s.testProcessedTasks)
	t.Run("ConcurrentAccess", s.testConcurrentAccess)
}

func (s *TestSuite) testPerformerState(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	performerState := &PerformerState{
		PerformerId:        "performer-123",
		AvsAddress:         "0xavs123",
		ResourceId:         "container-abc",
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
	assert.ErrorIs(t, err, ErrNotFound)

	// Test saving and getting performer state
	err = store.SavePerformerState(ctx, performerState.PerformerId, performerState)
	require.NoError(t, err)

	retrieved, err := store.GetPerformerState(ctx, performerState.PerformerId)
	require.NoError(t, err)
	assert.Equal(t, performerState.PerformerId, retrieved.PerformerId)
	assert.Equal(t, performerState.AvsAddress, retrieved.AvsAddress)
	assert.Equal(t, performerState.ResourceId, retrieved.ResourceId)
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
	assert.ErrorIs(t, err, ErrNotFound)

	// Test deleting non-existent performer
	err = store.DeletePerformerState(ctx, "non-existent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func (s *TestSuite) testLifecycle(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)

	ctx := context.Background()

	// Add some data
	performerState := &PerformerState{
		PerformerId: "test-performer",
		AvsAddress:  "0xavs123",
		Status:      "running",
		CreatedAt:   time.Now(),
	}
	err = store.SavePerformerState(ctx, performerState.PerformerId, performerState)
	require.NoError(t, err)

	// Close the store
	err = store.Close()
	require.NoError(t, err)

	// Operations after close should fail
	err = store.SavePerformerState(ctx, "new-performer", performerState)
	assert.ErrorIs(t, err, ErrStoreClosed)

	_, err = store.GetPerformerState(ctx, performerState.PerformerId)
	assert.ErrorIs(t, err, ErrStoreClosed)
}

func (s *TestSuite) testProcessedTasks(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test checking non-processed task
	_, err = store.IsTaskProcessed(ctx, "task-123")
	require.NoError(t, err)

	// Test marking task as processed
	err = store.MarkTaskProcessed(ctx, "task-123")
	require.NoError(t, err)

	// Test checking after marking
	_, err = store.IsTaskProcessed(ctx, "task-123")
	require.NoError(t, err)

	// Test marking same task again (should be idempotent)
	err = store.MarkTaskProcessed(ctx, "task-123")
	require.NoError(t, err)

	// Test empty task ID validation
	err = store.MarkTaskProcessed(ctx, "")
	assert.Error(t, err, "empty task ID should return error")
}

func (s *TestSuite) testConcurrentAccess(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	done := make(chan bool)
	errors := make(chan error, 10)

	// Concurrent writes to different performers
	for i := 0; i < 5; i++ {
		go func(id int) {
			performerState := &PerformerState{
				PerformerId: fmt.Sprintf("performer-%d", id),
				AvsAddress:  fmt.Sprintf("0xavs%d", id),
				Status:      "running",
				CreatedAt:   time.Now(),
			}
			for j := 0; j < 10; j++ {
				performerState.Status = fmt.Sprintf("status-%d", j)
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
			performerId := fmt.Sprintf("performer-%d", id)
			for j := 0; j < 10; j++ {
				_, err := store.GetPerformerState(ctx, performerId)
				if err != nil && err != ErrNotFound {
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
