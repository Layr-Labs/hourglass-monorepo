package persistence_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregatorCrashRecovery(t *testing.T) {
	ctx := context.Background()

	// Create in-memory store
	store := memory.NewInMemoryAggregatorStore()

	// Simulate some state before "crash"
	avsAddress := "0xtest"
	chainId := config.ChainId(1) // Ethereum mainnet
	require.NoError(t, store.SetLastProcessedBlock(ctx, avsAddress, chainId, 1000))

	// Save some tasks
	deadline1 := time.Now().Add(1 * time.Hour)
	task1 := &types.Task{
		TaskId:                 "task-1",
		AVSAddress:             "0xavs1",
		OperatorSetId:          1,
		CallbackAddr:           "0xcallback",
		DeadlineUnixSeconds:    &deadline1,
		ThresholdBips:          6700,
		Payload:                []byte("test payload 1"),
		ChainId:                chainId,
		SourceBlockNumber:      990,
		L1ReferenceBlockNumber: 990,
		ReferenceTimestamp:     990,
		BlockHash:              "0xhash1",
	}
	require.NoError(t, store.SavePendingTask(ctx, task1))
	require.NoError(t, store.UpdateTaskStatus(ctx, "task-1", storage.TaskStatusProcessing))

	deadline2 := time.Now().Add(2 * time.Hour)
	task2 := &types.Task{
		TaskId:                 "task-2",
		AVSAddress:             "0xavs1",
		OperatorSetId:          1,
		CallbackAddr:           "0xcallback",
		DeadlineUnixSeconds:    &deadline2,
		ThresholdBips:          6700,
		Payload:                []byte("test payload 2"),
		ChainId:                chainId,
		SourceBlockNumber:      995,
		L1ReferenceBlockNumber: 995,
		ReferenceTimestamp:     995,
		BlockHash:              "0xhash2",
	}
	require.NoError(t, store.SavePendingTask(ctx, task2))

	// Simulate "crash" by creating new aggregator with same store
	// Note: In real usage, aggregator would be created with all dependencies
	// Here we're just verifying that storage preserved the state correctly

	// Verify state was preserved
	recoveredBlock, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), recoveredBlock)

	pendingTasks, err := store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, pendingTasks, 1) // Only task-2 should be pending
	assert.Equal(t, "task-2", pendingTasks[0].TaskId)

	recoveredTask1, err := store.GetTask(ctx, "task-1")
	require.NoError(t, err)
	assert.Equal(t, "task-1", recoveredTask1.TaskId)
}

func TestAggregatorTaskFlowWithPersistence(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()

	// Test task lifecycle
	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:                 "task-flow-1",
		AVSAddress:             "0xavs2",
		OperatorSetId:          1,
		CallbackAddr:           "0xcallback",
		DeadlineUnixSeconds:    &deadline,
		ThresholdBips:          6700,
		Payload:                []byte("test payload"),
		ChainId:                config.ChainId(1),
		SourceBlockNumber:      1000,
		L1ReferenceBlockNumber: 1000,
		ReferenceTimestamp:     1000,
		BlockHash:              "0xhash",
	}

	// Save task
	require.NoError(t, store.SavePendingTask(ctx, task))

	// Update status through lifecycle
	require.NoError(t, store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing))

	retrievedTask, err := store.GetTask(ctx, task.TaskId)
	require.NoError(t, err)
	assert.Equal(t, task.TaskId, retrievedTask.TaskId)

	// Complete task
	require.NoError(t, store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted))

	// Verify pending tasks doesn't include completed
	pendingTasks, err := store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, pendingTasks, 0)

	// Delete completed task
	require.NoError(t, store.DeleteTask(ctx, task.TaskId))

	// Verify deletion
	_, err = store.GetTask(ctx, task.TaskId)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestAggregatorConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()

	// Run concurrent operations
	done := make(chan bool)
	errors := make(chan error, 100)

	// Writer goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				deadline := time.Now().Add(time.Hour)
				task := &types.Task{
					TaskId:                 fmt.Sprintf("concurrent-task-%d-%d", id, j),
					AVSAddress:             fmt.Sprintf("0xavs%d", id),
					OperatorSetId:          uint32(id),
					CallbackAddr:           "0xcallback",
					DeadlineUnixSeconds:    &deadline,
					ThresholdBips:          6700,
					Payload:                []byte(fmt.Sprintf("payload-%d-%d", id, j)),
					ChainId:                config.ChainId(1),
					SourceBlockNumber:      uint64(1000 + j),
					L1ReferenceBlockNumber: uint64(1000 + j),
					ReferenceTimestamp:     1000,
					BlockHash:              fmt.Sprintf("0xhash%d%d", id, j),
				}
				if err := store.SavePendingTask(ctx, task); err != nil {
					errors <- err
				}
			}
			done <- true
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 200; j++ {
				_, _ = store.ListPendingTasks(ctx)
				time.Sleep(time.Microsecond * 10)
			}
			done <- true
		}()
	}

	// Wait for completion
	for i := 0; i < 15; i++ {
		<-done
	}

	close(errors)

	// Check for errors
	for err := range errors {
		t.Fatalf("Concurrent operation failed: %v", err)
	}

	// Verify all tasks were saved
	allTasks, err := store.ListPendingTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, allTasks, 1000) // 10 writers * 100 tasks each
}

func TestChainPollerRecovery(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()

	// Test multiple chains
	chains := []config.ChainId{
		config.ChainId(1),     // Ethereum
		config.ChainId(8453),  // Base
		config.ChainId(42161), // Arbitrum
	}

	// Set different block heights
	avsAddress := "0xtest"
	for i, chain := range chains {
		blockNum := uint64(1000 * (i + 1))
		require.NoError(t, store.SetLastProcessedBlock(ctx, avsAddress, chain, blockNum))
	}

	// Simulate recovery
	for i, chain := range chains {
		expectedBlock := uint64(1000 * (i + 1))
		recoveredBlock, err := store.GetLastProcessedBlock(ctx, avsAddress, chain)
		require.NoError(t, err)
		assert.Equal(t, expectedBlock, recoveredBlock, "Chain %d block mismatch", chain)
	}

	// Test non-existent chain
	nonExistentChain := config.ChainId(99999)
	block, err := store.GetLastProcessedBlock(ctx, avsAddress, nonExistentChain)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Equal(t, uint64(0), block)
}
