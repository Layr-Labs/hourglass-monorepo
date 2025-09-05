package upgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	aggregatorBadger "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/badger"
	aggregatorMemory "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	executorStorage "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	executorBadger "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/badger"
	executorMemory "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/require"
)

// TestRollingUpgrade simulates upgrading services one at a time
func TestRollingUpgrade(t *testing.T) {
	tests := []struct {
		name         string
		storeFactory func(t *testing.T) (storage.AggregatorStore, executorStorage.ExecutorStore, func())
	}{
		{
			name: "Memory",
			storeFactory: func(t *testing.T) (storage.AggregatorStore, executorStorage.ExecutorStore, func()) {
				aggStore := aggregatorMemory.NewInMemoryAggregatorStore()
				execStore := executorMemory.NewInMemoryExecutorStore()
				return aggStore, execStore, func() {
					aggStore.Close()
					execStore.Close()
				}
			},
		},
		{
			name: "BadgerDB",
			storeFactory: func(t *testing.T) (storage.AggregatorStore, executorStorage.ExecutorStore, func()) {
				// Create persistent directories that won't be cleaned up until end of test
				aggDir := filepath.Join(t.TempDir(), "aggregator")
				execDir := filepath.Join(t.TempDir(), "executor")

				require.NoError(t, os.MkdirAll(aggDir, 0755))
				require.NoError(t, os.MkdirAll(execDir, 0755))

				// Store directories in test context for reuse
				t.Setenv("TEST_AGG_DIR", aggDir)
				t.Setenv("TEST_EXEC_DIR", execDir)

				aggStore, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
					Dir: aggDir,
				})
				require.NoError(t, err)

				execStore, err := executorBadger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
					Dir: execDir,
				})
				require.NoError(t, err)

				return aggStore, execStore, func() {
					aggStore.Close()
					execStore.Close()
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Phase 1: Create initial state
			aggStore1, execStore1, cleanup1 := tt.storeFactory(t)

			// Populate aggregator state
			avsAddress := "0xtest"
			chainId := config.ChainId(1)
			require.NoError(t, aggStore1.SetLastProcessedBlock(ctx, avsAddress, chainId, 1000))

			task := &types.Task{
				TaskId:                 "upgrade-task-1",
				AVSAddress:             "0x123",
				OperatorSetId:          1,
				SourceBlockNumber:      1000,
				L1ReferenceBlockNumber: 1000,
				ChainId:                config.ChainId(1),
			}
			require.NoError(t, aggStore1.SavePendingTask(ctx, task))

			// Populate executor state
			performer := &executorStorage.PerformerState{
				PerformerId: "performer-1",
				AvsAddress:  "0x123",
				ContainerId: "container-1",
				Status:      "running",
				CreatedAt:   time.Now(),
			}
			require.NoError(t, execStore1.SavePerformerState(ctx, "performer-1", performer))

			// Simulate shutdown
			cleanup1()

			// Phase 2: Restart with "new version" (same storage)
			var aggStore2 storage.AggregatorStore
			var execStore2 executorStorage.ExecutorStore
			var cleanup2 func()

			if tt.name == "BadgerDB" {
				// For BadgerDB, reuse the same directories to preserve data
				aggDir := os.Getenv("TEST_AGG_DIR")
				execDir := os.Getenv("TEST_EXEC_DIR")

				require.NotEmpty(t, aggDir, "TEST_AGG_DIR not set")
				require.NotEmpty(t, execDir, "TEST_EXEC_DIR not set")

				aggStore, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
					Dir: aggDir,
				})
				require.NoError(t, err)

				execStore, err := executorBadger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
					Dir: execDir,
				})
				require.NoError(t, err)

				aggStore2 = aggStore
				execStore2 = execStore
				cleanup2 = func() {
					aggStore.Close()
					execStore.Close()
				}
			} else {
				// Memory stores don't persist, so we simulate by copying
				aggStore2, execStore2, cleanup2 = tt.storeFactory(t)
			}
			defer cleanup2()

			// Verify state persisted (for BadgerDB)
			if tt.name == "BadgerDB" {
				// Check aggregator state
				block, err := aggStore2.GetLastProcessedBlock(ctx, avsAddress, chainId)
				require.NoError(t, err)
				require.Equal(t, uint64(1000), block)

				loadedTask, err := aggStore2.GetTask(ctx, "upgrade-task-1")
				require.NoError(t, err)
				require.Equal(t, task.TaskId, loadedTask.TaskId)

				// Check executor state
				loadedPerformer, err := execStore2.GetPerformerState(ctx, "performer-1")
				require.NoError(t, err)
				require.Equal(t, performer.PerformerId, loadedPerformer.PerformerId)
			}

			// Phase 3: Continue operations
			require.NoError(t, aggStore2.SetLastProcessedBlock(ctx, avsAddress, chainId, 2000))

			newTask := &types.Task{
				TaskId:                 "upgrade-task-2",
				AVSAddress:             "0x123",
				OperatorSetId:          2,
				SourceBlockNumber:      2000,
				L1ReferenceBlockNumber: 2000,
				ChainId:                config.ChainId(1),
			}
			require.NoError(t, aggStore2.SavePendingTask(ctx, newTask))
		})
	}
}

// TestStorageMigration tests migrating from one storage backend to another
func TestStorageMigration(t *testing.T) {
	ctx := context.Background()

	// Step 1: Create data in memory store
	memStore := aggregatorMemory.NewInMemoryAggregatorStore()

	// Populate with test data
	avsAddress := "0xtest"
	chainId := config.ChainId(1)
	require.NoError(t, memStore.SetLastProcessedBlock(ctx, avsAddress, chainId, 5000))

	tasks := []*types.Task{
		{TaskId: "task-1", AVSAddress: "0x123", OperatorSetId: 1, SourceBlockNumber: 4990, L1ReferenceBlockNumber: 4990, ChainId: chainId},
		{TaskId: "task-2", AVSAddress: "0x123", OperatorSetId: 2, SourceBlockNumber: 4995, L1ReferenceBlockNumber: 4995, ChainId: chainId},
		{TaskId: "task-3", AVSAddress: "0x123", OperatorSetId: 3, SourceBlockNumber: 5000, L1ReferenceBlockNumber: 5000, ChainId: chainId},
	}

	for _, task := range tasks {
		require.NoError(t, memStore.SavePendingTask(ctx, task))
	}
	// Proper status transitions: pending -> processing -> completed
	require.NoError(t, memStore.UpdateTaskStatus(ctx, "task-1", storage.TaskStatusProcessing))
	require.NoError(t, memStore.UpdateTaskStatus(ctx, "task-1", storage.TaskStatusCompleted))
	require.NoError(t, memStore.UpdateTaskStatus(ctx, "task-2", storage.TaskStatusProcessing))

	// Step 2: Migrate to BadgerDB
	badgerDir := t.TempDir()
	badgerStore, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
		Dir: badgerDir,
	})
	require.NoError(t, err)
	defer badgerStore.Close()

	// Migrate block heights
	block, err := memStore.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	require.NoError(t, badgerStore.SetLastProcessedBlock(ctx, avsAddress, chainId, block))

	// Migrate all tasks (both pending and non-pending)
	for _, task := range tasks {
		if loadedTask, err := memStore.GetTask(ctx, task.TaskId); err == nil {
			require.NoError(t, badgerStore.SavePendingTask(ctx, loadedTask))
			// Migrate status - need proper transitions
			if task.TaskId == "task-1" {
				// task-1 was already transitioned to completed in memory store
				require.NoError(t, badgerStore.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing))
				require.NoError(t, badgerStore.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted))
			} else if task.TaskId == "task-2" {
				// task-2 was already marked as processing in memory store
				require.NoError(t, badgerStore.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing))
			}
			// task-3 remains pending (no status update needed)
		}
	}

	memStore.Close()

	// Step 3: Verify migration
	migratedBlock, err := badgerStore.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	require.Equal(t, uint64(5000), migratedBlock)

	pendingTasks, err := badgerStore.ListPendingTasks(ctx)
	require.NoError(t, err)
	require.Len(t, pendingTasks, 1) // Only task-3 is pending
	require.Equal(t, "task-3", pendingTasks[0].TaskId)

	// Verify task statuses
	task1, err := badgerStore.GetTask(ctx, "task-1")
	require.NoError(t, err)
	require.NotNil(t, task1)

	task2, err := badgerStore.GetTask(ctx, "task-2")
	require.NoError(t, err)
	require.NotNil(t, task2)
}

// TestBackwardCompatibility ensures new code can read old data formats
func TestBackwardCompatibility(t *testing.T) {
	// This test would be expanded when there are actual schema changes
	// For now, we ensure the current format is stable

	ctx := context.Background()
	dir := t.TempDir()

	// Create store and save data
	store, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
		Dir: dir,
	})
	require.NoError(t, err)

	// Save various data types
	avsAddress := "0xtest"
	chainId := config.ChainId(1)
	require.NoError(t, store.SetLastProcessedBlock(ctx, avsAddress, chainId, 1000))

	task := &types.Task{
		TaskId:                 "compat-task",
		AVSAddress:             "0x123",
		OperatorSetId:          1,
		SourceBlockNumber:      1000,
		L1ReferenceBlockNumber: 1000,
		ChainId:                chainId,
		Payload:                []byte("test payload"),
	}
	require.NoError(t, store.SavePendingTask(ctx, task))

	opConfig := &storage.OperatorSetTaskConfig{
		TaskSLA:   3600,
		CurveType: config.CurveTypeBN254,
		Consensus: storage.OperatorSetTaskConsensus{
			ConsensusType: storage.ConsensusTypeStakeProportionThreshold,
			Threshold:     6600, // 66%
		},
	}
	require.NoError(t, store.SaveOperatorSetConfig(ctx, "0x123", 1, opConfig))

	avsConfig := &storage.AvsConfig{
		AggregatorOperatorSetId: 1,
		ExecutorOperatorSetIds:  []uint32{1, 2},
		CurveType:               config.CurveTypeBN254,
	}
	require.NoError(t, store.SaveAVSConfig(ctx, "0x123", avsConfig))

	store.Close()

	// Reopen and verify data integrity
	store2, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
		Dir: dir,
	})
	require.NoError(t, err)
	defer store2.Close()

	// Verify all data is readable
	block, err := store2.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	require.Equal(t, uint64(1000), block)

	loadedTask, err := store2.GetTask(ctx, "compat-task")
	require.NoError(t, err)
	require.Equal(t, task.TaskId, loadedTask.TaskId)
	require.Equal(t, task.Payload, loadedTask.Payload)

	loadedOpConfig, err := store2.GetOperatorSetConfig(ctx, "0x123", 1)
	require.NoError(t, err)
	require.Equal(t, opConfig.TaskSLA, loadedOpConfig.TaskSLA)

	loadedAvsConfig, err := store2.GetAVSConfig(ctx, "0x123")
	require.NoError(t, err)
	require.Equal(t, avsConfig.AggregatorOperatorSetId, loadedAvsConfig.AggregatorOperatorSetId)
}

// TestUpgradeUnderLoad simulates upgrading while system is under load
func TestUpgradeUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping upgrade under load test in short mode")
	}

	ctx := context.Background()
	dir := t.TempDir()

	// Start with first instance
	store1, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
		Dir: dir,
	})
	require.NoError(t, err)

	// Simulate ongoing operations
	done := make(chan bool)
	go func() {
		avsAddress := "0xtest"
		chainId := config.ChainId(1)
		blockNum := uint64(1000)
		taskId := 0

		for {
			select {
			case <-done:
				return
			default:
				// Write operations
				_ = store1.SetLastProcessedBlock(ctx, avsAddress, chainId, blockNum)
				task := &types.Task{
					TaskId:                 fmt.Sprintf("load-task-%d", taskId),
					AVSAddress:             "0x123",
					OperatorSetId:          uint32(taskId),
					SourceBlockNumber:      blockNum,
					L1ReferenceBlockNumber: blockNum,
					ChainId:                chainId,
				}
				_ = store1.SavePendingTask(ctx, task)

				blockNum++
				taskId++
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Let it run for a bit
	time.Sleep(1 * time.Second)

	// Get current state
	avsAddress := "0xtest"
	chainId := config.ChainId(1)
	lastBlock1, err := store1.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)

	// Simulate upgrade: close first store
	close(done)
	time.Sleep(100 * time.Millisecond) // Let operations finish
	store1.Close()

	// Open second instance (simulating upgraded version)
	store2, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
		Dir: dir,
	})
	require.NoError(t, err)
	defer store2.Close()

	// Verify state preserved
	lastBlock2, err := store2.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	require.GreaterOrEqual(t, lastBlock2, lastBlock1)

	// Verify can continue operations
	require.NoError(t, store2.SetLastProcessedBlock(ctx, avsAddress, chainId, lastBlock2+100))

	newTask := &types.Task{
		TaskId:                 "post-upgrade-task",
		AVSAddress:             "0x123",
		OperatorSetId:          9999,
		SourceBlockNumber:      lastBlock2 + 100,
		L1ReferenceBlockNumber: lastBlock2 + 100,
		ChainId:                chainId,
	}
	require.NoError(t, store2.SavePendingTask(ctx, newTask))

	// Verify task saved
	loadedTask, err := store2.GetTask(ctx, "post-upgrade-task")
	require.NoError(t, err)
	require.Equal(t, newTask.TaskId, loadedTask.TaskId)
}
