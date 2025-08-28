package persistence_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	executorStorage "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	executorMemory "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

// Baseline sync.Map implementation for comparison
type syncMapBaseline struct {
	tasks   *sync.Map
	blocks  *sync.Map
	configs *sync.Map
}

func newSyncMapBaseline() *syncMapBaseline {
	return &syncMapBaseline{
		tasks:   &sync.Map{},
		blocks:  &sync.Map{},
		configs: &sync.Map{},
	}
}

// Helper to create a task
func createTask(id string) *types.Task {
	deadline := time.Now().Add(time.Hour)
	return &types.Task{
		TaskId:              id,
		AVSAddress:          "0xavs1",
		OperatorSetId:       1,
		CallbackAddr:        "0xcallback",
		DeadlineUnixSeconds: &deadline,
		ThresholdBips:       6700,
		Payload:             []byte(fmt.Sprintf("payload-%s", id)),
		ChainId:             config.ChainId(1),
		SourceBlockNumber:   1000,
		BlockHash:           fmt.Sprintf("0xhash%s", id),
	}
}

// Benchmark task operations
func BenchmarkTaskOperations(b *testing.B) {
	ctx := context.Background()

	b.Run("InMemoryStore/SavePendingTask", func(b *testing.B) {
		store := memory.NewInMemoryAggregatorStore()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			task := createTask(fmt.Sprintf("task-%d", i))
			_ = store.SavePendingTask(ctx, task)
		}
	})

	b.Run("SyncMap/SavePendingTask", func(b *testing.B) {
		baseline := newSyncMapBaseline()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			task := createTask(fmt.Sprintf("task-%d", i))
			baseline.tasks.Store(task.TaskId, task)
		}
	})

	b.Run("InMemoryStore/GetTask", func(b *testing.B) {
		store := memory.NewInMemoryAggregatorStore()
		// Pre-populate
		for i := 0; i < 1000; i++ {
			task := createTask(fmt.Sprintf("task-%d", i))
			_ = store.SavePendingTask(ctx, task)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetTask(ctx, fmt.Sprintf("task-%d", i%1000))
		}
	})

	b.Run("SyncMap/GetTask", func(b *testing.B) {
		baseline := newSyncMapBaseline()
		// Pre-populate
		for i := 0; i < 1000; i++ {
			task := createTask(fmt.Sprintf("task-%d", i))
			baseline.tasks.Store(task.TaskId, task)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = baseline.tasks.Load(fmt.Sprintf("task-%d", i%1000))
		}
	})

	b.Run("InMemoryStore/ListPendingTasks", func(b *testing.B) {
		store := memory.NewInMemoryAggregatorStore()
		// Pre-populate with mixed statuses
		for i := 0; i < 1000; i++ {
			task := createTask(fmt.Sprintf("task-%d", i))
			_ = store.SavePendingTask(ctx, task)
			if i%3 == 0 {
				_ = store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted)
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.ListPendingTasks(ctx)
		}
	})
}

// Benchmark concurrent operations
func BenchmarkConcurrentOperations(b *testing.B) {
	ctx := context.Background()

	b.Run("InMemoryStore/ConcurrentReadWrite", func(b *testing.B) {
		store := memory.NewInMemoryAggregatorStore()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					task := createTask(fmt.Sprintf("task-%d-%d", i, time.Now().UnixNano()))
					_ = store.SavePendingTask(ctx, task)
				} else {
					_, _ = store.ListPendingTasks(ctx)
				}
				i++
			}
		})
	})

	b.Run("SyncMap/ConcurrentReadWrite", func(b *testing.B) {
		baseline := newSyncMapBaseline()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					task := createTask(fmt.Sprintf("task-%d-%d", i, time.Now().UnixNano()))
					baseline.tasks.Store(task.TaskId, task)
				} else {
					// Simulate list by iterating
					count := 0
					baseline.tasks.Range(func(_, _ interface{}) bool {
						count++
						return true
					})
				}
				i++
			}
		})
	})
}

// Benchmark block operations
func BenchmarkBlockOperations(b *testing.B) {
	ctx := context.Background()

	b.Run("InMemoryStore/SetBlock", func(b *testing.B) {
		store := memory.NewInMemoryAggregatorStore()
		chain := config.ChainId(1)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = store.SetLastProcessedBlock(ctx, chain, uint64(i))
		}
	})

	b.Run("SyncMap/SetBlock", func(b *testing.B) {
		baseline := newSyncMapBaseline()
		chain := config.ChainId(1)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("chain-%d", chain)
			baseline.blocks.Store(key, uint64(i))
		}
	})
}

// Benchmark executor operations
func BenchmarkExecutorOperations(b *testing.B) {
	ctx := context.Background()

	b.Run("InMemoryExecutor/SavePerformer", func(b *testing.B) {
		store := executorMemory.NewInMemoryExecutorStore()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			performer := &executorStorage.PerformerState{
				PerformerId:        fmt.Sprintf("performer-%d", i),
				AvsAddress:         "0xavs1",
				ContainerId:        fmt.Sprintf("container-%d", i),
				Status:             "running",
				ArtifactRegistry:   "registry.io/avs",
				ArtifactTag:        "v1.0.0",
				ArtifactDigest:     "sha256:abc123",
				DeploymentMode:     "docker",
				CreatedAt:          time.Now(),
				LastHealthCheck:    time.Now(),
				ContainerHealthy:   true,
				ApplicationHealthy: true,
			}
			_ = store.SavePerformerState(ctx, performer.PerformerId, performer)
		}
	})

	b.Run("InMemoryExecutor/ListPerformers", func(b *testing.B) {
		store := executorMemory.NewInMemoryExecutorStore()
		// Pre-populate
		for i := 0; i < 100; i++ {
			performer := &executorStorage.PerformerState{
				PerformerId:        fmt.Sprintf("performer-%d", i),
				AvsAddress:         fmt.Sprintf("0xavs%d", i%10),
				ContainerId:        fmt.Sprintf("container-%d", i),
				Status:             "running",
				ArtifactRegistry:   "registry.io/avs",
				ArtifactTag:        "v1.0.0",
				ArtifactDigest:     "sha256:abc123",
				DeploymentMode:     "docker",
				CreatedAt:          time.Now(),
				LastHealthCheck:    time.Now(),
				ContainerHealthy:   true,
				ApplicationHealthy: true,
			}
			_ = store.SavePerformerState(ctx, performer.PerformerId, performer)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.ListPerformerStates(ctx)
		}
	})
}

// Benchmark memory usage
func BenchmarkMemoryUsage(b *testing.B) {
	ctx := context.Background()

	b.Run("InMemoryStore/LargeDataset", func(b *testing.B) {
		store := memory.NewInMemoryAggregatorStore()
		b.ResetTimer()

		// Populate with large dataset
		for i := 0; i < 10000; i++ {
			task := createTask(fmt.Sprintf("task-%d", i))
			task.AVSAddress = fmt.Sprintf("0xavs%d", i%100)
			task.OperatorSetId = uint32(i % 10)
			task.Payload = []byte(fmt.Sprintf("payload-data-%d", i))
			_ = store.SavePendingTask(ctx, task)

			if i%100 == 0 {
				config := &storage.OperatorSetTaskConfig{
					TaskSLA:      3600,
					CurveType:    config.CurveTypeBN254,
					TaskMetadata: []byte(fmt.Sprintf("metadata-%d", i)),
					Consensus: storage.OperatorSetTaskConsensus{
						ConsensusType: storage.ConsensusTypeStakeProportionThreshold,
						Threshold:     6700,
					},
				}
				_ = store.SaveOperatorSetConfig(ctx, fmt.Sprintf("0xavs%d", i%100), uint32(i%10), config)
			}
		}

		// Perform operations on large dataset
		for i := 0; i < b.N; i++ {
			_, _ = store.ListPendingTasks(ctx)
		}
	})
}

// Benchmark config operations
func BenchmarkConfigOperations(b *testing.B) {
	ctx := context.Background()

	b.Run("InMemoryStore/SaveOperatorSetConfig", func(b *testing.B) {
		store := memory.NewInMemoryAggregatorStore()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opConfig := &storage.OperatorSetTaskConfig{
				TaskSLA:      int64(3600 + i),
				CurveType:    config.CurveTypeBN254,
				TaskMetadata: []byte(fmt.Sprintf("metadata-%d", i)),
				Consensus: storage.OperatorSetTaskConsensus{
					ConsensusType: storage.ConsensusTypeStakeProportionThreshold,
					Threshold:     uint16(6700 + i%100),
				},
			}
			_ = store.SaveOperatorSetConfig(ctx, fmt.Sprintf("0xavs%d", i%10), uint32(i%5), opConfig)
		}
	})

	b.Run("SyncMap/SaveOperatorSetConfig", func(b *testing.B) {
		baseline := newSyncMapBaseline()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opConfig := &storage.OperatorSetTaskConfig{
				TaskSLA:      int64(3600 + i),
				CurveType:    config.CurveTypeBN254,
				TaskMetadata: []byte(fmt.Sprintf("metadata-%d", i)),
				Consensus: storage.OperatorSetTaskConsensus{
					ConsensusType: storage.ConsensusTypeStakeProportionThreshold,
					Threshold:     uint16(6700 + i%100),
				},
			}
			key := fmt.Sprintf("opset-%s-%d", fmt.Sprintf("0xavs%d", i%10), i%5)
			baseline.configs.Store(key, opConfig)
		}
	})
}
