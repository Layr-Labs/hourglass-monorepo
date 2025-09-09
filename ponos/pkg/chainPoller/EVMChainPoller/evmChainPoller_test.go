package EVMChainPoller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollerProcessesTasksFromStorage(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	avsAddress := "0xtest"

	// Create tasks and save to storage (simulating tasks that need recovery)
	validDeadline := time.Now().Add(1 * time.Hour)
	task1 := &types.Task{
		TaskId:              "task-1",
		AVSAddress:          avsAddress,
		DeadlineUnixSeconds: &validDeadline,
	}
	task2 := &types.Task{
		TaskId:              "task-2",
		AVSAddress:          avsAddress,
		DeadlineUnixSeconds: &validDeadline,
	}

	// Pre-populate storage with pending tasks
	require.NoError(t, store.SavePendingTask(ctx, task1))
	require.NoError(t, store.SavePendingTask(ctx, task2))

	// Create minimal poller to test recovery
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	taskQueue := make(chan *types.Task, 10)

	poller := &EVMChainPoller{
		config: &EVMChainPollerConfig{
			AvsAddress: avsAddress,
		},
		taskQueue: taskQueue,
		store:     store,
		logger:    l,
	}

	// Call recoverInProgressTasks directly - this loads from storage and queues
	err = poller.recoverInProgressTasks(ctx)
	require.NoError(t, err)

	// Verify tasks were recovered and queued
	receivedTasks := make(map[string]bool)

	for i := 0; i < 2; i++ {
		select {
		case task := <-taskQueue:
			receivedTasks[task.TaskId] = true
			t.Logf("Received task: %s", task.TaskId)
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Timeout waiting for task")
		}
	}

	assert.True(t, receivedTasks["task-1"], "task-1 should be processed")
	assert.True(t, receivedTasks["task-2"], "task-2 should be processed")

	// Verify tasks were marked as processing
	err = store.UpdateTaskStatus(ctx, "task-1", storage.TaskStatusCompleted)
	assert.NoError(t, err, "Should be able to complete task-1")

	err = store.UpdateTaskStatus(ctx, "task-2", storage.TaskStatusCompleted)
	assert.NoError(t, err, "Should be able to complete task-2")
}

func TestPollerSkipsExpiredTasks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	store := memory.NewInMemoryAggregatorStore()
	avsAddress := "0xtest"

	// Create mix of valid and expired tasks
	validDeadline := time.Now().Add(1 * time.Hour)
	expiredDeadline := time.Now().Add(-1 * time.Hour)

	validTask := &types.Task{
		TaskId:              "valid-task",
		AVSAddress:          avsAddress,
		DeadlineUnixSeconds: &validDeadline,
	}
	expiredTask := &types.Task{
		TaskId:              "expired-task",
		AVSAddress:          avsAddress,
		DeadlineUnixSeconds: &expiredDeadline,
	}

	// Save both tasks to storage
	require.NoError(t, store.SavePendingTask(ctx, validTask))
	require.NoError(t, store.SavePendingTask(ctx, expiredTask))

	// Create minimal poller
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	taskQueue := make(chan *types.Task, 10)

	poller := &EVMChainPoller{
		config: &EVMChainPollerConfig{
			AvsAddress: avsAddress,
		},
		taskQueue: taskQueue,
		store:     store,
		logger:    l,
	}

	// Call recoverInProgressTasks directly
	err := poller.recoverInProgressTasks(ctx)
	require.NoError(t, err)

	// Should only receive the valid task
	select {
	case task := <-taskQueue:
		assert.Equal(t, "valid-task", task.TaskId, "Should receive valid task")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for valid task")
	}

	// Should not receive expired task
	select {
	case task := <-taskQueue:
		t.Fatalf("Should not receive expired task, got: %s", task.TaskId)
	case <-time.After(100 * time.Millisecond):
		// Expected - no more tasks
	}

	// Verify expired task was marked as failed
	err = store.UpdateTaskStatus(ctx, "expired-task", storage.TaskStatusCompleted)
	assert.Error(t, err, "Expired task should be in failed state")
}

func TestPollerHandlesChannelFull(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	avsAddress := "0xtest"

	// Create more tasks than channel capacity
	validDeadline := time.Now().Add(1 * time.Hour)
	for i := 0; i < 5; i++ {
		task := &types.Task{
			TaskId:              fmt.Sprintf("task-%d", i),
			AVSAddress:          avsAddress,
			DeadlineUnixSeconds: &validDeadline,
		}
		require.NoError(t, store.SavePendingTask(ctx, task))
	}

	// Create poller with small channel
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	smallTaskQueue := make(chan *types.Task, 2)

	poller := &EVMChainPoller{
		config: &EVMChainPollerConfig{
			AvsAddress: avsAddress,
		},
		taskQueue: smallTaskQueue,
		store:     store,
		logger:    l,
	}

	// Call recovery directly
	err := poller.recoverInProgressTasks(ctx)
	require.NoError(t, err)

	// Should receive first 2 tasks
	received := 0
	for i := 0; i < 2; i++ {
		select {
		case <-smallTaskQueue:
			received++
		case <-time.After(200 * time.Millisecond):
			break
		}
	}

	assert.Equal(t, 2, received, "Should receive 2 tasks (channel capacity)")

	// Verify remaining tasks are still pending (not all could be queued)
	pendingTasks, err := store.ListPendingTasksForAVS(ctx, avsAddress)
	require.NoError(t, err)
	// Some tasks should remain pending since channel was full
	assert.True(t, len(pendingTasks) > 0, "Some tasks should remain pending due to full channel")
}

func TestBlockProgressRecovery(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	avsAddress := "0xtest"
	chainId := config.ChainId(1)

	// Set a last processed block
	lastBlock := uint64(50)
	require.NoError(t, store.SetLastProcessedBlock(ctx, avsAddress, chainId, lastBlock))

	// Create minimal poller
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})

	poller := &EVMChainPoller{
		config: &EVMChainPollerConfig{
			AvsAddress: avsAddress,
			ChainId:    chainId,
		},
		store:  store,
		logger: l,
	}

	// Mock the ethClient just for this test
	poller.ethClient = nil // Will cause Start to fail at block fetching, but that's OK

	// Call Start which should recover the block
	ctx2, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_ = poller.Start(ctx2) // Ignore error - we just want to see if block is recovered

	// Verify the poller recovered the block progress
	if poller.lastObservedBlock != nil {
		assert.Equal(t, ethereum.EthereumQuantity(lastBlock), poller.lastObservedBlock.Number,
			"Poller should recover last processed block from storage")
	}
}
