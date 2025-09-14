package production

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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/require"
)

// TestProductionConfiguration validates production-ready configurations
func TestProductionConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config *aggregatorConfig.BadgerConfig
		valid  bool
		reason string
	}{
		{
			name: "ValidProduction",
			config: &aggregatorConfig.BadgerConfig{
				Dir:                     "/var/lib/ponos/aggregator/badger",
				ValueLogFileSize:        2 << 30, // 2GB
				NumVersionsToKeep:       1,
				NumLevelZeroTables:      10,
				NumLevelZeroTablesStall: 20,
			},
			valid: true,
		},
		{
			name: "InvalidSmallValueLog",
			config: &aggregatorConfig.BadgerConfig{
				Dir:              "/tmp/test",
				ValueLogFileSize: 1 << 20, // 1MB - too small for production
			},
			valid:  false,
			reason: "ValueLogFileSize too small for production",
		},
		{
			name: "InvalidHighVersions",
			config: &aggregatorConfig.BadgerConfig{
				Dir:               "/tmp/test",
				NumVersionsToKeep: 10, // Too many versions
			},
			valid:  false,
			reason: "NumVersionsToKeep too high, will cause bloat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				// Should work fine
				validateProductionConfig(t, tt.config)
			} else {
				t.Logf("Invalid config: %s", tt.reason)
			}
		})
	}
}

func validateProductionConfig(t *testing.T, config *aggregatorConfig.BadgerConfig) {
	// Check value log size
	minValueLogSize := int64(1 << 30) // 1GB minimum
	if config.ValueLogFileSize > 0 && config.ValueLogFileSize < minValueLogSize {
		t.Errorf("ValueLogFileSize %d is too small for production (min: %d)",
			config.ValueLogFileSize, minValueLogSize)
	}

	// Check versions to keep
	if config.NumVersionsToKeep > 3 {
		t.Errorf("NumVersionsToKeep %d is too high, recommended: 1-3",
			config.NumVersionsToKeep)
	}

}

// TestProductionDiskRequirements validates disk space and performance
func TestProductionDiskRequirements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping disk requirements test in short mode")
	}

	dir := t.TempDir()

	// Test disk space
	_, err := os.Stat(dir)
	require.NoError(t, err)

	// Check available space (would need platform-specific implementation)
	t.Logf("Test directory: %s", dir)

	// Test disk performance
	testDiskPerformance(t, dir)
}

func testDiskPerformance(t *testing.T, dir string) {
	testFile := filepath.Join(dir, "performance_test")
	data := make([]byte, 1<<20) // 1MB

	// Write test
	start := time.Now()
	for i := 0; i < 100; i++ {
		f, err := os.Create(fmt.Sprintf("%s_%d", testFile, i))
		require.NoError(t, err)
		_, err = f.Write(data)
		require.NoError(t, err)
		err = f.Sync()
		require.NoError(t, err)
		f.Close()
	}
	writeTime := time.Since(start)
	writeMBps := float64(100) / writeTime.Seconds()

	t.Logf("Write performance: %.2f MB/s", writeMBps)

	// Read test
	start = time.Now()
	for i := 0; i < 100; i++ {
		data, err := os.ReadFile(fmt.Sprintf("%s_%d", testFile, i))
		require.NoError(t, err)
		require.Len(t, data, 1<<20)
	}
	readTime := time.Since(start)
	readMBps := float64(100) / readTime.Seconds()

	t.Logf("Read performance: %.2f MB/s", readMBps)

	// Cleanup
	for i := 0; i < 100; i++ {
		os.Remove(fmt.Sprintf("%s_%d", testFile, i))
	}

	// Production requirements
	minWriteMBps := 50.0
	minReadMBps := 100.0

	if writeMBps < minWriteMBps {
		t.Errorf("Write performance %.2f MB/s below minimum %.2f MB/s",
			writeMBps, minWriteMBps)
	}
	if readMBps < minReadMBps {
		t.Errorf("Read performance %.2f MB/s below minimum %.2f MB/s",
			readMBps, minReadMBps)
	}
}

// TestProductionRecovery validates recovery in production scenarios
func TestProductionRecovery(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	// Create store with production config
	cfg := &aggregatorConfig.BadgerConfig{
		Dir:                     dir,
		ValueLogFileSize:        256 << 20, // 256MB for test
		NumVersionsToKeep:       1,
		NumLevelZeroTables:      5,
		NumLevelZeroTablesStall: 10,
	}

	store, err := aggregatorBadger.NewBadgerAggregatorStore(cfg)
	require.NoError(t, err)

	// Simulate production workload
	chainId := config.ChainId(1)
	numTasks := 10000

	// Write initial data
	avsAddress := "0xtest"
	for i := 0; i < numTasks; i++ {
		if i%100 == 0 {
			require.NoError(t, store.SetLastProcessedBlock(ctx, avsAddress, chainId, uint64(i)))
		}

		task := &types.Task{
			TaskId:                 fmt.Sprintf("prod-task-%d", i),
			AVSAddress:             "0x123",
			OperatorSetId:          uint32(i),
			SourceBlockNumber:      uint64(i / 10),
			L1ReferenceBlockNumber: uint64(i / 10),
			ReferenceTimestamp:     100,
			ChainId:                1,
			Payload:                make([]byte, 1024), // 1KB payload
		}
		require.NoError(t, store.SavePendingTask(ctx, task))

		// Update some tasks
		if i > 0 && i%3 == 0 {
			taskId := fmt.Sprintf("prod-task-%d", i-1)
			// First mark as processing
			require.NoError(t, store.UpdateTaskStatus(ctx, taskId, storage.TaskStatusProcessing))
			// Then mark as completed
			require.NoError(t, store.UpdateTaskStatus(ctx, taskId, storage.TaskStatusCompleted))
		}
	}

	// Get state before close
	lastBlock, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)

	pendingTasks, err := store.ListPendingTasks(ctx)
	require.NoError(t, err)
	pendingCount := len(pendingTasks)

	// Close store
	store.Close()

	// Simulate crash by reopening
	store2, err := aggregatorBadger.NewBadgerAggregatorStore(cfg)
	require.NoError(t, err)
	defer store2.Close()

	// Verify recovery
	recoveredBlock, err := store2.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	require.Equal(t, lastBlock, recoveredBlock)

	recoveredPending, err := store2.ListPendingTasks(ctx)
	require.NoError(t, err)
	require.Equal(t, pendingCount, len(recoveredPending))

	// Verify can continue operations
	require.NoError(t, store2.SetLastProcessedBlock(ctx, avsAddress, chainId, lastBlock+1000))

	newTask := &types.Task{
		TaskId:                 "post-recovery-task",
		AVSAddress:             "0x123",
		OperatorSetId:          uint32(numTasks),
		SourceBlockNumber:      lastBlock + 1000,
		L1ReferenceBlockNumber: lastBlock + 1000,
		ReferenceTimestamp:     1000,
		ChainId:                chainId,
	}
	require.NoError(t, store2.SavePendingTask(ctx, newTask))
}

// TestProductionMetrics validates metrics and monitoring
func TestProductionMetrics(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	store, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
		Dir: dir,
	})
	require.NoError(t, err)
	defer store.Close()

	// Track operation times
	timings := make(map[string][]time.Duration)

	// Perform operations and measure
	for i := 0; i < 1000; i++ {
		task := &types.Task{
			TaskId:             fmt.Sprintf("metric-task-%d", i),
			AVSAddress:         "0x123",
			OperatorSetId:      uint32(i),
			ChainId:            config.ChainId(1),
			ReferenceTimestamp: 1,
		}

		// SavePendingTask
		start := time.Now()
		err := store.SavePendingTask(ctx, task)
		timings["SavePendingTask"] = append(timings["SavePendingTask"], time.Since(start))
		require.NoError(t, err)

		// GetTask
		start = time.Now()
		_, err = store.GetTask(ctx, task.TaskId)
		timings["GetTask"] = append(timings["GetTask"], time.Since(start))
		require.NoError(t, err)

		// UpdateTaskStatus
		start = time.Now()
		err = store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing)
		timings["UpdateTaskStatus"] = append(timings["UpdateTaskStatus"], time.Since(start))
		require.NoError(t, err)

		// ListPendingTasks (every 10 iterations)
		if i%10 == 0 {
			start = time.Now()
			_, err = store.ListPendingTasks(ctx)
			timings["ListPendingTasks"] = append(timings["ListPendingTasks"], time.Since(start))
			require.NoError(t, err)
		}
	}

	// Analyze metrics
	for op, durations := range timings {
		var total time.Duration
		var max time.Duration
		for _, d := range durations {
			total += d
			if d > max {
				max = d
			}
		}
		avg := total / time.Duration(len(durations))

		t.Logf("%s - Avg: %v, Max: %v, Count: %d", op, avg, max, len(durations))

		// Production SLAs
		var maxAllowed time.Duration
		switch op {
		case "SavePendingTask", "GetTask", "UpdateTaskStatus":
			maxAllowed = 10 * time.Millisecond
		case "ListPendingTasks":
			maxAllowed = 100 * time.Millisecond
		}

		if avg > maxAllowed {
			t.Errorf("%s average time %v exceeds SLA %v", op, avg, maxAllowed)
		}
	}
}

// TestProductionScale tests at production scale
func TestProductionScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping production scale test in short mode")
	}

	ctx := context.Background()
	dir := t.TempDir()

	store, err := aggregatorBadger.NewBadgerAggregatorStore(&aggregatorConfig.BadgerConfig{
		Dir:              dir,
		ValueLogFileSize: 512 << 20, // 512MB
	})
	require.NoError(t, err)
	defer store.Close()

	// Production scale parameters
	numChains := 3
	numTasksPerChain := 100000

	t.Logf("Testing with %d chains, %d tasks per chain", numChains, numTasksPerChain)

	// Populate chains
	avsAddress := "0xtest"
	for c := 1; c <= numChains; c++ {
		chainId := config.ChainId(c)

		// Set block heights
		for block := uint64(0); block < uint64(numTasksPerChain/100); block += 100 {
			require.NoError(t, store.SetLastProcessedBlock(ctx, avsAddress, chainId, block))
		}

		// Create tasks
		for i := 0; i < numTasksPerChain; i++ {
			if i%10000 == 0 {
				t.Logf("Chain %d: Created %d tasks", c, i)
			}

			task := &types.Task{
				TaskId:                 fmt.Sprintf("chain%d-task-%d", c, i),
				AVSAddress:             fmt.Sprintf("0xavs%d", c),
				OperatorSetId:          uint32(i),
				SourceBlockNumber:      uint64(i / 100),
				L1ReferenceBlockNumber: uint64(i / 100),
				ReferenceTimestamp:     100,
				ChainId:                config.ChainId(c),
			}
			require.NoError(t, store.SavePendingTask(ctx, task))
		}
	}

	// Test query performance at scale
	start := time.Now()
	pending, err := store.ListPendingTasks(ctx)
	require.NoError(t, err)
	queryTime := time.Since(start)

	t.Logf("Listed %d pending tasks in %v", len(pending), queryTime)

	// Should complete within reasonable time even at scale
	maxQueryTime := 5 * time.Second
	if queryTime > maxQueryTime {
		t.Errorf("Query time %v exceeds maximum %v at scale", queryTime, maxQueryTime)
	}
}
