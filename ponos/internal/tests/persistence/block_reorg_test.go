package persistence_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockEthereumClient struct {
	mu                sync.RWMutex
	blocks            map[uint64]*ethereum.EthereumBlock
	latestBlockNumber uint64
	returnErrors      bool
	logs              map[string]map[blockRange][]*ethereum.EthereumEventLog
}

type blockRange struct {
	from, to uint64
}

type mockContractStore struct {
	contracts map[string]*contracts.Contract
}

func newMockContractStore(mailboxAddress string, chainId config.ChainId) *mockContractStore {

	taskMailboxABI := `[{"abi": string}]`
	mailboxContract := &contracts.Contract{
		Name:        "TaskMailbox",
		Address:     mailboxAddress,
		AbiVersions: []string{taskMailboxABI},
		ChainId:     chainId,
	}

	return &mockContractStore{
		contracts: map[string]*contracts.Contract{
			strings.ToLower(mailboxAddress): mailboxContract,
			"TaskMailbox":                   mailboxContract,
		},
	}
}

func (m *mockContractStore) GetContractByAddress(address string) (*contracts.Contract, error) {
	contract, exists := m.contracts[strings.ToLower(address)]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", address)
	}
	return contract, nil
}

func (m *mockContractStore) GetContractByNameForChainId(name string, chainId config.ChainId) (*contracts.Contract, error) {
	contract, exists := m.contracts[name]
	if !exists || contract.ChainId != chainId {
		return nil, fmt.Errorf("contract not found: %s on chain %d", name, chainId)
	}
	return contract, nil
}

func (m *mockContractStore) ListContractAddressesForChain(chainId config.ChainId) []string {
	var addresses []string
	for addr, contract := range m.contracts {
		if contract.ChainId == chainId && addr != "TaskMailbox" {
			addresses = append(addresses, addr)
		}
	}
	return addresses
}

func (m *mockContractStore) ListContracts() []*contracts.Contract {
	var result []*contracts.Contract
	for _, contract := range m.contracts {
		result = append(result, contract)
	}
	return result
}

func (m *mockContractStore) OverrideContract(contractName string, chainIds []config.ChainId, contract *contracts.Contract) error {
	m.contracts[contractName] = contract
	return nil
}

func newMockEthereumClient() *mockEthereumClient {
	return &mockEthereumClient{
		blocks: make(map[uint64]*ethereum.EthereumBlock),
		logs:   make(map[string]map[blockRange][]*ethereum.EthereumEventLog),
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

	var result []*ethereum.EthereumEventLog

	if addressLogs, exists := m.logs[strings.ToLower(address)]; exists {
		for br, logs := range addressLogs {
			if br.from <= toBlock && br.to >= fromBlock {
				for _, log := range logs {
					if log.BlockNumber.Value() >= fromBlock && log.BlockNumber.Value() <= toBlock {
						result = append(result, log)
					}
				}
			}
		}
	}

	return result, nil
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

// addLogs adds event logs for a specific address and block range
func (m *mockEthereumClient) addLogs(address string, fromBlock, toBlock uint64, logs []*ethereum.EthereumEventLog) {
	m.mu.Lock()
	defer m.mu.Unlock()

	addr := strings.ToLower(address)
	if m.logs[addr] == nil {
		m.logs[addr] = make(map[blockRange][]*ethereum.EthereumEventLog)
	}

	br := blockRange{from: fromBlock, to: toBlock}
	m.logs[addr][br] = logs
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
		{
			name: "InMemory",
			storeFactory: func(t *testing.T) storage.AggregatorStore {
				return memory.NewInMemoryAggregatorStore()
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
			block99 := createBlock(99, "0xhash99", "0xhash98", chainId)
			mockClient.addBlock(block99)

			// Create initial blocks (100-105)
			initialBlocks := createBlockChain(100, 105, "", chainId)
			addBlocksToClient(mockClient, initialBlocks)

			// Initialize the last processed block to 99 so poller starts from block 100
			err := store.SetLastProcessedBlock(ctx, avsAddress, chainId, 99)
			require.NoError(t, err)
			blockInfo := &storage.BlockInfo{
				Number:     99,
				Hash:       "0xhash99",
				ParentHash: "0xhash98",
				Timestamp:  10000,
				ChainId:    chainId,
			}
			err = store.SaveBlock(ctx, avsAddress, blockInfo)
			require.NoError(t, err)

			// Step 2: Start poller and wait for initial processing
			taskQueue := make(chan *types.Task, 100)
			logger := zap.NewNop()

			// Setup mailbox contract address
			mailboxAddress := "0xmailbox123"

			pollerConfig := &EVMChainPoller.EVMChainPollerConfig{
				ChainId:              chainId,
				PollingInterval:      50 * time.Millisecond,
				InterestingContracts: []string{mailboxAddress},
				AvsAddress:           avsAddress,
				MaxReorgDepth:        10,
				BlockHistorySize:     100,
				ReorgCheckEnabled:    true,
			}

			// Create real TransactionLogParser with mock contract store
			contractStore := newMockContractStore(mailboxAddress, chainId)
			logParser := transactionLogParser.NewTransactionLogParser(contractStore, logger)

			poller := EVMChainPoller.NewEVMChainPoller(
				mockClient,
				taskQueue,
				logParser,
				pollerConfig,
				contractStore,
				store,
				logger,
			)

			// Start the poller
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			err = poller.Start(ctx)
			if err != nil && ctx.Err() == nil {
				t.Errorf("Poller failed to start: %v", err)
			}

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

			// Step 5: Verify tasks in queue is empty since there is no log provided in events
			select {
			case task := <-taskQueue:
				t.Errorf("Unexpected task in queue: %v", task)
			case <-time.After(100 * time.Millisecond):
				t.Log("No tasks in queue as expected (no logs added)")
			}
		})
	}
}
