package EVMChainPoller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/mocks"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// Test helper to create a test poller
func createTestPoller(ethClient ethereum.Client, store storage.AggregatorStore) *EVMChainPoller {
	return &EVMChainPoller{
		ethClient: ethClient,
		store:     store,
		config: &EVMChainPollerConfig{
			AvsAddress:        "0xtest",
			ChainId:           config.ChainId(1),
			MaxReorgDepth:     10,
			ReorgCheckEnabled: true,
		},
		logger: zap.NewNop(),
	}
}

// Test Scenario 1: No reorg - all blocks match
func TestFindOrphanedBlocks_NoReorg_AllBlocksMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	// Setup: Save block 99 in store
	err := store.SaveBlock(ctx, "0xtest", &storage.BlockRecord{
		Number:     99,
		Hash:       "0x99",
		ParentHash: "0x98",
		ChainId:    config.ChainId(1),
	})
	require.NoError(t, err)

	startBlock := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(100),
		Hash:       ethereum.EthereumHexString("0xstart"),
		ParentHash: ethereum.EthereumHexString("0x99"),
		ChainId:    config.ChainId(1),
	}

	// Mock chain returns matching block 99
	chainBlock99 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(99),
		Hash:       ethereum.EthereumHexString("0x99"),
		ParentHash: ethereum.EthereumHexString("0x98"),
		ChainId:    config.ChainId(1),
	}
	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(99)).Return(chainBlock99, nil)

	poller := createTestPoller(mockClient, store)
	poller.lastObservedBlock = &ethereum.EthereumBlock{
		Number: ethereum.EthereumQuantity(98),
		Hash:   "0x98",
	}

	orphaned, err := poller.findOrphanedBlocks(ctx, startBlock, 10)

	assert.NoError(t, err)
	assert.Empty(t, orphaned, "Should find no orphaned blocks when all blocks match")
	assert.Equal(t, uint64(99), poller.lastObservedBlock.Number.Value(), "Should update lastObservedBlock to matching block")

	// Verify SetLastProcessedBlock was called
	lastProcessed, err := store.GetLastProcessedBlock(ctx, "0xtest", config.ChainId(1))
	require.NoError(t, err)
	assert.Equal(t, uint64(99), lastProcessed)
}

// Test Scenario 2: Simple reorg - finds orphaned blocks and common ancestor
func TestFindOrphanedBlocks_SimpleReorg_FindsOrphanedAndAncestor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	// Setup: Save blocks 97-99 in store (old chain)
	for i := uint64(97); i <= 99; i++ {
		err := store.SaveBlock(ctx, "0xtest", &storage.BlockRecord{
			Number:     i,
			Hash:       fmt.Sprintf("0x%d_old", i),
			ParentHash: fmt.Sprintf("0x%d_old", i-1),
			ChainId:    config.ChainId(1),
		})
		require.NoError(t, err)
	}

	// Also save block 97 with the correct hash that will match
	err := store.SaveBlock(ctx, "0xtest", &storage.BlockRecord{
		Number:     97,
		Hash:       "0x97",
		ParentHash: "0x96",
		ChainId:    config.ChainId(1),
	})
	require.NoError(t, err)

	startBlock := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(100),
		Hash:       ethereum.EthereumHexString("0xstart"),
		ParentHash: ethereum.EthereumHexString("0x99_new"),
		ChainId:    config.ChainId(1),
	}

	// Mock chain blocks
	chainBlock99 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(99),
		Hash:       ethereum.EthereumHexString("0x99_new"),
		ParentHash: ethereum.EthereumHexString("0x98_new"),
		ChainId:    config.ChainId(1),
	}
	chainBlock98 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(98),
		Hash:       ethereum.EthereumHexString("0x98_new"),
		ParentHash: ethereum.EthereumHexString("0x97"),
		ChainId:    config.ChainId(1),
	}
	chainBlock97 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(97),
		Hash:       ethereum.EthereumHexString("0x97"),
		ParentHash: ethereum.EthereumHexString("0x96"),
		ChainId:    config.ChainId(1),
	}

	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(99)).Return(chainBlock99, nil)
	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(98)).Return(chainBlock98, nil)
	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(97)).Return(chainBlock97, nil)

	poller := createTestPoller(mockClient, store)

	orphaned, err := poller.findOrphanedBlocks(ctx, startBlock, 10)

	assert.NoError(t, err)
	assert.Len(t, orphaned, 2, "Should find 2 orphaned blocks")
	assert.Equal(t, uint64(99), orphaned[0].Number.Value())
	assert.Equal(t, uint64(98), orphaned[1].Number.Value())
	assert.Equal(t, uint64(97), poller.lastObservedBlock.Number.Value(), "Should update lastObservedBlock to common ancestor")

	// Verify SetLastProcessedBlock was called
	lastProcessed, err := store.GetLastProcessedBlock(ctx, "0xtest", config.ChainId(1))
	require.NoError(t, err)
	assert.Equal(t, uint64(97), lastProcessed)
}

// Test Scenario 3: Deep reorg that hits maxDepth limit
func TestFindOrphanedBlocks_DeepReorg_HitsMaxDepth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	// Setup: Save blocks 97-99 in store (all with old hashes)
	for i := uint64(97); i <= 99; i++ {
		err := store.SaveBlock(ctx, "0xtest", &storage.BlockRecord{
			Number:     i,
			Hash:       fmt.Sprintf("0x%d_old", i),
			ParentHash: fmt.Sprintf("0x%d_old", i-1),
			ChainId:    config.ChainId(1),
		})
		require.NoError(t, err)
	}

	startBlock := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(100),
		Hash:       ethereum.EthereumHexString("0xstart"),
		ParentHash: ethereum.EthereumHexString("0x99_new"),
		ChainId:    config.ChainId(1),
	}

	maxDepth := 3

	// Mock chain blocks - all different from stored
	for i := uint64(99); i >= 97; i-- {
		chainBlock := &ethereum.EthereumBlock{
			Number:     ethereum.EthereumQuantity(i),
			Hash:       ethereum.EthereumHexString(fmt.Sprintf("0x%d_new", i)),
			ParentHash: ethereum.EthereumHexString(fmt.Sprintf("0x%d_new", i-1)),
			ChainId:    config.ChainId(1),
		}
		mockClient.EXPECT().GetBlockByNumber(ctx, i).Return(chainBlock, nil)
	}

	poller := createTestPoller(mockClient, store)

	orphaned, err := poller.findOrphanedBlocks(ctx, startBlock, maxDepth)

	assert.NoError(t, err)
	assert.Len(t, orphaned, 3, "Should find orphaned blocks up to maxDepth")
	assert.Equal(t, uint64(99), orphaned[0].Number.Value())
	assert.Equal(t, uint64(98), orphaned[1].Number.Value())
	assert.Equal(t, uint64(97), orphaned[2].Number.Value())
}

// Test Scenario 4: Block not found in storage (returns early)
func TestFindOrphanedBlocks_BlockNotInStorage_ReturnsEarly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	// Don't save block 99 in store - it will be missing

	startBlock := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(100),
		Hash:       ethereum.EthereumHexString("0xstart"),
		ParentHash: ethereum.EthereumHexString("0x99"),
		Timestamp:  ethereum.EthereumQuantity(1234567890),
		ChainId:    config.ChainId(1),
	}

	chainBlock99 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(99),
		Hash:       ethereum.EthereumHexString("0x99"),
		ParentHash: ethereum.EthereumHexString("0x98"),
		ChainId:    config.ChainId(1),
	}

	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(99)).Return(chainBlock99, nil)

	poller := createTestPoller(mockClient, store)

	orphaned, err := poller.findOrphanedBlocks(ctx, startBlock, 10)

	assert.NoError(t, err)
	assert.Empty(t, orphaned, "Should return empty orphaned list when block not found")
	assert.Equal(t, uint64(99), poller.lastObservedBlock.Number.Value(), "Should update lastObservedBlock to chainBlock")

	// Verify chainBlock99 info was saved when block not found in storage
	savedBlock, err := store.GetBlock(ctx, "0xtest", config.ChainId(1), 99)
	require.NoError(t, err)
	assert.Equal(t, "0x99", savedBlock.Hash)
	assert.Equal(t, "0x98", savedBlock.ParentHash)
}

// Test Scenario 5: Chain fetch fails
func TestFindOrphanedBlocks_ChainFetchFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	startBlock := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(100),
		Hash:       ethereum.EthereumHexString("0xstart"),
		ParentHash: ethereum.EthereumHexString("0x99"),
		ChainId:    config.ChainId(1),
	}

	expectedErr := errors.New("network error")
	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(99)).Return(nil, expectedErr)

	poller := createTestPoller(mockClient, store)

	orphaned, err := poller.findOrphanedBlocks(ctx, startBlock, 10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch block 99 from chain")
	assert.Nil(t, orphaned)
}

// Test state changes
func TestFindOrphanedBlocks_StateChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	// Setup: Save block 99 in store
	err := store.SaveBlock(ctx, "0xtest", &storage.BlockRecord{
		Number:     99,
		Hash:       "0x99",
		ParentHash: "0x98",
		ChainId:    config.ChainId(1),
	})
	require.NoError(t, err)

	startBlock := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(100),
		Hash:       ethereum.EthereumHexString("0xstart"),
		ParentHash: ethereum.EthereumHexString("0x99"),
		ChainId:    config.ChainId(1),
	}

	chainBlock99 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(99),
		Hash:       ethereum.EthereumHexString("0x99"),
		ParentHash: ethereum.EthereumHexString("0x98"),
		ChainId:    config.ChainId(1),
	}

	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(99)).Return(chainBlock99, nil)

	poller := createTestPoller(mockClient, store)
	// Initial state
	initialLastObserved := &ethereum.EthereumBlock{
		Number: ethereum.EthereumQuantity(95),
		Hash:   ethereum.EthereumHexString("0x95"),
	}
	poller.lastObservedBlock = initialLastObserved

	orphaned, err := poller.findOrphanedBlocks(ctx, startBlock, 10)

	require.NoError(t, err)
	assert.Empty(t, orphaned)

	// Verify state change
	assert.NotEqual(t, initialLastObserved, poller.lastObservedBlock, "lastObservedBlock should be updated")
	assert.Equal(t, uint64(99), poller.lastObservedBlock.Number.Value(), "lastObservedBlock should be set to matching block")
	assert.Equal(t, "0x99", poller.lastObservedBlock.Hash.Value(), "lastObservedBlock hash should match")
}

// Test reconcileReorg successfully deletes orphaned blocks
func TestReconcileReorg_Success_DeletesOrphanedBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	// Setup: Save blocks 98-99 in store (will be orphaned)
	for i := uint64(98); i <= 99; i++ {
		err := store.SaveBlock(ctx, "0xtest", &storage.BlockRecord{
			Number:     i,
			Hash:       fmt.Sprintf("0x%d_old", i),
			ParentHash: fmt.Sprintf("0x%d_old", i-1),
			ChainId:    config.ChainId(1),
		})
		require.NoError(t, err)
	}

	// Also save block 97 that will match (common ancestor)
	err := store.SaveBlock(ctx, "0xtest", &storage.BlockRecord{
		Number:     97,
		Hash:       "0x97",
		ParentHash: "0x96",
		ChainId:    config.ChainId(1),
	})
	require.NoError(t, err)

	startBlock := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(100),
		Hash:       ethereum.EthereumHexString("0x100_new"),
		ParentHash: ethereum.EthereumHexString("0x99_new"),
		ChainId:    config.ChainId(1),
	}

	// Mock chain blocks - different from stored until block 97
	chainBlock99 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(99),
		Hash:       ethereum.EthereumHexString("0x99_new"),
		ParentHash: ethereum.EthereumHexString("0x98_new"),
		ChainId:    config.ChainId(1),
	}
	chainBlock98 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(98),
		Hash:       ethereum.EthereumHexString("0x98_new"),
		ParentHash: ethereum.EthereumHexString("0x97"),
		ChainId:    config.ChainId(1),
	}
	chainBlock97 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(97),
		Hash:       ethereum.EthereumHexString("0x97"),
		ParentHash: ethereum.EthereumHexString("0x96"),
		ChainId:    config.ChainId(1),
	}

	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(99)).Return(chainBlock99, nil)
	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(98)).Return(chainBlock98, nil)
	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(97)).Return(chainBlock97, nil)

	poller := &EVMChainPoller{
		ethClient: mockClient,
		store:     store,
		config: &EVMChainPollerConfig{
			AvsAddress:    "0xtest",
			ChainId:       config.ChainId(1),
			MaxReorgDepth: 10,
		},
		logger: zap.NewNop(),
	}

	// Execute reconcileReorg
	err = poller.reconcileReorg(ctx, startBlock)

	// Verify no error
	assert.NoError(t, err)

	// Verify orphaned blocks were deleted from storage
	_, err = store.GetBlock(ctx, "0xtest", config.ChainId(1), 99)
	assert.Error(t, err, "Block 99 should be deleted")
	assert.True(t, errors.Is(err, storage.ErrNotFound))

	_, err = store.GetBlock(ctx, "0xtest", config.ChainId(1), 98)
	assert.Error(t, err, "Block 98 should be deleted")
	assert.True(t, errors.Is(err, storage.ErrNotFound))

	// Verify common ancestor block 97 still exists
	block97, err := store.GetBlock(ctx, "0xtest", config.ChainId(1), 97)
	assert.NoError(t, err, "Block 97 should still exist")
	assert.Equal(t, "0x97", block97.Hash)
}

// Test reconcileReorg returns error when no orphaned blocks are found
func TestReconcileReorg_NoOrphanedBlocks_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	// Setup: Save block 99 that matches chain (no reorg)
	err := store.SaveBlock(ctx, "0xtest", &storage.BlockRecord{
		Number:     99,
		Hash:       "0x99",
		ParentHash: "0x98",
		ChainId:    config.ChainId(1),
	})
	require.NoError(t, err)

	startBlock := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(100),
		Hash:       ethereum.EthereumHexString("0x100"),
		ParentHash: ethereum.EthereumHexString("0x99"),
		ChainId:    config.ChainId(1),
	}

	// Mock chain returns matching block (no reorg)
	chainBlock99 := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(99),
		Hash:       ethereum.EthereumHexString("0x99"),
		ParentHash: ethereum.EthereumHexString("0x98"),
		ChainId:    config.ChainId(1),
	}
	mockClient.EXPECT().GetBlockByNumber(ctx, uint64(99)).Return(chainBlock99, nil)

	poller := &EVMChainPoller{
		ethClient: mockClient,
		store:     store,
		config: &EVMChainPollerConfig{
			AvsAddress:    "0xtest",
			ChainId:       config.ChainId(1),
			MaxReorgDepth: 10,
		},
		logger: zap.NewNop(),
	}
	poller.lastObservedBlock = &ethereum.EthereumBlock{
		Number: ethereum.EthereumQuantity(98),
		Hash:   ethereum.EthereumHexString("0x98"),
	}

	// Execute reconcileReorg
	err = poller.reconcileReorg(ctx, startBlock)

	// Should return error when no orphaned blocks found
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no orphaned blocks found")
}

// Test that expired tasks are marked as failed during recovery
func TestRecoverInProgressTasks_ExpiredTasks_MarkedAsFailed(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	taskQueue := make(chan *types.Task, 10)

	// Create expired and non-expired deadlines
	expiredDeadline := time.Now().Add(-1 * time.Hour)
	validDeadline := time.Now().Add(1 * time.Hour)

	// Create tasks with different deadlines
	expiredTask1 := &types.Task{
		TaskId:              "expired-task-1",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &expiredDeadline,
		ReferenceTimestamp:  100,
	}
	expiredTask2 := &types.Task{
		TaskId:              "expired-task-2",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &expiredDeadline,
		ReferenceTimestamp:  100,
	}
	validTask := &types.Task{
		TaskId:              "valid-task",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		ReferenceTimestamp:  100,
	}

	// Save all tasks as pending
	require.NoError(t, store.SavePendingTask(ctx, expiredTask1))
	require.NoError(t, store.SavePendingTask(ctx, expiredTask2))
	require.NoError(t, store.SavePendingTask(ctx, validTask))

	poller := &EVMChainPoller{
		taskQueue: taskQueue,
		store:     store,
		config: &EVMChainPollerConfig{
			AvsAddress: "0xtest",
			ChainId:    config.ChainId(1),
		},
		logger: zap.NewNop(),
	}

	// Execute recoverInProgressTasks
	err := poller.recoverInProgressTasks(ctx)
	assert.NoError(t, err)

	// Verify only valid task is in the queue (expired tasks should not be queued)
	assert.Equal(t, 1, len(taskQueue), "Only valid task should be in queue")
	recoveredTask := <-taskQueue
	assert.Equal(t, "valid-task", recoveredTask.TaskId)

	// Verify expired tasks are no longer in pending list
	pendingTasks, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pendingTasks), "No tasks should remain pending after recovery")

	// Expired tasks should not be in pending list
	for _, task := range pendingTasks {
		assert.NotContains(t, []string{"expired-task-1", "expired-task-2"}, task.TaskId,
			"Expired tasks should not be in pending list")
	}
}

// Test that tasks are properly queued and their status is updated correctly
func TestRecoverInProgressTasks_TasksQueued_StatusUpdated(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	taskQueue := make(chan *types.Task, 10)

	validDeadline := time.Now().Add(1 * time.Hour)

	// Create multiple valid tasks
	task1 := &types.Task{
		TaskId:              "task-1",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		Payload:             []byte("payload-1"),
		ReferenceTimestamp:  100,
	}
	task2 := &types.Task{
		TaskId:              "task-2",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		Payload:             []byte("payload-2"),
		ReferenceTimestamp:  100,
	}
	task3 := &types.Task{
		TaskId:              "task-3",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		Payload:             []byte("payload-3"),
		ReferenceTimestamp:  100,
	}

	// Save all tasks as pending
	require.NoError(t, store.SavePendingTask(ctx, task1))
	require.NoError(t, store.SavePendingTask(ctx, task2))
	require.NoError(t, store.SavePendingTask(ctx, task3))

	// Verify tasks are initially in pending list
	pendingBefore, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(pendingBefore), "Should have 3 pending tasks before recovery")

	poller := &EVMChainPoller{
		taskQueue: taskQueue,
		store:     store,
		config: &EVMChainPollerConfig{
			AvsAddress: "0xtest",
			ChainId:    config.ChainId(1),
		},
		logger: zap.NewNop(),
	}

	// Execute recoverInProgressTasks
	err = poller.recoverInProgressTasks(ctx)
	assert.NoError(t, err)

	// Verify all tasks are in the queue
	assert.Equal(t, 3, len(taskQueue), "All 3 tasks should be in queue")

	// Pull tasks from queue and verify they match what's in storage
	tasksFromQueue := make(map[string]*types.Task)
	for i := 0; i < 3; i++ {
		task := <-taskQueue
		tasksFromQueue[task.TaskId] = task

		// Verify task data matches what was stored
		storedTask, err := store.GetTask(ctx, task.TaskId)
		assert.NoError(t, err)
		assert.Equal(t, task.TaskId, storedTask.TaskId)
		assert.Equal(t, task.AVSAddress, storedTask.AVSAddress)
		assert.Equal(t, task.Payload, storedTask.Payload)
	}

	// Verify we got all expected tasks
	assert.NotNil(t, tasksFromQueue["task-1"])
	assert.NotNil(t, tasksFromQueue["task-2"])
	assert.NotNil(t, tasksFromQueue["task-3"])

	// Verify payloads match
	assert.Equal(t, []byte("payload-1"), tasksFromQueue["task-1"].Payload)
	assert.Equal(t, []byte("payload-2"), tasksFromQueue["task-2"].Payload)
	assert.Equal(t, []byte("payload-3"), tasksFromQueue["task-3"].Payload)

	// After recovery, no tasks should remain in pending list (all moved to processing)
	pendingAfter, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pendingAfter), "No tasks should remain pending after recovery")
}

// Test recovery with no pending tasks
func TestRecoverInProgressTasks_NoPendingTasks_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	taskQueue := make(chan *types.Task, 10)

	poller := &EVMChainPoller{
		taskQueue: taskQueue,
		store:     store,
		config: &EVMChainPollerConfig{
			AvsAddress: "0xtest",
			ChainId:    config.ChainId(1),
		},
		logger: zap.NewNop(),
	}

	// Execute recoverInProgressTasks with no pending tasks
	err := poller.recoverInProgressTasks(ctx)
	assert.NoError(t, err)

	// Verify queue is empty
	assert.Equal(t, 0, len(taskQueue), "Queue should be empty")
}

// Test recovery when task queue is full
func TestRecoverInProgressTasks_QueueFull_TasksRemainPending(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	taskQueue := make(chan *types.Task, 2) // Small queue that can only hold 2 tasks

	validDeadline := time.Now().Add(1 * time.Hour)

	// Create 4 tasks but queue can only hold 2
	for i := 1; i <= 4; i++ {
		task := &types.Task{
			TaskId:              fmt.Sprintf("task-%d", i),
			AVSAddress:          "0xtest",
			DeadlineUnixSeconds: &validDeadline,
			ReferenceTimestamp:  100,
		}
		require.NoError(t, store.SavePendingTask(ctx, task))
	}

	poller := &EVMChainPoller{
		taskQueue: taskQueue,
		store:     store,
		config: &EVMChainPollerConfig{
			AvsAddress: "0xtest",
			ChainId:    config.ChainId(1),
		},
		logger: zap.NewNop(),
	}

	// Execute recoverInProgressTasks
	err := poller.recoverInProgressTasks(ctx)
	assert.NoError(t, err)

	// Only 2 tasks should be in queue
	assert.Equal(t, 2, len(taskQueue), "Queue should contain only 2 tasks")

	// Verify which tasks were queued
	queuedTaskIds := make(map[string]bool)
	for len(taskQueue) > 0 {
		task := <-taskQueue
		queuedTaskIds[task.TaskId] = true
	}

	// Exactly 2 tasks should have been queued
	assert.Equal(t, 2, len(queuedTaskIds), "Exactly 2 tasks should have been queued")

	// The remaining tasks should still be in pending state
	pendingTasks, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	// Since the queue was full, 2 tasks should remain pending
	assert.Equal(t, 2, len(pendingTasks), "2 tasks should remain pending when queue is full")
}

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
		ReferenceTimestamp:  100,
	}
	task2 := &types.Task{
		TaskId:              "task-2",
		AVSAddress:          avsAddress,
		DeadlineUnixSeconds: &validDeadline,
		ReferenceTimestamp:  100,
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
		ReferenceTimestamp:  100,
	}
	expiredTask := &types.Task{
		TaskId:              "expired-task",
		AVSAddress:          avsAddress,
		DeadlineUnixSeconds: &expiredDeadline,
		ReferenceTimestamp:  100,
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
			ReferenceTimestamp:  100,
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
