package performance

import (
	"context"
	"fmt"
	"sync"
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
)

// Benchmark individual operations
func BenchmarkAggregatorOperations(b *testing.B) {
	stores := []struct {
		name         string
		storeFactory func(b *testing.B) (storage.AggregatorStore, func())
	}{
		{
			name: "Memory",
			storeFactory: func(b *testing.B) (storage.AggregatorStore, func()) {
				store := aggregatorMemory.NewInMemoryAggregatorStore()
				return store, func() { store.Close() }
			},
		},
		{
			name: "BadgerDB",
			storeFactory: func(b *testing.B) (storage.AggregatorStore, func()) {
				dir := b.TempDir()
				store, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
					Dir: dir,
				})
				if err != nil {
					b.Fatal(err)
				}
				return store, func() { store.Close() }
			},
		},
	}

	operations := []struct {
		name string
		fn   func(b *testing.B, store storage.AggregatorStore)
	}{
		{
			name: "SavePendingTask",
			fn:   benchmarkSaveTask,
		},
		{
			name: "GetTask",
			fn:   benchmarkGetTask,
		},
		{
			name: "UpdateTaskStatus",
			fn:   benchmarkUpdateTaskStatus,
		},
		{
			name: "ListPendingTasks",
			fn:   benchmarkListPendingTasks,
		},
		{
			name: "SetLastProcessedBlock",
			fn:   benchmarkSetLastProcessedBlock,
		},
	}

	for _, st := range stores {
		b.Run(st.name, func(b *testing.B) {
			for _, op := range operations {
				b.Run(op.name, func(b *testing.B) {
					store, cleanup := st.storeFactory(b)
					defer cleanup()
					op.fn(b, store)
				})
			}
		})
	}
}

func benchmarkSaveTask(b *testing.B, store storage.AggregatorStore) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &types.Task{
			TaskId:                 fmt.Sprintf("bench-task-%d", i),
			AVSAddress:             "0x123",
			OperatorSetId:          uint32(i),
			SourceBlockNumber:      uint64(i / 100),
			L1ReferenceBlockNumber: uint64(i / 100),
			ReferenceTimestamp:     100,
			ChainId:                config.ChainId(1),
			Payload:                make([]byte, 256), // 256 bytes payload
			Version:                1,
		}
		if err := store.SavePendingTask(ctx, task); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkGetTask(b *testing.B, store storage.AggregatorStore) {
	ctx := context.Background()

	// Prepare data
	numTasks := 10000
	for i := 0; i < numTasks; i++ {
		task := &types.Task{
			TaskId:             fmt.Sprintf("get-task-%d", i),
			AVSAddress:         "0x123",
			OperatorSetId:      uint32(i),
			ChainId:            config.ChainId(1),
			ReferenceTimestamp: 100,
			Version:            1,
		}
		if err := store.SavePendingTask(ctx, task); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		taskId := fmt.Sprintf("get-task-%d", i%numTasks)
		if _, err := store.GetTask(ctx, taskId); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkUpdateTaskStatus(b *testing.B, store storage.AggregatorStore) {
	ctx := context.Background()

	// Prepare data
	numTasks := 10000
	for i := 0; i < numTasks; i++ {
		task := &types.Task{
			TaskId:             fmt.Sprintf("update-task-%d", i),
			AVSAddress:         "0x123",
			OperatorSetId:      uint32(i),
			ReferenceTimestamp: 100,
			ChainId:            config.ChainId(1),
			Version:            1,
		}
		if err := store.SavePendingTask(ctx, task); err != nil {
			b.Fatal(err)
		}
	}

	statuses := []storage.TaskStatus{
		storage.TaskStatusProcessing,
		storage.TaskStatusCompleted,
		storage.TaskStatusFailed,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		taskId := fmt.Sprintf("update-task-%d", i%numTasks)
		status := statuses[i%len(statuses)]
		if err := store.UpdateTaskStatus(ctx, taskId, status); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkListPendingTasks(b *testing.B, store storage.AggregatorStore) {
	ctx := context.Background()

	// Prepare data with mix of statuses
	numTasks := 10000
	for i := 0; i < numTasks; i++ {
		task := &types.Task{
			TaskId:             fmt.Sprintf("list-task-%d", i),
			AVSAddress:         "0x123",
			OperatorSetId:      uint32(i),
			ReferenceTimestamp: 100,
			ChainId:            config.ChainId(1),
			Version:            1,
		}
		if err := store.SavePendingTask(ctx, task); err != nil {
			b.Fatal(err)
		}

		// Make 30% completed
		if i%3 == 0 {
			if err := store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted); err != nil {
				b.Fatal(err)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := store.ListPendingTasks(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkSetLastProcessedBlock(b *testing.B, store storage.AggregatorStore) {
	ctx := context.Background()
	avsAddress := "0xtest"
	chainId := config.ChainId(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := store.SetLastProcessedBlock(ctx, avsAddress, chainId, uint64(i)); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark concurrent operations
func BenchmarkConcurrentOperations(b *testing.B) {
	stores := []struct {
		name         string
		storeFactory func(b *testing.B) (storage.AggregatorStore, func())
	}{
		{
			name: "Memory",
			storeFactory: func(b *testing.B) (storage.AggregatorStore, func()) {
				store := aggregatorMemory.NewInMemoryAggregatorStore()
				return store, func() { store.Close() }
			},
		},
		{
			name: "BadgerDB",
			storeFactory: func(b *testing.B) (storage.AggregatorStore, func()) {
				dir := b.TempDir()
				store, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
					Dir: dir,
				})
				if err != nil {
					b.Fatal(err)
				}
				return store, func() { store.Close() }
			},
		},
	}

	concurrencyLevels := []int{1, 10, 50, 100}

	for _, st := range stores {
		b.Run(st.name, func(b *testing.B) {
			for _, concurrency := range concurrencyLevels {
				b.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(b *testing.B) {
					store, cleanup := st.storeFactory(b)
					defer cleanup()
					benchmarkConcurrentMixedOps(b, store, concurrency)
				})
			}
		})
	}
}

func benchmarkConcurrentMixedOps(b *testing.B, store storage.AggregatorStore, concurrency int) {
	ctx := context.Background()

	// Pre-populate some data
	for i := 0; i < 1000; i++ {
		task := &types.Task{
			TaskId:             fmt.Sprintf("pre-task-%d", i),
			AVSAddress:         "0x123",
			OperatorSetId:      uint32(i),
			ReferenceTimestamp: 100,
			ChainId:            config.ChainId(1),
			Version:            1,
		}
		if err := store.SavePendingTask(ctx, task); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	var wg sync.WaitGroup
	opsPerGoroutine := b.N / concurrency

	for g := 0; g < concurrency; g++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()

			for i := 0; i < opsPerGoroutine; i++ {
				// Mix of operations
				switch i % 4 {
				case 0: // Save new task
					task := &types.Task{
						TaskId:             fmt.Sprintf("concurrent-task-%d-%d", goroutineId, i),
						AVSAddress:         "0x123",
						OperatorSetId:      uint32(i),
						ReferenceTimestamp: 100,
						ChainId:            config.ChainId(1),
						Version:            1,
					}
					if err := store.SavePendingTask(ctx, task); err != nil {
						b.Error(err)
					}
				case 1: // Get existing task
					taskId := fmt.Sprintf("pre-task-%d", i%1000)
					if _, err := store.GetTask(ctx, taskId); err != nil && err != storage.ErrNotFound {
						b.Error(err)
					}
				case 2: // Update status
					taskId := fmt.Sprintf("pre-task-%d", i%1000)
					if err := store.UpdateTaskStatus(ctx, taskId, storage.TaskStatusProcessing); err != nil && err != storage.ErrNotFound {
						b.Error(err)
					}
				case 3: // Update block
					avsAddress := "0xtest"
					chainId := config.ChainId(goroutineId%3 + 1)
					if err := store.SetLastProcessedBlock(ctx, avsAddress, chainId, uint64(i)); err != nil {
						b.Error(err)
					}
				}
			}
		}(g)
	}

	wg.Wait()
}

// Benchmark executor operations
func BenchmarkExecutorOperations(b *testing.B) {
	stores := []struct {
		name         string
		storeFactory func(b *testing.B) (executorStorage.ExecutorStore, func())
	}{
		{
			name: "Memory",
			storeFactory: func(b *testing.B) (executorStorage.ExecutorStore, func()) {
				store := executorMemory.NewInMemoryExecutorStore()
				return store, func() { store.Close() }
			},
		},
		{
			name: "BadgerDB",
			storeFactory: func(b *testing.B) (executorStorage.ExecutorStore, func()) {
				dir := b.TempDir()
				store, err := executorBadger.NewBadgerExecutorStore(&executorConfig.BadgerConfig{
					Dir: dir,
				})
				if err != nil {
					b.Fatal(err)
				}
				return store, func() { store.Close() }
			},
		},
	}

	for _, st := range stores {
		b.Run(st.name, func(b *testing.B) {
			store, cleanup := st.storeFactory(b)
			defer cleanup()

			b.Run("SavePerformerState", func(b *testing.B) {
				benchmarkSavePerformerState(b, store)
			})
		})
	}
}

func benchmarkSavePerformerState(b *testing.B, store executorStorage.ExecutorStore) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		performerId := fmt.Sprintf("performer-%d", i)
		state := &executorStorage.PerformerState{
			PerformerId: performerId,
			AvsAddress:  "0xavs123",
			ContainerId: fmt.Sprintf("container-%d", i),
			Status:      "running",
			CreatedAt:   time.Now(),
		}
		if err := store.SavePerformerState(ctx, performerId, state); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark memory usage
func BenchmarkMemoryUsage(b *testing.B) {
	stores := []struct {
		name         string
		storeFactory func(b *testing.B) (storage.AggregatorStore, func())
	}{
		{
			name: "Memory-10K",
			storeFactory: func(b *testing.B) (storage.AggregatorStore, func()) {
				store := aggregatorMemory.NewInMemoryAggregatorStore()
				return store, func() { store.Close() }
			},
		},
		{
			name: "BadgerDB-10K",
			storeFactory: func(b *testing.B) (storage.AggregatorStore, func()) {
				dir := b.TempDir()
				store, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
					Dir: dir,
				})
				if err != nil {
					b.Fatal(err)
				}
				return store, func() { store.Close() }
			},
		},
	}

	taskCounts := []int{1000, 10000, 100000}

	for _, st := range stores {
		for _, count := range taskCounts {
			b.Run(fmt.Sprintf("%s-Tasks-%d", st.name, count), func(b *testing.B) {
				store, cleanup := st.storeFactory(b)
				defer cleanup()
				benchmarkMemoryWithTasks(b, store, count)
			})
		}
	}
}

func benchmarkMemoryWithTasks(b *testing.B, store storage.AggregatorStore, taskCount int) {
	ctx := context.Background()

	// Populate store
	for i := 0; i < taskCount; i++ {
		task := &types.Task{
			TaskId:             fmt.Sprintf("mem-task-%d", i),
			AVSAddress:         "0x123",
			OperatorSetId:      uint32(i),
			ReferenceTimestamp: 100,
			ChainId:            config.ChainId(1),
			Payload:            make([]byte, 1024), // 1KB payload
		}
		if err := store.SavePendingTask(ctx, task); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	// Perform operations to test memory behavior
	for i := 0; i < b.N; i++ {
		// Read random task
		taskId := fmt.Sprintf("mem-task-%d", i%taskCount)
		if _, err := store.GetTask(ctx, taskId); err != nil {
			b.Fatal(err)
		}

		// Occasionally list tasks
		if i%100 == 0 {
			if _, err := store.ListPendingTasks(ctx); err != nil {
				b.Fatal(err)
			}
		}
	}
}
