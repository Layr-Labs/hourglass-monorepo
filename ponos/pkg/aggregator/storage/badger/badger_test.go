package badger

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadgerAggregatorStore(t *testing.T) {
	// Create a temporary directory for BadgerDB
	tmpDir, err := os.MkdirTemp("", "badger-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Run the reusable test suite
	suite := &storage.TestSuite{
		NewStore: func() (storage.AggregatorStore, error) {
			cfg := &aggregatorConfig.BadgerConfig{
				Dir: tmpDir,
			}
			return NewBadgerAggregatorStore(cfg)
		},
	}
	suite.Run(t)
}

func TestBadgerAggregatorStore_Persistence(t *testing.T) {
	// Test that data persists across store restarts
	tmpDir, err := os.MkdirTemp("", "badger-persist-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := &aggregatorConfig.BadgerConfig{
		Dir: tmpDir,
	}

	ctx := context.Background()
	taskId := "test-task-1"

	// Create store, save data, and close
	{
		store, err := NewBadgerAggregatorStore(cfg)
		require.NoError(t, err)

		// Save a task
		deadline := time.Now().Add(time.Hour)
		task := &types.Task{
			TaskId:              taskId,
			AVSAddress:          "0xAVS1",
			Payload:             []byte("test payload"),
			ChainId:             config.ChainId(1),
			SourceBlockNumber:   12345,
			OperatorSetId:       1,
			CallbackAddr:        "0xcallback123",
			ThresholdBips:       5000,
			DeadlineUnixSeconds: &deadline,
			BlockHash:           "0xblockhash123",
		}
		err = store.SavePendingTask(ctx, task)
		require.NoError(t, err)

		// Set last processed block
		avsAddress := "0xtest"
		err = store.SetLastProcessedBlock(ctx, avsAddress, config.ChainId(1), 12345)
		require.NoError(t, err)

		// Close store
		err = store.Close()
		require.NoError(t, err)
	}

	// Reopen store and verify data persists
	{
		store, err := NewBadgerAggregatorStore(cfg)
		require.NoError(t, err)
		defer store.Close()

		// Verify task exists
		retrievedTask, err := store.GetTask(ctx, taskId)
		require.NoError(t, err)
		assert.Equal(t, taskId, retrievedTask.TaskId)

		// Verify block number
		avsAddress := "0xtest"
		blockNum, err := store.GetLastProcessedBlock(ctx, avsAddress, config.ChainId(1))
		require.NoError(t, err)
		assert.Equal(t, uint64(12345), blockNum)
	}
}

func TestBadgerAggregatorStore_InMemory(t *testing.T) {
	// Test in-memory mode
	cfg := &aggregatorConfig.BadgerConfig{
		Dir:      "",
		InMemory: true,
	}

	store, err := NewBadgerAggregatorStore(cfg)
	require.NoError(t, err)
	defer store.Close()

	// Run basic operations
	ctx := context.Background()
	deadline := time.Now().Add(time.Hour)
	task := &types.Task{
		TaskId:                 "task-1",
		AVSAddress:             "0xAVS1",
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
	err = store.SavePendingTask(ctx, task)
	require.NoError(t, err)

	retrieved, err := store.GetTask(ctx, "task-1")
	require.NoError(t, err)
	assert.Equal(t, task.TaskId, retrieved.TaskId)
}

func TestBadgerAggregatorStore_LargeDataSet(t *testing.T) {
	// Test with a large number of tasks
	tmpDir, err := os.MkdirTemp("", "badger-large-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := &aggregatorConfig.BadgerConfig{
		Dir: tmpDir,
	}

	store, err := NewBadgerAggregatorStore(cfg)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	numTasks := 1000

	// Save many tasks
	for i := 0; i < numTasks; i++ {
		deadline := time.Now().Add(time.Hour)
		task := &types.Task{
			TaskId:                 fmt.Sprintf("task-%d", i),
			AVSAddress:             "0xAVS1",
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
		err := store.SavePendingTask(ctx, task)
		require.NoError(t, err)
	}

	// List pending tasks
	pendingTasks, err := store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Equal(t, numTasks, len(pendingTasks))

	// Update half to processing
	for i := 0; i < numTasks/2; i++ {
		err := store.UpdateTaskStatus(ctx, fmt.Sprintf("task-%d", i), storage.TaskStatusProcessing)
		require.NoError(t, err)
	}

	// Verify pending count
	pendingTasks, err = store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Equal(t, numTasks/2, len(pendingTasks))
}

func BenchmarkBadgerAggregatorStore(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "badger-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	cfg := &aggregatorConfig.BadgerConfig{
		Dir: tmpDir,
	}

	store, err := NewBadgerAggregatorStore(cfg)
	require.NoError(b, err)
	defer store.Close()

	ctx := context.Background()

	b.Run("SavePendingTask", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			deadline := time.Now().Add(time.Hour)
			task := &types.Task{
				TaskId:                 fmt.Sprintf("task-bench-%d", i),
				AVSAddress:             "0xAVS1",
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
			_ = store.SavePendingTask(ctx, task)
		}
	})

	b.Run("GetTask", func(b *testing.B) {
		// Pre-populate some tasks
		for i := 0; i < 100; i++ {
			deadline := time.Now().Add(time.Hour)
			task := &types.Task{
				TaskId:                 fmt.Sprintf("task-get-%d", i),
				AVSAddress:             "0xAVS1",
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
			_ = store.SavePendingTask(ctx, task)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetTask(ctx, fmt.Sprintf("task-get-%d", i%100))
		}
	})

	b.Run("ListPendingTasks", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.ListPendingTasks(ctx)
		}
	})
}
