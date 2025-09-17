package EVMChainPoller

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/mocks"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contextManager/taskBlockContextManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type testContextKey string

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

	orphaned, err := poller.findOrphanedBlocks(ctx, startBlock, 10)

	assert.NoError(t, err)
	assert.Empty(t, orphaned, "Should find no orphaned blocks when all blocks match")

	// Verify block was saved via SaveBlock
	lastProcessed, err := store.GetLastProcessedBlock(ctx, "0xtest", config.ChainId(1))
	require.NoError(t, err)
	assert.Equal(t, uint64(99), lastProcessed.Number)
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
	assert.Equal(t, uint64(99), orphaned[0].Number)
	assert.Equal(t, uint64(98), orphaned[1].Number)

	// Verify block was saved via SaveBlock
	lastProcessed, err := store.GetLastProcessedBlock(ctx, "0xtest", config.ChainId(1))
	require.NoError(t, err)
	assert.Equal(t, uint64(97), lastProcessed.Number)
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
	assert.Equal(t, uint64(99), orphaned[0].Number)
	assert.Equal(t, uint64(98), orphaned[1].Number)
	assert.Equal(t, uint64(97), orphaned[2].Number)
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

	orphaned, err := poller.findOrphanedBlocks(ctx, startBlock, 10)

	require.NoError(t, err)
	assert.Empty(t, orphaned)

	// Verify block was saved to storage
	savedBlock, err := store.GetBlock(ctx, "0xtest", config.ChainId(1), 99)
	require.NoError(t, err)
	assert.Equal(t, "0x99", savedBlock.Hash)
	assert.Equal(t, "0x98", savedBlock.ParentHash)
}

// Test reconcileReorg successfully deletes orphaned blocks
func TestReconcileReorg_Success_DeletesOrphanedBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	mockBlockContextManager := mocks.NewMockIBlockContextManager(ctrl)
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

	// Expect CancelBlock to be called for each orphaned block
	mockBlockContextManager.EXPECT().CancelBlock(uint64(99))
	mockBlockContextManager.EXPECT().CancelBlock(uint64(98))

	poller := &EVMChainPoller{
		ethClient:           mockClient,
		store:               store,
		blockContextManager: mockBlockContextManager,
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
		SourceBlockNumber:   1000,
	}
	expiredTask2 := &types.Task{
		TaskId:              "expired-task-2",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &expiredDeadline,
		ReferenceTimestamp:  100,
		SourceBlockNumber:   1001,
	}
	validTask := &types.Task{
		TaskId:              "valid-task",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		ReferenceTimestamp:  100,
		SourceBlockNumber:   1002,
	}

	// Save all tasks as pending
	require.NoError(t, store.SavePendingTask(ctx, expiredTask1))
	require.NoError(t, store.SavePendingTask(ctx, expiredTask2))
	require.NoError(t, store.SavePendingTask(ctx, validTask))

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBlockContextManager := mocks.NewMockIBlockContextManager(mockCtrl)

	// Setup mock to return expired context for expired tasks and valid context for valid task
	expiredCtx, expiredCancel := context.WithCancel(ctx)
	expiredCancel() // Cancel immediately to simulate expired context
	validCtx, cancel := context.WithDeadline(ctx, validDeadline)
	defer cancel()

	mockBlockContextManager.EXPECT().GetContext(uint64(1000), expiredTask1).Return(expiredCtx)
	mockBlockContextManager.EXPECT().GetContext(uint64(1001), expiredTask2).Return(expiredCtx)
	mockBlockContextManager.EXPECT().GetContext(uint64(1002), validTask).Return(validCtx)

	poller := &EVMChainPoller{
		taskQueue:           taskQueue,
		store:               store,
		blockContextManager: mockBlockContextManager,
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

	// Verify expired tasks are no longer pending and valid task remains pending
	pendingTasks, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pendingTasks), "Valid task should remain pending until processed by AvsExecutionManager")
	assert.Equal(t, "valid-task", pendingTasks[0].TaskId, "Only valid task should be pending")

	// Verify expired tasks are not in the pending list
	for _, task := range pendingTasks {
		assert.NotContains(t, []string{"expired-task-1", "expired-task-2"}, task.TaskId,
			"Expired tasks should not be in pending list after being marked as failed")
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
		SourceBlockNumber:   2000,
	}
	task2 := &types.Task{
		TaskId:              "task-2",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		Payload:             []byte("payload-2"),
		ReferenceTimestamp:  100,
		SourceBlockNumber:   2001,
	}
	task3 := &types.Task{
		TaskId:              "task-3",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		Payload:             []byte("payload-3"),
		ReferenceTimestamp:  100,
		SourceBlockNumber:   2002,
	}

	// Save all tasks as pending
	require.NoError(t, store.SavePendingTask(ctx, task1))
	require.NoError(t, store.SavePendingTask(ctx, task2))
	require.NoError(t, store.SavePendingTask(ctx, task3))

	// Verify tasks are initially in pending list
	pendingBefore, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(pendingBefore), "Should have 3 pending tasks before recovery")

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBlockContextManager := mocks.NewMockIBlockContextManager(mockCtrl)

	// Setup mock to return valid contexts for all tasks
	validCtx, cancel := context.WithDeadline(ctx, validDeadline)
	defer cancel()

	mockBlockContextManager.EXPECT().GetContext(uint64(2000), task1).Return(validCtx)
	mockBlockContextManager.EXPECT().GetContext(uint64(2001), task2).Return(validCtx)
	mockBlockContextManager.EXPECT().GetContext(uint64(2002), task3).Return(validCtx)

	poller := &EVMChainPoller{
		taskQueue:           taskQueue,
		store:               store,
		blockContextManager: mockBlockContextManager,
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

	// After recovery, tasks remain in pending status until processed by AvsExecutionManager
	pendingAfter, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(pendingAfter), "All tasks should remain pending until processed by AvsExecutionManager")
}

// Test recovery with no pending tasks
func TestRecoverInProgressTasks_NoPendingTasks_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	taskQueue := make(chan *types.Task, 10)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBlockContextManager := mocks.NewMockIBlockContextManager(mockCtrl)

	poller := &EVMChainPoller{
		taskQueue:           taskQueue,
		store:               store,
		blockContextManager: mockBlockContextManager,
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
	tasks := make([]*types.Task, 4)
	for i := 0; i < 4; i++ {
		tasks[i] = &types.Task{
			TaskId:              fmt.Sprintf("task-%d", i+1),
			AVSAddress:          "0xtest",
			DeadlineUnixSeconds: &validDeadline,
			ReferenceTimestamp:  100,
			SourceBlockNumber:   uint64(3000 + i),
		}
		require.NoError(t, store.SavePendingTask(ctx, tasks[i]))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBlockContextManager := mocks.NewMockIBlockContextManager(mockCtrl)

	// Setup mock to return valid contexts for all tasks
	validCtx, cancel := context.WithDeadline(ctx, validDeadline)
	defer cancel()

	// We expect GetContext to be called for at least the first 2 tasks (queue capacity)
	// Since the order of task recovery may vary, use AnyTimes()
	mockBlockContextManager.EXPECT().GetContext(gomock.Any(), gomock.Any()).Return(validCtx).AnyTimes()

	poller := &EVMChainPoller{
		taskQueue:           taskQueue,
		store:               store,
		blockContextManager: mockBlockContextManager,
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

	// All tasks should still be in pending state (they're queued but not yet processed)
	pendingTasks, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	// All 4 tasks remain pending until AvsExecutionManager processes them
	assert.Equal(t, 4, len(pendingTasks), "All tasks remain pending until processed by AvsExecutionManager")
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
		SourceBlockNumber:   4000,
	}
	task2 := &types.Task{
		TaskId:              "task-2",
		AVSAddress:          avsAddress,
		DeadlineUnixSeconds: &validDeadline,
		ReferenceTimestamp:  100,
		SourceBlockNumber:   4001,
	}

	// Pre-populate storage with pending tasks
	require.NoError(t, store.SavePendingTask(ctx, task1))
	require.NoError(t, store.SavePendingTask(ctx, task2))

	// Create minimal poller to test recovery
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	taskQueue := make(chan *types.Task, 10)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBlockContextManager := mocks.NewMockIBlockContextManager(mockCtrl)

	// Setup mock to return valid contexts
	validCtx, cancel := context.WithDeadline(ctx, validDeadline)
	defer cancel()
	mockBlockContextManager.EXPECT().GetContext(uint64(4000), task1).Return(validCtx)
	mockBlockContextManager.EXPECT().GetContext(uint64(4001), task2).Return(validCtx)

	poller := &EVMChainPoller{
		config: &EVMChainPollerConfig{
			AvsAddress: avsAddress,
		},
		taskQueue:           taskQueue,
		store:               store,
		blockContextManager: mockBlockContextManager,
		logger:              l,
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

	// Verify tasks are still accessible in storage
	task1FromStore, err := store.GetTask(ctx, "task-1")
	assert.NoError(t, err, "Should be able to get task-1 from store")
	assert.Equal(t, "task-1", task1FromStore.TaskId)

	task2FromStore, err := store.GetTask(ctx, "task-2")
	assert.NoError(t, err, "Should be able to get task-2 from store")
	assert.Equal(t, "task-2", task2FromStore.TaskId)
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
		SourceBlockNumber:   5000,
	}
	expiredTask := &types.Task{
		TaskId:              "expired-task",
		AVSAddress:          avsAddress,
		DeadlineUnixSeconds: &expiredDeadline,
		ReferenceTimestamp:  100,
		SourceBlockNumber:   4999,
	}

	// Save both tasks to storage
	require.NoError(t, store.SavePendingTask(ctx, validTask))
	require.NoError(t, store.SavePendingTask(ctx, expiredTask))

	// Create minimal poller
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	taskQueue := make(chan *types.Task, 10)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBlockContextManager := mocks.NewMockIBlockContextManager(mockCtrl)

	// Setup mock to return expired context for expired task and valid context for valid task
	expiredCtx, expiredCancel := context.WithCancel(ctx)
	expiredCancel() // Cancel immediately to simulate expired context
	validCtx, cancel := context.WithDeadline(ctx, validDeadline)
	defer cancel()

	mockBlockContextManager.EXPECT().GetContext(uint64(5000), validTask).Return(validCtx)
	mockBlockContextManager.EXPECT().GetContext(uint64(4999), expiredTask).Return(expiredCtx)

	poller := &EVMChainPoller{
		config: &EVMChainPollerConfig{
			AvsAddress: avsAddress,
		},
		taskQueue:           taskQueue,
		store:               store,
		blockContextManager: mockBlockContextManager,
		logger:              l,
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
	tasks := make([]*types.Task, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = &types.Task{
			TaskId:              fmt.Sprintf("task-%d", i),
			AVSAddress:          avsAddress,
			DeadlineUnixSeconds: &validDeadline,
			ReferenceTimestamp:  100,
			SourceBlockNumber:   uint64(6000 + i),
		}
		require.NoError(t, store.SavePendingTask(ctx, tasks[i]))
	}

	// Create poller with small channel
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	smallTaskQueue := make(chan *types.Task, 2)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBlockContextManager := mocks.NewMockIBlockContextManager(mockCtrl)

	// Setup mock to return valid contexts
	validCtx, cancel := context.WithDeadline(ctx, validDeadline)
	defer cancel()
	mockBlockContextManager.EXPECT().GetContext(gomock.Any(), gomock.Any()).Return(validCtx).AnyTimes()

	poller := &EVMChainPoller{
		config: &EVMChainPollerConfig{
			AvsAddress: avsAddress,
		},
		taskQueue:           smallTaskQueue,
		store:               store,
		blockContextManager: mockBlockContextManager,
		logger:              l,
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

func TestEVMChainPoller_TaskContextAssignment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	mockLogParser := mocks.NewMockLogParser(ctrl)
	mockContractStore := mocks.NewMockIContractStore(ctrl)
	mockBlockContextManager := mocks.NewMockIBlockContextManager(ctrl)
	store := memory.NewInMemoryAggregatorStore()
	taskQueue := make(chan *types.Task, 10)

	// Use a valid hex address for AVS
	avsAddress := "0x0000000000000000000000000000000000000001"

	poller := NewEVMChainPoller(
		mockClient,
		taskQueue,
		mockLogParser,
		&EVMChainPollerConfig{
			AvsAddress:           avsAddress,
			ChainId:              config.ChainId(1),
			InterestingContracts: []string{"0xmailbox"},
		},
		mockContractStore,
		store,
		mockBlockContextManager,
		zap.NewNop(),
	)

	// Setup test data
	blockNumber := uint64(100)
	const testKey testContextKey = "test-key"
	testContext := context.WithValue(context.Background(), testKey, "test-value")

	block := &ethereum.EthereumBlock{
		Number:     ethereum.EthereumQuantity(blockNumber),
		Hash:       ethereum.EthereumHexString("0x100"),
		ParentHash: ethereum.EthereumHexString("0x99"),
		ChainId:    config.ChainId(1),
	}

	// Create a TaskCreated event log
	// TaskCreated event has 3 indexed parameters
	// The AVS address needs to be a common.Address type
	decodedLog := &log.DecodedLog{
		EventName: "TaskCreated",
		Address:   "0xmailbox",
		Arguments: []log.Argument{
			{Name: "creator", Value: "0xmailbox", Indexed: true, Type: "address"},
			{Name: "taskHash", Value: "0xtask1", Indexed: true, Type: "bytes32"},
			{Name: "avs", Value: common.HexToAddress(avsAddress), Indexed: true, Type: "address"},
		},
		OutputData: map[string]interface{}{
			"ExecutorOperatorSetId":           uint32(1),
			"OperatorTableReferenceTimestamp": uint32(1234567890),
			"TaskDeadline":                    big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
			"Payload":                         []byte("test-payload"),
		},
	}

	lwb := &chainPoller.LogWithBlock{
		Block: block,
		RawLog: &ethereum.EthereumEventLog{
			Address:         ethereum.EthereumHexString("0xmailbox"),
			TransactionHash: ethereum.EthereumHexString("0xtx1"),
			LogIndex:        ethereum.EthereumQuantity(0),
		},
		Log: decodedLog,
	}

	mockBlockContextManager.EXPECT().
		GetContext(blockNumber, gomock.Any()).
		DoAndReturn(func(bn uint64, task *types.Task) context.Context {
			assert.Equal(t, blockNumber, bn)
			assert.Equal(t, "0xtask1", task.TaskId)
			assert.Equal(t, avsAddress, task.AVSAddress)
			return testContext
		})

	// Process the task
	err := poller.processTask(ctx, lwb)
	require.NoError(t, err)

	// Verify task was queued with context
	select {
	case task := <-taskQueue:
		assert.NotNil(t, task.Context)
		assert.Equal(t, "test-value", task.Context.Value(testKey))
		assert.Equal(t, "0xtask1", task.TaskId)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Task not queued")
	}
}

func TestEVMChainPoller_ContextCancellationOnReorg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mocks.NewMockClient(ctrl)
	mockBlockContextManager := mocks.NewMockIBlockContextManager(ctrl)
	store := memory.NewInMemoryAggregatorStore()

	// Setup: Save blocks 98-99 in store (will be orphaned)
	for i := uint64(98); i <= 99; i++ {
		err := store.SaveBlock(ctx, "0xavs", &storage.BlockRecord{
			Number:     i,
			Hash:       fmt.Sprintf("0x%d_old", i),
			ParentHash: fmt.Sprintf("0x%d_old", i-1),
			ChainId:    config.ChainId(1),
		})
		require.NoError(t, err)
	}

	// Save block 97 that will match (common ancestor)
	err := store.SaveBlock(ctx, "0xavs", &storage.BlockRecord{
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

	// Expect CancelBlock to be called for orphaned blocks
	mockBlockContextManager.EXPECT().CancelBlock(uint64(99))
	mockBlockContextManager.EXPECT().CancelBlock(uint64(98))

	poller := &EVMChainPoller{
		ethClient:           mockClient,
		store:               store,
		blockContextManager: mockBlockContextManager,
		config: &EVMChainPollerConfig{
			AvsAddress:    "0xavs",
			ChainId:       config.ChainId(1),
			MaxReorgDepth: 10,
		},
		logger: zap.NewNop(),
	}

	// Execute reconcileReorg
	err = poller.reconcileReorg(ctx, startBlock)
	assert.NoError(t, err)

	// Verify CancelBlock was called for each orphaned block
	ctrl.Finish() // This will verify all expectations were met
}

// Test TaskBlockContextManager implementation
func TestTaskBlockContextManager_GetContext(t *testing.T) {
	parentCtx := context.Background()
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := taskBlockContextManager.NewTaskBlockContextManager(parentCtx, store, logger)

	// Create a task with deadline
	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	// Get context for block 100
	ctx1 := mgr.GetContext(100, task)
	assert.NotNil(t, ctx1)

	// Get context again for same block - should return same context
	ctx2 := mgr.GetContext(100, task)
	assert.Equal(t, ctx1, ctx2)

	// Verify context has deadline
	deadline1, ok := ctx1.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, deadline, deadline1, 100*time.Millisecond)
}

// Test TaskBlockContextManager cancellation
func TestTaskBlockContextManager_CancelBlock(t *testing.T) {
	parentCtx := context.Background()
	l := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()
	mgr := taskBlockContextManager.NewTaskBlockContextManager(parentCtx, store, l)

	// Create contexts for multiple blocks
	deadline := time.Now().Add(1 * time.Hour)
	task := &types.Task{
		TaskId:              "task-1",
		DeadlineUnixSeconds: &deadline,
	}

	ctx1 := mgr.GetContext(100, task)
	ctx2 := mgr.GetContext(101, task)
	ctx3 := mgr.GetContext(102, task)

	// Cancel block 101
	mgr.CancelBlock(101)

	// Check context states
	select {
	case <-ctx1.Done():
		t.Fatal("Context 1 should not be cancelled")
	default:
		// Expected
	}

	select {
	case <-ctx2.Done():
		// Expected - should be cancelled
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context 2 should be cancelled")
	}

	select {
	case <-ctx3.Done():
		t.Fatal("Context 3 should not be cancelled")
	default:
		// Expected
	}
}

// Test that tasks loaded from storage are populated with their context
func TestRecoverInProgressTasks_PopulatesContextFromStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	store := memory.NewInMemoryAggregatorStore()
	taskQueue := make(chan *types.Task, 10)
	mockBlockContextManager := mocks.NewMockIBlockContextManager(ctrl)

	// Create tasks with different source block numbers
	validDeadline := time.Now().Add(1 * time.Hour)
	expiredDeadline := time.Now().Add(-1 * time.Hour)

	task1 := &types.Task{
		TaskId:              "task-1",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		SourceBlockNumber:   100,
		ReferenceTimestamp:  1234567890,
	}

	task2 := &types.Task{
		TaskId:              "task-2",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &validDeadline,
		SourceBlockNumber:   101,
		ReferenceTimestamp:  1234567891,
	}

	expiredTask := &types.Task{
		TaskId:              "expired-task",
		AVSAddress:          "0xtest",
		DeadlineUnixSeconds: &expiredDeadline,
		SourceBlockNumber:   99,
		ReferenceTimestamp:  1234567889,
	}

	// Save all tasks as pending
	require.NoError(t, store.SavePendingTask(ctx, task1))
	require.NoError(t, store.SavePendingTask(ctx, task2))
	require.NoError(t, store.SavePendingTask(ctx, expiredTask))

	// Create test contexts with different states
	validCtx1 := context.WithValue(context.Background(), testContextKey("block"), "100")
	validCtx2 := context.WithValue(context.Background(), testContextKey("block"), "101")
	expiredCtx, cancelExpired := context.WithCancel(context.Background())
	cancelExpired() // Make it expired

	// Set up expectations for GetContext calls
	mockBlockContextManager.EXPECT().
		GetContext(uint64(100), gomock.Any()).
		DoAndReturn(func(blockNum uint64, task *types.Task) context.Context {
			assert.Equal(t, uint64(100), blockNum)
			assert.Equal(t, "task-1", task.TaskId)
			return validCtx1
		})

	mockBlockContextManager.EXPECT().
		GetContext(uint64(101), gomock.Any()).
		DoAndReturn(func(blockNum uint64, task *types.Task) context.Context {
			assert.Equal(t, uint64(101), blockNum)
			assert.Equal(t, "task-2", task.TaskId)
			return validCtx2
		})

	mockBlockContextManager.EXPECT().
		GetContext(uint64(99), gomock.Any()).
		DoAndReturn(func(blockNum uint64, task *types.Task) context.Context {
			assert.Equal(t, uint64(99), blockNum)
			assert.Equal(t, "expired-task", task.TaskId)
			return expiredCtx
		})

	poller := &EVMChainPoller{
		taskQueue:           taskQueue,
		store:               store,
		blockContextManager: mockBlockContextManager,
		config: &EVMChainPollerConfig{
			AvsAddress: "0xtest",
			ChainId:    config.ChainId(1),
		},
		logger: zap.NewNop(),
	}

	// Execute recoverInProgressTasks
	err := poller.recoverInProgressTasks(ctx)
	assert.NoError(t, err)

	// Verify that valid tasks are queued with proper contexts
	queuedTasks := make(map[string]*types.Task)
	for i := 0; i < 2; i++ {
		select {
		case task := <-taskQueue:
			queuedTasks[task.TaskId] = task
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected 2 tasks to be queued")
		}
	}

	// Verify task 1 has the correct context
	assert.NotNil(t, queuedTasks["task-1"])
	assert.NotNil(t, queuedTasks["task-1"].Context)
	assert.Equal(t, "100", queuedTasks["task-1"].Context.Value(testContextKey("block")))

	// Verify task 2 has the correct context
	assert.NotNil(t, queuedTasks["task-2"])
	assert.NotNil(t, queuedTasks["task-2"].Context)
	assert.Equal(t, "101", queuedTasks["task-2"].Context.Value(testContextKey("block")))

	// Verify expired task was not queued
	select {
	case task := <-taskQueue:
		t.Fatalf("Should not receive expired task, got: %s", task.TaskId)
	case <-time.After(50 * time.Millisecond):
		// Expected - no more tasks
	}

	// Verify that valid tasks remain pending (they're queued but not yet picked up by AvsExecutionManager)
	pendingTasks, err := store.ListPendingTasksForAVS(ctx, "0xtest")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(pendingTasks), "Valid tasks should remain pending until AvsExecutionManager processes them")

	// Verify the pending tasks are the valid ones, not the expired one
	pendingTaskIds := make(map[string]bool)
	for _, task := range pendingTasks {
		pendingTaskIds[task.TaskId] = true
	}
	assert.True(t, pendingTaskIds["task-1"], "task-1 should be pending")
	assert.True(t, pendingTaskIds["task-2"], "task-2 should be pending")
	assert.False(t, pendingTaskIds["expired-task"], "expired-task should not be pending")
}
