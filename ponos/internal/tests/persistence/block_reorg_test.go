package persistence_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/badger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBlockInfoBasicOperations tests basic save, get, delete operations
func TestBlockInfoBasicOperations(t *testing.T) {
	testCases := []struct {
		name        string
		storeFactory func(t *testing.T) storage.AggregatorStore
	}{
		{
			name: "InMemory",
			storeFactory: func(t *testing.T) storage.AggregatorStore {
				return memory.NewInMemoryAggregatorStore()
			},
		},
		{
			name: "Badger",
			storeFactory: func(t *testing.T) storage.AggregatorStore {
				tmpDir := t.TempDir()
				cfg := &aggregatorConfig.BadgerConfig{
					Dir:      filepath.Join(tmpDir, "badger"),
					InMemory: false,
				}
				store, err := badger.NewBadgerAggregatorStore(cfg)
				require.NoError(t, err)
				t.Cleanup(func() { store.Close() })
				return store
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := tc.storeFactory(t)
			avsAddress := "0xavs123"
			chainId := config.ChainId(1)

			// Test Save and Get
			block1 := &storage.BlockInfo{
				Number:     100,
				Hash:       "0xhash100",
				ParentHash: "0xhash99",
				Timestamp:  1234567890,
				ChainId:    chainId,
			}

			// Save block
			err := store.SaveBlock(ctx, avsAddress, block1)
			require.NoError(t, err)

			// Retrieve block
			retrievedBlock, err := store.GetBlock(ctx, avsAddress, chainId, 100)
			require.NoError(t, err)
			assert.Equal(t, block1.Hash, retrievedBlock.Hash)
			assert.Equal(t, block1.ParentHash, retrievedBlock.ParentHash)
			assert.Equal(t, block1.Number, retrievedBlock.Number)

			// Test Delete
			err = store.DeleteBlock(ctx, avsAddress, chainId, 100)
			require.NoError(t, err)

			// Verify block is deleted
			_, err = store.GetBlock(ctx, avsAddress, chainId, 100)
			assert.ErrorIs(t, err, storage.ErrNotFound)
		})
	}
}

// TestBlockchainSequence tests saving and retrieving a sequence of blocks
func TestBlockchainSequence(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	cfg := &aggregatorConfig.BadgerConfig{
		Dir:      filepath.Join(tmpDir, "badger"),
		InMemory: false,
	}
	store, err := badger.NewBadgerAggregatorStore(cfg)
	require.NoError(t, err)
	defer store.Close()

	avsAddress := "0xavs456"
	chainId := config.ChainId(1)

	// Create a chain of blocks
	blocks := []storage.BlockInfo{
		{Number: 100, Hash: "0xhash100", ParentHash: "0xhash99", Timestamp: 1000, ChainId: chainId},
		{Number: 101, Hash: "0xhash101", ParentHash: "0xhash100", Timestamp: 1001, ChainId: chainId},
		{Number: 102, Hash: "0xhash102", ParentHash: "0xhash101", Timestamp: 1002, ChainId: chainId},
		{Number: 103, Hash: "0xhash103", ParentHash: "0xhash102", Timestamp: 1003, ChainId: chainId},
		{Number: 104, Hash: "0xhash104", ParentHash: "0xhash103", Timestamp: 1004, ChainId: chainId},
	}

	// Save all blocks
	for i := range blocks {
		err := store.SaveBlock(ctx, avsAddress, &blocks[i])
		require.NoError(t, err)
	}

	// Verify all blocks can be retrieved
	for _, expectedBlock := range blocks {
		retrievedBlock, err := store.GetBlock(ctx, avsAddress, chainId, expectedBlock.Number)
		require.NoError(t, err)
		assert.Equal(t, expectedBlock.Hash, retrievedBlock.Hash)
		assert.Equal(t, expectedBlock.ParentHash, retrievedBlock.ParentHash)
	}

	// Verify parent-child relationships
	for i := 1; i < len(blocks); i++ {
		currentBlock, err := store.GetBlock(ctx, avsAddress, chainId, blocks[i].Number)
		require.NoError(t, err)
		
		previousBlock, err := store.GetBlock(ctx, avsAddress, chainId, blocks[i-1].Number)
		require.NoError(t, err)
		
		// Current block's parent hash should match previous block's hash
		assert.Equal(t, previousBlock.Hash, currentBlock.ParentHash)
	}
}

// TestReorganizationScenario simulates a blockchain reorganization
func TestReorganizationScenario(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	cfg := &aggregatorConfig.BadgerConfig{
		Dir:      filepath.Join(tmpDir, "badger"),
		InMemory: false,
	}
	store, err := badger.NewBadgerAggregatorStore(cfg)
	require.NoError(t, err)
	defer store.Close()

	avsAddress := "0xavs789"
	chainId := config.ChainId(1)

	// Create initial chain
	originalBlocks := []storage.BlockInfo{
		{Number: 100, Hash: "0xhash100", ParentHash: "0xhash99", Timestamp: 1000, ChainId: chainId},
		{Number: 101, Hash: "0xhash101", ParentHash: "0xhash100", Timestamp: 1001, ChainId: chainId},
		{Number: 102, Hash: "0xhash102", ParentHash: "0xhash101", Timestamp: 1002, ChainId: chainId},
		{Number: 103, Hash: "0xhash103_old", ParentHash: "0xhash102", Timestamp: 1003, ChainId: chainId},
		{Number: 104, Hash: "0xhash104_old", ParentHash: "0xhash103_old", Timestamp: 1004, ChainId: chainId},
		{Number: 105, Hash: "0xhash105_old", ParentHash: "0xhash104_old", Timestamp: 1005, ChainId: chainId},
	}

	// Save original chain
	for i := range originalBlocks {
		err := store.SaveBlock(ctx, avsAddress, &originalBlocks[i])
		require.NoError(t, err)
	}

	// Set last processed block
	err = store.SetLastProcessedBlock(ctx, avsAddress, chainId, 105)
	require.NoError(t, err)

	// Simulate reorg: blocks 103-105 are replaced
	// Delete old blocks
	for blockNum := uint64(103); blockNum <= 105; blockNum++ {
		err := store.DeleteBlock(ctx, avsAddress, chainId, blockNum)
		require.NoError(t, err)
	}

	// Add new blocks (reorged chain)
	reorgedBlocks := []storage.BlockInfo{
		{Number: 103, Hash: "0xhash103_new", ParentHash: "0xhash102", Timestamp: 1003, ChainId: chainId},
		{Number: 104, Hash: "0xhash104_new", ParentHash: "0xhash103_new", Timestamp: 1004, ChainId: chainId},
		{Number: 105, Hash: "0xhash105_new", ParentHash: "0xhash104_new", Timestamp: 1005, ChainId: chainId},
		{Number: 106, Hash: "0xhash106_new", ParentHash: "0xhash105_new", Timestamp: 1006, ChainId: chainId},
	}

	for i := range reorgedBlocks {
		err := store.SaveBlock(ctx, avsAddress, &reorgedBlocks[i])
		require.NoError(t, err)
	}

	// Update last processed block
	err = store.SetLastProcessedBlock(ctx, avsAddress, chainId, 106)
	require.NoError(t, err)

	// Verify blocks 100-102 are unchanged
	for i := 0; i < 3; i++ {
		block, err := store.GetBlock(ctx, avsAddress, chainId, originalBlocks[i].Number)
		require.NoError(t, err)
		assert.Equal(t, originalBlocks[i].Hash, block.Hash)
	}

	// Verify blocks 103-106 are the new versions
	for _, expectedBlock := range reorgedBlocks {
		block, err := store.GetBlock(ctx, avsAddress, chainId, expectedBlock.Number)
		require.NoError(t, err)
		assert.Equal(t, expectedBlock.Hash, block.Hash)
		assert.Equal(t, expectedBlock.ParentHash, block.ParentHash)
	}

	// Verify last processed block is updated
	lastBlock, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
	require.NoError(t, err)
	assert.Equal(t, uint64(106), lastBlock)
}

// TestBlockPruning tests that old blocks can be pruned
func TestBlockPruning(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	cfg := &aggregatorConfig.BadgerConfig{
		Dir:      filepath.Join(tmpDir, "badger"),
		InMemory: false,
	}
	store, err := badger.NewBadgerAggregatorStore(cfg)
	require.NoError(t, err)
	defer store.Close()

	avsAddress := "0xavsABC"
	chainId := config.ChainId(1)
	retentionLimit := 10 // Keep only last 10 blocks

	// Save many blocks
	for i := uint64(1); i <= 20; i++ {
		block := &storage.BlockInfo{
			Number:     i,
			Hash:       fmt.Sprintf("0xhash%d", i),
			ParentHash: fmt.Sprintf("0xhash%d", i-1),
			Timestamp:  1000 + i,
			ChainId:    chainId,
		}
		err := store.SaveBlock(ctx, avsAddress, block)
		require.NoError(t, err)

		// Prune old blocks if we exceed retention limit
		if i > uint64(retentionLimit) {
			oldBlockNum := i - uint64(retentionLimit)
			err := store.DeleteBlock(ctx, avsAddress, chainId, oldBlockNum)
			// Ignore error as block might not exist
			_ = err
		}
	}

	// Verify old blocks are pruned
	for i := uint64(1); i <= 10; i++ {
		_, err := store.GetBlock(ctx, avsAddress, chainId, i)
		assert.ErrorIs(t, err, storage.ErrNotFound, "Block %d should be pruned", i)
	}

	// Verify recent blocks are retained
	for i := uint64(11); i <= 20; i++ {
		block, err := store.GetBlock(ctx, avsAddress, chainId, i)
		require.NoError(t, err, "Block %d should be retained", i)
		assert.Equal(t, fmt.Sprintf("0xhash%d", i), block.Hash)
	}
}

// TestBlockPersistenceAcrossRestarts tests that blocks persist when store is closed and reopened
func TestBlockPersistenceAcrossRestarts(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "badger")
	
	avsAddress := "0xavsDEF"
	chainId := config.ChainId(1)

	// Phase 1: Save blocks and close store
	{
		cfg := &aggregatorConfig.BadgerConfig{
			Dir:      dbPath,
			InMemory: false,
		}
		store, err := badger.NewBadgerAggregatorStore(cfg)
		require.NoError(t, err)

		// Save some blocks
		blocks := []storage.BlockInfo{
			{Number: 200, Hash: "0xhash200", ParentHash: "0xhash199", Timestamp: 2000, ChainId: chainId},
			{Number: 201, Hash: "0xhash201", ParentHash: "0xhash200", Timestamp: 2001, ChainId: chainId},
			{Number: 202, Hash: "0xhash202", ParentHash: "0xhash201", Timestamp: 2002, ChainId: chainId},
		}

		for i := range blocks {
			err := store.SaveBlock(ctx, avsAddress, &blocks[i])
			require.NoError(t, err)
		}

		// Set last processed block
		err = store.SetLastProcessedBlock(ctx, avsAddress, chainId, 202)
		require.NoError(t, err)

		// Close the store
		err = store.Close()
		require.NoError(t, err)
	}

	// Phase 2: Reopen store and verify persistence
	{
		cfg := &aggregatorConfig.BadgerConfig{
			Dir:      dbPath,
			InMemory: false,
		}
		store, err := badger.NewBadgerAggregatorStore(cfg)
		require.NoError(t, err)
		defer store.Close()

		// Verify blocks are still there
		expectedBlocks := map[uint64]string{
			200: "0xhash200",
			201: "0xhash201",
			202: "0xhash202",
		}

		for blockNum, expectedHash := range expectedBlocks {
			block, err := store.GetBlock(ctx, avsAddress, chainId, blockNum)
			require.NoError(t, err, "Block %d should persist", blockNum)
			assert.Equal(t, expectedHash, block.Hash)
		}

		// Verify last processed block persists
		lastBlock, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
		require.NoError(t, err)
		assert.Equal(t, uint64(202), lastBlock)

		// Add more blocks to verify store is still functional
		newBlock := &storage.BlockInfo{
			Number:     203,
			Hash:       "0xhash203",
			ParentHash: "0xhash202",
			Timestamp:  2003,
			ChainId:    chainId,
		}
		err = store.SaveBlock(ctx, avsAddress, newBlock)
		require.NoError(t, err)

		// Verify new block was saved
		retrievedBlock, err := store.GetBlock(ctx, avsAddress, chainId, 203)
		require.NoError(t, err)
		assert.Equal(t, "0xhash203", retrievedBlock.Hash)
	}
}

// TestMultipleAVSIsolation tests that blocks from different AVS addresses are isolated
func TestMultipleAVSIsolation(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	cfg := &aggregatorConfig.BadgerConfig{
		Dir:      filepath.Join(tmpDir, "badger"),
		InMemory: false,
	}
	store, err := badger.NewBadgerAggregatorStore(cfg)
	require.NoError(t, err)
	defer store.Close()

	avs1 := "0xavs111"
	avs2 := "0xavs222"
	chainId := config.ChainId(1)

	// Save blocks for AVS1
	block1AVS1 := &storage.BlockInfo{
		Number:     100,
		Hash:       "0xavs1_hash100",
		ParentHash: "0xavs1_hash99",
		Timestamp:  1000,
		ChainId:    chainId,
	}
	err = store.SaveBlock(ctx, avs1, block1AVS1)
	require.NoError(t, err)

	// Save blocks for AVS2 with same block number
	block1AVS2 := &storage.BlockInfo{
		Number:     100,
		Hash:       "0xavs2_hash100",
		ParentHash: "0xavs2_hash99",
		Timestamp:  2000,
		ChainId:    chainId,
	}
	err = store.SaveBlock(ctx, avs2, block1AVS2)
	require.NoError(t, err)

	// Verify isolation - each AVS gets its own block
	retrievedAVS1, err := store.GetBlock(ctx, avs1, chainId, 100)
	require.NoError(t, err)
	assert.Equal(t, "0xavs1_hash100", retrievedAVS1.Hash)

	retrievedAVS2, err := store.GetBlock(ctx, avs2, chainId, 100)
	require.NoError(t, err)
	assert.Equal(t, "0xavs2_hash100", retrievedAVS2.Hash)

	// Delete block for AVS1 shouldn't affect AVS2
	err = store.DeleteBlock(ctx, avs1, chainId, 100)
	require.NoError(t, err)

	// AVS1 block should be gone
	_, err = store.GetBlock(ctx, avs1, chainId, 100)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// AVS2 block should still exist
	retrievedAVS2Again, err := store.GetBlock(ctx, avs2, chainId, 100)
	require.NoError(t, err)
	assert.Equal(t, "0xavs2_hash100", retrievedAVS2Again.Hash)
}

// Helper function for creating test blocks
func createTestBlock(number uint64, chainId config.ChainId) *storage.BlockInfo {
	return &storage.BlockInfo{
		Number:     number,
		Hash:       fmt.Sprintf("0xhash%d", number),
		ParentHash: fmt.Sprintf("0xhash%d", number-1),
		Timestamp:  1000000 + number,
		ChainId:    chainId,
	}
}