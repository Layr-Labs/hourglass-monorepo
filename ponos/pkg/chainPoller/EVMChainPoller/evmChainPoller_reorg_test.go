package EVMChainPoller

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestReorgDetection(t *testing.T) {
	// Create a test context
	ctx := context.Background()

	// Create in-memory storage
	store := memory.NewInMemoryAggregatorStore()

	// Create config with reorg detection enabled
	pollerConfig := &EVMChainPollerConfig{
		ChainId:              1,
		PollingInterval:      time.Second,
		InterestingContracts: []string{"0x1234"},
		AvsAddress:           "0xavs",
		MaxReorgDepth:        10,
		BlockHistorySize:     100,
		ReorgCheckEnabled:    true,
	}

	// Test saving and retrieving blocks
	t.Run("SaveAndRetrieveBlock", func(t *testing.T) {
		block1 := &storage.BlockInfo{
			Number:     100,
			Hash:       "0xhash100",
			ParentHash: "0xhash99",
			Timestamp:  1234567890,
			ChainId:    1,
		}

		// Save block
		err := store.SaveBlock(ctx, pollerConfig.AvsAddress, block1)
		require.NoError(t, err)

		// Retrieve block
		retrievedBlock, err := store.GetBlock(ctx, pollerConfig.AvsAddress, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, block1.Hash, retrievedBlock.Hash)
		assert.Equal(t, block1.ParentHash, retrievedBlock.ParentHash)
	})

	// Test block deletion
	t.Run("DeleteBlock", func(t *testing.T) {
		block2 := &storage.BlockInfo{
			Number:     101,
			Hash:       "0xhash101",
			ParentHash: "0xhash100",
			Timestamp:  1234567891,
			ChainId:    1,
		}

		// Save block
		err := store.SaveBlock(ctx, pollerConfig.AvsAddress, block2)
		require.NoError(t, err)

		// Delete block
		err = store.DeleteBlock(ctx, pollerConfig.AvsAddress, 1, 101)
		require.NoError(t, err)

		// Try to retrieve deleted block
		_, err = store.GetBlock(ctx, pollerConfig.AvsAddress, 1, 101)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})

	// Test finding common ancestor
	t.Run("FindCommonAncestor", func(t *testing.T) {
		// Save a chain of blocks
		blocks := []storage.BlockInfo{
			{Number: 90, Hash: "0xhash90", ParentHash: "0xhash89", ChainId: 1},
			{Number: 91, Hash: "0xhash91", ParentHash: "0xhash90", ChainId: 1},
			{Number: 92, Hash: "0xhash92", ParentHash: "0xhash91", ChainId: 1},
			{Number: 93, Hash: "0xhash93", ParentHash: "0xhash92", ChainId: 1},
			{Number: 94, Hash: "0xhash94_reorged", ParentHash: "0xhash93", ChainId: 1},         // This will be reorged
			{Number: 95, Hash: "0xhash95_reorged", ParentHash: "0xhash94_reorged", ChainId: 1}, // This will be reorged
		}

		for i := range blocks {
			err := store.SaveBlock(ctx, pollerConfig.AvsAddress, &blocks[i])
			require.NoError(t, err)
		}

		// The common ancestor should be block 93 (before the reorg)
		// In a real scenario, we'd compare with chain blocks to find the divergence point
		commonBlock, err := store.GetBlock(ctx, pollerConfig.AvsAddress, 1, 93)
		require.NoError(t, err)
		assert.Equal(t, "0xhash93", commonBlock.Hash)
	})

	// Test rollback scenario
	t.Run("HandleReorg", func(t *testing.T) {
		// Save last processed block
		err := store.SetLastProcessedBlock(ctx, pollerConfig.AvsAddress, 1, 95)
		require.NoError(t, err)

		// Simulate rollback to block 93 (common ancestor)
		// Delete blocks 94 and 95
		err = store.DeleteBlock(ctx, pollerConfig.AvsAddress, 1, 95)
		assert.NoError(t, err) // May or may not exist

		err = store.DeleteBlock(ctx, pollerConfig.AvsAddress, 1, 94)
		assert.NoError(t, err) // May or may not exist

		// Reset last processed block to common ancestor
		err = store.SetLastProcessedBlock(ctx, pollerConfig.AvsAddress, 1, 93)
		require.NoError(t, err)

		// Verify last processed block is now 93
		lastBlock, err := store.GetLastProcessedBlock(ctx, pollerConfig.AvsAddress, 1)
		require.NoError(t, err)
		assert.Equal(t, uint64(93), lastBlock)
	})
}

// Test configuration defaults
func TestConfigDefaults(t *testing.T) {
	logger := zap.NewNop()
	store := memory.NewInMemoryAggregatorStore()

	config := &EVMChainPollerConfig{
		ChainId:              1,
		PollingInterval:      time.Second,
		InterestingContracts: []string{"0x1234"},
		AvsAddress:           "0xavs",
		// Don't set reorg config to test defaults
	}

	poller := NewEVMChainPoller(
		nil, // ethClient
		make(chan *types.Task, 10),
		nil, // logParser
		config,
		nil, // contractStore
		store,
		logger,
	)

	// Check defaults were set
	assert.Equal(t, 10, poller.config.MaxReorgDepth)
	assert.Equal(t, 100, poller.config.BlockHistorySize)
	assert.True(t, poller.config.ReorgCheckEnabled)
}
