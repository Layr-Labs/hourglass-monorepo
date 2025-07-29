package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StorageTestSuite defines a test suite that all storage implementations must pass
type StorageTestSuite struct {
	NewStore func() (storage.AggregatorStore, error)
}

// TestAggregatorStore runs all storage interface compliance tests
func (s *StorageTestSuite) TestAggregatorStore(t *testing.T) {
	t.Run("ChainPollingState", s.testChainPollingState)
	t.Run("TaskManagement", s.testTaskManagement)
	t.Run("ConfigCaching", s.testConfigCaching)
	t.Run("Lifecycle", s.testLifecycle)
	t.Run("ConcurrentAccess", s.testConcurrentAccess)
}

func (s *StorageTestSuite) testChainPollingState(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	chainId := config.ChainId(1)

	// Test getting non-existent block
	_, err = store.GetLastProcessedBlock(ctx, chainId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test setting and getting block
	blockNum := uint64(12345)
	err = store.SetLastProcessedBlock(ctx, chainId, blockNum)
	require.NoError(t, err)

	retrieved, err := store.GetLastProcessedBlock(ctx, chainId)
	require.NoError(t, err)
	assert.Equal(t, blockNum, retrieved)

	// Test updating block
	newBlockNum := uint64(12346)
	err = store.SetLastProcessedBlock(ctx, chainId, newBlockNum)
	require.NoError(t, err)

	retrieved, err = store.GetLastProcessedBlock(ctx, chainId)
	require.NoError(t, err)
	assert.Equal(t, newBlockNum, retrieved)

	// Test multiple chains
	chainId2 := config.ChainId(2)
	blockNum2 := uint64(54321)
	err = store.SetLastProcessedBlock(ctx, chainId2, blockNum2)
	require.NoError(t, err)

	retrieved2, err := store.GetLastProcessedBlock(ctx, chainId2)
	require.NoError(t, err)
	assert.Equal(t, blockNum2, retrieved2)

	// Ensure first chain is unchanged
	retrieved, err = store.GetLastProcessedBlock(ctx, chainId)
	require.NoError(t, err)
	assert.Equal(t, newBlockNum, retrieved)
}

func (s *StorageTestSuite) testTaskManagement(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create test task
	deadline := time.Now().Add(time.Hour)
	task := &types.Task{
		TaskId:              "task-123",
		AVSAddress:          "0xavs123",
		Payload:             []byte("test payload"),
		ChainId:             config.ChainId(1),
		BlockNumber:         12345,
		OperatorSetId:       1,
		CallbackAddr:        "0xcallback123",
		ThresholdBips:       5000,
		DeadlineUnixSeconds: &deadline,
		BlockHash:           "0xblockhash123",
	}

	// Test getting non-existent task
	_, err = store.GetTask(ctx, task.TaskId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test saving and getting task
	err = store.SaveTask(ctx, task)
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

	// Test updating task status
	err = store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing)
	require.NoError(t, err)

	// Pending tasks should now be empty
	pendingTasks, err = store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, pendingTasks, 0)

	// Test updating to completed
	err = store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted)
	require.NoError(t, err)

	// Test deleting task
	err = store.DeleteTask(ctx, task.TaskId)
	require.NoError(t, err)

	_, err = store.GetTask(ctx, task.TaskId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test deleting non-existent task
	err = store.DeleteTask(ctx, "non-existent")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func (s *StorageTestSuite) testConfigCaching(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test OperatorSetConfig
	avsAddress := "0xavs123"
	operatorSetId := uint32(1)
	opsetConfig := &storage.OperatorSetTaskConfig{
		TaskSLA:      3600,
		CurveType:    config.CurveTypeBN254,
		TaskMetadata: []byte("metadata"),
		Consensus: storage.OperatorSetTaskConsensus{
			ConsensusType: storage.ConsensusTypeStakeProportionThreshold,
			Threshold:     6666,
		},
	}

	// Test getting non-existent config
	_, err = store.GetOperatorSetConfig(ctx, avsAddress, operatorSetId)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test saving and getting config
	err = store.SaveOperatorSetConfig(ctx, avsAddress, operatorSetId, opsetConfig)
	require.NoError(t, err)

	retrieved, err := store.GetOperatorSetConfig(ctx, avsAddress, operatorSetId)
	require.NoError(t, err)
	assert.Equal(t, opsetConfig.TaskSLA, retrieved.TaskSLA)
	assert.Equal(t, opsetConfig.CurveType, retrieved.CurveType)
	assert.Equal(t, opsetConfig.TaskMetadata, retrieved.TaskMetadata)
	assert.Equal(t, opsetConfig.Consensus.ConsensusType, retrieved.Consensus.ConsensusType)
	assert.Equal(t, opsetConfig.Consensus.Threshold, retrieved.Consensus.Threshold)

	// Test AVSConfig
	avsConfig := &storage.AvsConfig{
		AggregatorOperatorSetId: 1,
		ExecutorOperatorSetIds:  []uint32{2, 3, 4},
		CurveType:               config.CurveTypeECDSA,
	}

	// Test getting non-existent AVS config
	_, err = store.GetAVSConfig(ctx, avsAddress)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// Test saving and getting AVS config
	err = store.SaveAVSConfig(ctx, avsAddress, avsConfig)
	require.NoError(t, err)

	retrievedAVS, err := store.GetAVSConfig(ctx, avsAddress)
	require.NoError(t, err)
	assert.Equal(t, avsConfig.AggregatorOperatorSetId, retrievedAVS.AggregatorOperatorSetId)
	assert.Equal(t, avsConfig.ExecutorOperatorSetIds, retrievedAVS.ExecutorOperatorSetIds)
	assert.Equal(t, avsConfig.CurveType, retrievedAVS.CurveType)
}

func (s *StorageTestSuite) testLifecycle(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)

	ctx := context.Background()

	// Add some data
	err = store.SetLastProcessedBlock(ctx, config.ChainId(1), 12345)
	require.NoError(t, err)

	// Close the store
	err = store.Close()
	require.NoError(t, err)

	// Operations after close should fail
	err = store.SetLastProcessedBlock(ctx, config.ChainId(1), 12346)
	assert.ErrorIs(t, err, storage.ErrStoreClosed)

	_, err = store.GetLastProcessedBlock(ctx, config.ChainId(1))
	assert.ErrorIs(t, err, storage.ErrStoreClosed)
}

func (s *StorageTestSuite) testConcurrentAccess(t *testing.T) {
	store, err := s.NewStore()
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	done := make(chan bool)
	errors := make(chan error, 10)

	// Concurrent writes to different chains
	for i := 0; i < 5; i++ {
		go func(chainId config.ChainId) {
			for j := 0; j < 10; j++ {
				err := store.SetLastProcessedBlock(ctx, chainId, uint64(j))
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
				_, err := store.GetLastProcessedBlock(ctx, chainId)
				if err != nil && err != storage.ErrNotFound {
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