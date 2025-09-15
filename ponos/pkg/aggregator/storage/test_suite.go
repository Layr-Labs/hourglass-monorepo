package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSuite defines a test suite that all storage implementations must pass
type TestSuite struct {
	NewStore func() (AggregatorStore, error)
}

// Run executes all storage interface compliance tests
func (s *TestSuite) Run(t *testing.T) {
	t.Run("ChainPollingState", s.testChainPollingState)
	t.Run("TaskManagement", s.testTaskManagement)
	t.Run("Lifecycle", s.testLifecycle)
	t.Run("ConcurrentAccess", s.testConcurrentAccess)
}

func (s *TestSuite) testChainPollingState(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	avsAddress := "0xavs123"
	chainId := config.ChainId(1)

	// Test getting non-existent block
	_, err = store.GetLastProcessedBlock(ctx, avsAddress, chainId)
	assert.ErrorIs(t, err, ErrNotFound)

	// Test setting and getting block
	blockNum := uint64(12345)
	blockRecord := &BlockRecord{
		Number:     blockNum,
		Hash:       "0xhash12345",
		ParentHash: "0xparent12344",
		Timestamp:  1234567890,
		ChainId:    chainId,
	}
	err = store.SaveBlock(ctx, avsAddress, blockRecord)
	require.NoError(t, err)

	retrieved, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	assert.Equal(t, blockNum, retrieved.Number)

	// Test updating block
	newBlockNum := uint64(12346)
	newBlockRecord := &BlockRecord{
		Number:     newBlockNum,
		Hash:       "0xhash12346",
		ParentHash: "0xhash12345",
		Timestamp:  1234567891,
		ChainId:    chainId,
	}
	err = store.SaveBlock(ctx, avsAddress, newBlockRecord)
	require.NoError(t, err)

	retrieved, err = store.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	assert.Equal(t, newBlockNum, retrieved.Number)

	// Test multiple chains for same AVS
	chainId2 := config.ChainId(2)
	blockNum2 := uint64(54321)
	blockRecord2 := &BlockRecord{
		Number:     blockNum2,
		Hash:       "0xhash54321",
		ParentHash: "0xparent54320",
		Timestamp:  1234567892,
		ChainId:    chainId2,
	}
	err = store.SaveBlock(ctx, avsAddress, blockRecord2)
	require.NoError(t, err)

	retrieved2, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId2)
	require.NoError(t, err)
	assert.Equal(t, blockNum2, retrieved2.Number)

	// Ensure first chain is unchanged
	retrieved, err = store.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	assert.Equal(t, newBlockNum, retrieved.Number)

	// Test different AVS on same chain
	avsAddress2 := "0xavs456"
	blockNum3 := uint64(99999)
	blockRecord3 := &BlockRecord{
		Number:     blockNum3,
		Hash:       "0xhash99999",
		ParentHash: "0xparent99998",
		Timestamp:  1234567893,
		ChainId:    chainId,
	}
	err = store.SaveBlock(ctx, avsAddress2, blockRecord3)
	require.NoError(t, err)

	retrieved3, err := store.GetLastProcessedBlock(ctx, avsAddress2, chainId)
	require.NoError(t, err)
	assert.Equal(t, blockNum3, retrieved3.Number)

	// Ensure original AVS on same chain is unchanged
	retrieved, err = store.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	assert.Equal(t, newBlockNum, retrieved.Number)
}

func (s *TestSuite) testTaskManagement(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create test task
	deadline := time.Now().Add(time.Hour)
	task := &types.Task{
		TaskId:                 "task-123",
		AVSAddress:             "0xavs123",
		Payload:                []byte("test payload"),
		ChainId:                config.ChainId(1),
		SourceBlockNumber:      12345,
		L1ReferenceBlockNumber: 12345,
		OperatorSetId:          1,
		CallbackAddr:           "0xcallback123",
		ThresholdBips:          5000,
		DeadlineUnixSeconds:    &deadline,
		BlockHash:              "0xblockhash123",
	}

	// Test getting non-existent task
	_, err = store.GetTask(ctx, task.TaskId)
	assert.ErrorIs(t, err, ErrNotFound)

	// Test saving and getting task
	err = store.SavePendingTask(ctx, task)
	require.NoError(t, err)

	retrieved, err := store.GetTask(ctx, task.TaskId)
	require.NoError(t, err)
	assert.Equal(t, task.TaskId, retrieved.TaskId)
	assert.Equal(t, task.AVSAddress, retrieved.AVSAddress)
	assert.Equal(t, task.Payload, retrieved.Payload)

	// Test listing pending tasks
	pendingTasks, err := store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, pendingTasks, 1)
	assert.Equal(t, task.TaskId, pendingTasks[0].TaskId)

	// Test listing pending tasks for specific AVS
	avsPendingTasks, err := store.ListPendingTasksForAVS(ctx, "0xavs123")
	require.NoError(t, err)
	assert.Len(t, avsPendingTasks, 1)
	assert.Equal(t, task.TaskId, avsPendingTasks[0].TaskId)

	// Create another task for a different AVS
	task2 := &types.Task{
		TaskId:                 "task-456",
		AVSAddress:             "0xavs456",
		Payload:                []byte("test payload 2"),
		ChainId:                config.ChainId(1),
		SourceBlockNumber:      12346,
		L1ReferenceBlockNumber: 12345,
		OperatorSetId:          1,
		CallbackAddr:           "0xcallback456",
		ThresholdBips:          5000,
		DeadlineUnixSeconds:    &deadline,
		BlockHash:              "0xblockhash456",
	}
	err = store.SavePendingTask(ctx, task2)
	require.NoError(t, err)

	// Test listing all pending tasks (should have 2)
	allPendingTasks, err := store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, allPendingTasks, 2)

	// Test listing pending tasks for first AVS (should have 1)
	avs1Tasks, err := store.ListPendingTasksForAVS(ctx, "0xavs123")
	require.NoError(t, err)
	assert.Len(t, avs1Tasks, 1)
	assert.Equal(t, task.TaskId, avs1Tasks[0].TaskId)

	// Test listing pending tasks for second AVS (should have 1)
	avs2Tasks, err := store.ListPendingTasksForAVS(ctx, "0xavs456")
	require.NoError(t, err)
	assert.Len(t, avs2Tasks, 1)
	assert.Equal(t, task2.TaskId, avs2Tasks[0].TaskId)

	// Test listing pending tasks for non-existent AVS (should be empty)
	noAvsTasks, err := store.ListPendingTasksForAVS(ctx, "0xnonexistent")
	require.NoError(t, err)
	assert.Len(t, noAvsTasks, 0)

	// Test case insensitive AVS address matching
	avsUpperTasks, err := store.ListPendingTasksForAVS(ctx, "0xAVS123")
	require.NoError(t, err)
	assert.Len(t, avsUpperTasks, 1)
	assert.Equal(t, task.TaskId, avsUpperTasks[0].TaskId)

	// Test updating task status
	err = store.UpdateTaskStatus(ctx, task.TaskId, TaskStatusProcessing)
	require.NoError(t, err)

	// Pending tasks should now have only task2
	pendingTasks, err = store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, pendingTasks, 1)
	assert.Equal(t, task2.TaskId, pendingTasks[0].TaskId)

	// AVS-specific pending tasks should reflect the status change
	avs1PendingTasks, err := store.ListPendingTasksForAVS(ctx, "0xavs123")
	require.NoError(t, err)
	assert.Len(t, avs1PendingTasks, 0)

	// Test updating to completed
	err = store.UpdateTaskStatus(ctx, task.TaskId, TaskStatusCompleted)
	require.NoError(t, err)

	// Test deleting task
	err = store.DeleteTask(ctx, task.TaskId)
	require.NoError(t, err)

	_, err = store.GetTask(ctx, task.TaskId)
	assert.ErrorIs(t, err, ErrNotFound)

	// Test deleting non-existent task
	err = store.DeleteTask(ctx, "non-existent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func (s *TestSuite) testLifecycle(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)

	ctx := context.Background()

	// Add some data
	avsAddress := "0xavs123"
	blockRecord := &BlockRecord{
		Number:     12345,
		Hash:       "0xhash12345",
		ParentHash: "0xparent12344",
		Timestamp:  1234567890,
		ChainId:    config.ChainId(1),
	}
	err = store.SaveBlock(ctx, avsAddress, blockRecord)
	require.NoError(t, err)

	// Close the store
	err = store.Close()
	require.NoError(t, err)

	// Operations after close should fail
	newBlockRecord := &BlockRecord{
		Number:     12346,
		Hash:       "0xhash12346",
		ParentHash: "0xhash12345",
		Timestamp:  1234567891,
		ChainId:    config.ChainId(1),
	}
	err = store.SaveBlock(ctx, avsAddress, newBlockRecord)
	assert.ErrorIs(t, err, ErrStoreClosed)

	_, err = store.GetLastProcessedBlock(ctx, avsAddress, config.ChainId(1))
	assert.ErrorIs(t, err, ErrStoreClosed)
}

func (s *TestSuite) testConcurrentAccess(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	done := make(chan bool)
	errors := make(chan error, 10)

	avsAddress := "0xavs123"
	// Concurrent writes to different chains
	for i := 0; i < 5; i++ {
		go func(chainId config.ChainId) {
			for j := 0; j < 10; j++ {
				blockRecord := &BlockRecord{
					Number:     uint64(j),
					Hash:       fmt.Sprintf("0xhash%d_%d", chainId, j),
					ParentHash: fmt.Sprintf("0xparent%d_%d", chainId, j-1),
					Timestamp:  uint64(1234567890 + j),
					ChainId:    chainId,
				}
				err := store.SaveBlock(ctx, avsAddress, blockRecord)
				if err != nil {
					errors <- err
					return
				}
			}
			done <- true
		}(config.ChainId(i))
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func(chainId config.ChainId) {
			for j := 0; j < 10; j++ {
				_, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
				if err != nil && err != ErrNotFound {
					errors <- err
					return
				}
			}
			done <- true
		}(config.ChainId(i))
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
