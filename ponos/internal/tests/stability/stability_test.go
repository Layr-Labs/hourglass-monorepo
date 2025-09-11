package stability

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"runtime"

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

func shouldRunExtendedTests(t *testing.T) {
	// Skip extended tests unless explicitly enabled via environment variable
	if os.Getenv("RUN_EXTENDED_TESTS") != "true" {
		t.Skip("Skipping extended stability test. Set RUN_EXTENDED_TESTS=true to run")
	}
}

// TestShortStability runs a short-duration stability test suitable for CI
func TestShortStability(t *testing.T) {
	shouldRunExtendedTests(t)

	tests := []struct {
		name         string
		duration     time.Duration
		storeFactory func(t *testing.T) (storage.AggregatorStore, executorStorage.ExecutorStore, func())
	}{
		{
			name:     "Memory_5Min",
			duration: 5 * time.Minute,
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
			name:     "BadgerDB_5Min",
			duration: 5 * time.Minute,
			storeFactory: func(t *testing.T) (storage.AggregatorStore, executorStorage.ExecutorStore, func()) {
				aggDir := t.TempDir()
				execDir := t.TempDir()

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
			runStabilityTest(t, tt.storeFactory, tt.duration)
		})
	}
}

// TestExtendedStability runs a long-duration stability test
// This simulates real-world usage patterns over an extended period
func TestExtendedStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping extended stability test in short mode")
	}

	shouldRunExtendedTests(t)

	tests := []struct {
		name         string
		duration     time.Duration
		storeFactory func(t *testing.T) (storage.AggregatorStore, executorStorage.ExecutorStore, func())
	}{
		{
			name:     "Memory_1Hour",
			duration: 1 * time.Hour,
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
			name:     "BadgerDB_24Hours",
			duration: 24 * time.Hour,
			storeFactory: func(t *testing.T) (storage.AggregatorStore, executorStorage.ExecutorStore, func()) {
				aggDir := t.TempDir()
				execDir := t.TempDir()

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
			runStabilityTest(t, tt.storeFactory, tt.duration)
		})
	}
}

// runStabilityTest executes the stability test with given parameters
func runStabilityTest(t *testing.T, storeFactory func(t *testing.T) (storage.AggregatorStore, executorStorage.ExecutorStore, func()), duration time.Duration) {
	aggStore, execStore, cleanup := storeFactory(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var stats struct {
		tasksCreated   int64
		tasksProcessed int64
		errors         int64
		blockProcessed int64
	}

	var wg sync.WaitGroup

	// Aggregator simulation
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulateAggregator(ctx, t, aggStore, &stats)
	}()

	// Executor simulation
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulateExecutor(ctx, t, execStore, &stats)
	}()

	// Monitor goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorProgress(ctx, &stats)
	}()

	wg.Wait()

	// Verify final state
	t.Logf("Final stats: tasks=%d, processed=%d, errors=%d, blocks=%d",
		atomic.LoadInt64(&stats.tasksCreated),
		atomic.LoadInt64(&stats.tasksProcessed),
		atomic.LoadInt64(&stats.errors),
		atomic.LoadInt64(&stats.blockProcessed),
	)

	require.Equal(t, int64(0), atomic.LoadInt64(&stats.errors), "No errors should occur")
}

func simulateAggregator(ctx context.Context, t *testing.T, store storage.AggregatorStore, stats *struct {
	tasksCreated   int64
	tasksProcessed int64
	errors         int64
	blockProcessed int64
}) {
	chainId := config.ChainId(1)
	blockNum := uint64(1000000)
	taskCounter := 0

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate block processing
			avsAddress := "0xtest"
			err := store.SetLastProcessedBlock(ctx, avsAddress, chainId, blockNum)
			if err != nil {
				t.Logf("Error setting block: %v", err)
				atomic.AddInt64(&stats.errors, 1)
				continue
			}
			atomic.AddInt64(&stats.blockProcessed, 1)

			// Create random tasks
			if rand.Intn(10) < 3 { // 30% chance
				task := &types.Task{
					TaskId:            fmt.Sprintf("task-%d", taskCounter),
					AVSAddress:        "0x123",
					OperatorSetId:     uint32(taskCounter),
					SourceBlockNumber: blockNum,
					ChainId:           chainId,
					Version:           1,
				}
				taskCounter++

				err = store.SavePendingTask(ctx, task)
				if err != nil {
					t.Logf("Error saving task: %v", err)
					atomic.AddInt64(&stats.errors, 1)
					continue
				}
				atomic.AddInt64(&stats.tasksCreated, 1)
			}

			// Process pending tasks
			pending, err := store.ListPendingTasks(ctx)
			if err != nil {
				t.Logf("Error listing tasks: %v", err)
				atomic.AddInt64(&stats.errors, 1)
				continue
			}

			for _, task := range pending {
				// First transition to processing
				err = store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing)
				if err != nil {
					t.Logf("Error updating task to processing: %v", err)
					atomic.AddInt64(&stats.errors, 1)
					continue
				}

				// Then transition to final state
				if rand.Intn(10) < 8 { // 80% success rate
					err = store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted)
				} else {
					err = store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusFailed)
				}
				if err != nil {
					t.Logf("Error updating task: %v", err)
					atomic.AddInt64(&stats.errors, 1)
					continue
				}
				atomic.AddInt64(&stats.tasksProcessed, 1)
			}

			blockNum++
		}
	}
}

func simulateExecutor(ctx context.Context, t *testing.T, store executorStorage.ExecutorStore, stats *struct {
	tasksCreated   int64
	tasksProcessed int64
	errors         int64
	blockProcessed int64
}) {
	performerCounter := 0
	deploymentCounter := 0

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate performer lifecycle
			if rand.Intn(10) < 2 { // 20% chance to create performer
				performerId := fmt.Sprintf("performer-%d", performerCounter)
				performerCounter++

				state := &executorStorage.PerformerState{
					PerformerId: performerId,
					AvsAddress:  "0xavs123",
					ContainerId: fmt.Sprintf("container-%s", performerId),
					Status:      "running",
					CreatedAt:   time.Now(),
				}

				err := store.SavePerformerState(ctx, performerId, state)
				if err != nil {
					t.Logf("Error saving performer: %v", err)
					atomic.AddInt64(&stats.errors, 1)
				}

				// Create deployment
				deploymentId := fmt.Sprintf("deploy-%d", deploymentCounter)
				deployment := &executorStorage.DeploymentInfo{
					DeploymentId:   deploymentId,
					AvsAddress:     state.AvsAddress,
					ArtifactDigest: "sha256:abc123",
					Status:         executorStorage.DeploymentStatusRunning,
					StartedAt:      time.Now(),
				}
				deploymentCounter++

				err = store.SaveDeployment(ctx, deployment.DeploymentId, deployment)
				if err != nil {
					t.Logf("Error saving deployment: %v", err)
					atomic.AddInt64(&stats.errors, 1)
				}
			}

			// Simulate task execution
			if rand.Intn(10) < 5 { // 50% chance
				taskId := fmt.Sprintf("exec-task-%d", rand.Intn(1000))
				task := &executorStorage.TaskInfo{
					TaskId:          taskId,
					AvsAddress:      "0xavs123",
					OperatorAddress: fmt.Sprintf("operator-%d", rand.Intn(performerCounter+1)),
					ReceivedAt:      time.Now(),
				}

				err := store.SaveInflightTask(ctx, taskId, task)
				if err != nil {
					t.Logf("Error saving inflight task: %v", err)
					atomic.AddInt64(&stats.errors, 1)
				}

				// Complete task
				if rand.Intn(10) < 9 { // 90% completion rate
					err = store.DeleteInflightTask(ctx, taskId)
					if err != nil {
						t.Logf("Error deleting inflight task: %v", err)
						atomic.AddInt64(&stats.errors, 1)
					}
				}
			}
		}
	}
}

func monitorProgress(ctx context.Context, stats *struct {
	tasksCreated   int64
	tasksProcessed int64
	errors         int64
	blockProcessed int64
}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Printf("[%s] Progress: tasks=%d, processed=%d, errors=%d, blocks=%d\n",
				time.Now().Format("15:04:05"),
				atomic.LoadInt64(&stats.tasksCreated),
				atomic.LoadInt64(&stats.tasksProcessed),
				atomic.LoadInt64(&stats.errors),
				atomic.LoadInt64(&stats.blockProcessed),
			)
		}
	}
}

// TestConcurrentLoad tests high concurrent load scenarios
func TestConcurrentLoad(t *testing.T) {
	shouldRunExtendedTests(t)
	stores := []struct {
		name         string
		storeFactory func(t *testing.T) (storage.AggregatorStore, func())
	}{
		{
			name: "Memory",
			storeFactory: func(t *testing.T) (storage.AggregatorStore, func()) {
				store := aggregatorMemory.NewInMemoryAggregatorStore()
				return store, func() { store.Close() }
			},
		},
		{
			name: "BadgerDB",
			storeFactory: func(t *testing.T) (storage.AggregatorStore, func()) {
				dir := t.TempDir()
				store, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
					Dir: dir,
				})
				require.NoError(t, err)
				return store, func() { store.Close() }
			},
		},
	}

	for _, tt := range stores {
		t.Run(tt.name, func(t *testing.T) {
			store, cleanup := tt.storeFactory(t)
			defer cleanup()

			ctx := context.Background()
			numGoroutines := 20
			numOperations := 100

			var wg sync.WaitGroup
			errors := make(chan error, numGoroutines*numOperations)

			// Launch concurrent operations
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < numOperations; j++ {
						// Mix of operations
						switch rand.Intn(4) {
						case 0: // Save task
							task := &types.Task{
								TaskId:        fmt.Sprintf("task-%d-%d", id, j),
								AVSAddress:    "0x123",
								OperatorSetId: uint32(j),
								Version:       1,
							}
							if err := store.SavePendingTask(ctx, task); err != nil {
								errors <- err
							}
						case 1: // Update block
							avsAddress := "0xtest"
							chainId := config.ChainId(rand.Intn(3) + 1)
							blockNum := uint64(rand.Intn(1000000))
							if err := store.SetLastProcessedBlock(ctx, avsAddress, chainId, blockNum); err != nil {
								errors <- err
							}
						case 2: // List tasks
							if _, err := store.ListPendingTasks(ctx); err != nil {
								errors <- err
							}
						case 3: // Update task status
							taskId := fmt.Sprintf("task-%d-%d", rand.Intn(numGoroutines), rand.Intn(j+1))
							// Try different status transitions randomly
							statusChoice := rand.Intn(3)
							var err error
							switch statusChoice {
							case 0:
								err = store.UpdateTaskStatus(ctx, taskId, storage.TaskStatusProcessing)
							case 1:
								err = store.UpdateTaskStatus(ctx, taskId, storage.TaskStatusCompleted)
							case 2:
								err = store.UpdateTaskStatus(ctx, taskId, storage.TaskStatusFailed)
							}

							if err != nil {
								// Ignore expected errors in concurrent scenarios
								errStr := err.Error()
								isExpected := err == storage.ErrNotFound ||
									strings.Contains(errStr, "invalid status transition") ||
									strings.Contains(errStr, "cannot transition") ||
									strings.Contains(errStr, "Transaction Conflict")

								if !isExpected {
									errors <- err
								}
							}
						}
					}
				}(i)
			}

			wg.Wait()
			close(errors)

			// Check for errors
			var errorCount int
			for err := range errors {
				t.Logf("Concurrent operation error: %v", err)
				errorCount++
			}

			require.Equal(t, 0, errorCount, "No errors should occur during concurrent operations")
		})
	}
}

// TestMemoryLeaks checks for memory leaks over extended operation
func TestMemoryLeaks(t *testing.T) {
	shouldRunExtendedTests(t)

	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	store := aggregatorMemory.NewInMemoryAggregatorStore()
	defer store.Close()

	ctx := context.Background()
	iterations := 100000

	// Force GC and measure initial memory
	runtime.GC()
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Perform many operations
	for i := 0; i < iterations; i++ {
		task := &types.Task{
			TaskId:        fmt.Sprintf("task-%d", i),
			AVSAddress:    "0x123",
			OperatorSetId: uint32(i),
			Version:       1,
		}

		// Create and delete tasks
		require.NoError(t, store.SavePendingTask(ctx, task))
		require.NoError(t, store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing))
		require.NoError(t, store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted))

		// Periodically delete old tasks
		if i%1000 == 0 && i > 0 {
			require.NoError(t, store.DeleteTask(ctx, fmt.Sprintf("task-%d", i-1000)))
		}
	}

	// Force GC and measure memory
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// Memory should not grow excessively
	// Use HeapAlloc which is more stable for this comparison
	memGrowth := int64(m2.HeapAlloc) - int64(m1.HeapAlloc)
	t.Logf("Memory growth: %d bytes", memGrowth)

	// Allow some growth but flag excessive growth
	maxGrowth := int64(50 * 1024 * 1024) // 50MB
	if memGrowth > maxGrowth {
		t.Errorf("Excessive memory growth: %d bytes (max: %d)", memGrowth, maxGrowth)
	}
}
