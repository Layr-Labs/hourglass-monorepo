package persistence_test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/badger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockEthereumClient implements the ethereum.Client interface for testing
type mockEthereumClient struct {
	mu                sync.RWMutex
	blocks            map[uint64]*ethereum.EthereumBlock
	latestBlockNumber uint64
	returnErrors      bool
}

// mockLogParser implements a minimal log parser for testing
type mockLogParser struct{}

func (m *mockLogParser) DecodeLog(abi interface{}, log *ethereum.EthereumEventLog) (*types.DecodedLog, error) {
	// Return minimal decoded log for testing
	// Since we're testing reorg detection, not log processing, this is sufficient
	return &types.DecodedLog{
		EventName: "MockEvent",
	}, nil
}

func newMockEthereumClient() *mockEthereumClient {
	return &mockEthereumClient{
		blocks: make(map[uint64]*ethereum.EthereumBlock),
	}
}

func (m *mockEthereumClient) GetLatestBlock(ctx context.Context) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.returnErrors {
		return 0, fmt.Errorf("mock error")
	}
	return m.latestBlockNumber, nil
}

func (m *mockEthereumClient) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*ethereum.EthereumBlock, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.returnErrors {
		return nil, fmt.Errorf("mock error")
	}

	block, exists := m.blocks[blockNumber]
	if !exists {
		return nil, fmt.Errorf("block %d not found", blockNumber)
	}
	return block, nil
}

func (m *mockEthereumClient) GetLogs(ctx context.Context, address string, fromBlock uint64, toBlock uint64) ([]*ethereum.EthereumEventLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.returnErrors {
		return nil, fmt.Errorf("mock error")
	}
	// Return empty logs for this test since we're only testing block storage
	return []*ethereum.EthereumEventLog{}, nil
}

func (m *mockEthereumClient) addBlock(block *ethereum.EthereumBlock) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks[block.Number.Value()] = block
	if block.Number.Value() > m.latestBlockNumber {
		m.latestBlockNumber = block.Number.Value()
	}
}

func (m *mockEthereumClient) setReturnErrors(returnErrors bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.returnErrors = returnErrors
}

// createBlock creates a single ethereum block with the given parameters
func createBlock(number uint64, hash, parentHash string, chainId config.ChainId) *ethereum.EthereumBlock {
	return &ethereum.EthereumBlock{
		Hash:       ethereum.EthereumHexString(hash),
		ParentHash: ethereum.EthereumHexString(parentHash),
		Number:     ethereum.EthereumQuantity(number),
		Timestamp:  ethereum.EthereumQuantity(1000 + number),
		ChainId:    chainId,
	}
}

// createBlockChain creates a chain of blocks with sequential parent-child relationships
func createBlockChain(startNum, endNum uint64, hashSuffix string, chainId config.ChainId) []*ethereum.EthereumBlock {
	blocks := make([]*ethereum.EthereumBlock, 0, endNum-startNum+1)
	
	for i := startNum; i <= endNum; i++ {
		hash := fmt.Sprintf("0xhash%d%s", i, hashSuffix)
		parentHash := fmt.Sprintf("0xhash%d%s", i-1, hashSuffix)
		blocks = append(blocks, createBlock(i, hash, parentHash, chainId))
	}
	
	return blocks
}

// addBlocksToClient adds multiple blocks to the mock ethereum client
func addBlocksToClient(client *mockEthereumClient, blocks []*ethereum.EthereumBlock) {
	for _, block := range blocks {
		client.addBlock(block)
	}
}

// replaceBlocksInClient replaces blocks in the mock client (simulating a reorg)
func replaceBlocksInClient(client *mockEthereumClient, blocks []*ethereum.EthereumBlock) {
	client.mu.Lock()
	defer client.mu.Unlock()
	
	for _, block := range blocks {
		client.blocks[block.Number.Value()] = block
	}
	
	// Update latest block number
	if len(blocks) > 0 {
		lastBlock := blocks[len(blocks)-1]
		if lastBlock.Number.Value() > client.latestBlockNumber {
			client.latestBlockNumber = lastBlock.Number.Value()
		}
	}
}

// verifyBlocksInStore verifies that blocks in the store match expected values
func verifyBlocksInStore(t *testing.T, ctx context.Context, store storage.AggregatorStore, 
	avsAddress string, chainId config.ChainId, startNum, endNum uint64, expectedHashSuffix string) {
	
	for i := startNum; i <= endNum; i++ {
		block, err := store.GetBlock(ctx, avsAddress, chainId, i)
		require.NoError(t, err, "Block %d should exist", i)
		
		expectedHash := fmt.Sprintf("0xhash%d%s", i, expectedHashSuffix)
		assert.Equal(t, expectedHash, block.Hash, "Block %d should have expected hash", i)
	}
}

// TestEVMChainPollerReorgDetection tests basic reorganization detection and handling
func TestEVMChainPollerReorgDetection(t *testing.T) {
	testCases := []struct {
		name         string
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
			avsAddress := "0xavsReorg"
			chainId := config.ChainId(1)

			// Create mock client
			mockClient := newMockEthereumClient()

			// Step 1: Setup simple initial chain (blocks 100-105)
			// Add block 99 as starting point for the poller
			block99 := createBlock(99, "0xhash99", "0xhash98", chainId)
			mockClient.addBlock(block99)
			
			// Create initial blocks (100-105)
			initialBlocks := createBlockChain(100, 105, "", chainId)
			addBlocksToClient(mockClient, initialBlocks)

			// Step 2: Start poller and wait for initial processing
			taskQueue := make(chan *types.Task, 100)
			logger := zap.NewNop()

			pollerConfig := &EVMChainPoller.EVMChainPollerConfig{
				ChainId:              chainId,
				PollingInterval:      50 * time.Millisecond,
				InterestingContracts: []string{},
				AvsAddress:           avsAddress,
				MaxReorgDepth:        10,
				BlockHistorySize:     100,
				ReorgCheckEnabled:    true,
			}

			// Create mock log parser for safer testing
			// (avoids nil pointer issues even though logs are empty)
			logParser := &mockLogParser{}
			
			poller := EVMChainPoller.NewEVMChainPoller(
				mockClient,
				taskQueue,
				logParser,
				pollerConfig,
				nil, // contractStore not needed for this test
				store,
				logger,
			)

			// Start the poller
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				err := poller.Start(ctx)
				if err != nil && ctx.Err() == nil {
					t.Errorf("Poller failed to start: %v", err)
				}
			}()

			// Wait for initial blocks to be processed
			require.Eventually(t, func() bool {
				lastProcessed, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
				return err == nil && lastProcessed >= 105
			}, 5*time.Second, 50*time.Millisecond, "Initial blocks not processed")

			// Verify initial blocks were stored
			verifyBlocksInStore(t, ctx, store, avsAddress, chainId, 100, 105, "")

			// Step 3: Simulate a basic reorg (replace blocks 103-105 with new versions and add 106)
			newBlocks := createBlockChain(103, 106, "_new", chainId)
			// Connect to block 102 (which remains unchanged)
			newBlocks[0].ParentHash = "0xhash102"
			
			// Replace blocks in mock client (simulating the chain reorg)
			replaceBlocksInClient(mockClient, newBlocks)

			// Step 4: Verify reorg was handled correctly
			require.Eventually(t, func() bool {
				lastProcessed, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
				return err == nil && lastProcessed >= 106
			}, 5*time.Second, 50*time.Millisecond, "Reorged blocks not processed")

			// Verify blocks before reorg point (100-102) are unchanged
			verifyBlocksInStore(t, ctx, store, avsAddress, chainId, 100, 102, "")

			// Verify reorged blocks (103-106) have new hashes
			verifyBlocksInStore(t, ctx, store, avsAddress, chainId, 103, 106, "_new")

			// Verify parent-child relationships
			block103, err := store.GetBlock(ctx, avsAddress, chainId, 103)
			require.NoError(t, err)
			assert.Equal(t, "0xhash102", block103.ParentHash,
				"Block 103 should point to block 102")

			// Verify last processed block is updated
			lastProcessed, err := store.GetLastProcessedBlock(ctx, avsAddress, chainId)
			require.NoError(t, err)
			assert.Equal(t, uint64(106), lastProcessed, "Last processed should be 106")
		})
	}
}
