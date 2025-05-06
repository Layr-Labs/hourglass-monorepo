package EVMChainPoller

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"go.uber.org/zap"
)

type EVMChainPollerConfig struct {
	ChainId                 config.ChainId
	PollingInterval         time.Duration
	EigenLayerCoreContracts []string
	InterestingContracts    []string
}

type EVMChainPoller struct {
	ethClient         *ethereum.Client
	lastObservedBlock *ethereum.EthereumBlock
	chainEventsChan   chan *chainPoller.LogWithBlock
	logParser         *transactionLogParser.TransactionLogParser
	config            *EVMChainPollerConfig
	logger            *zap.Logger
}

func NewEVMChainPollerDefaultConfig(chainId config.ChainId, inboxAddr string) *EVMChainPollerConfig {
	return &EVMChainPollerConfig{
		ChainId:         chainId,
		PollingInterval: 10 * time.Millisecond,
	}
}

func NewEVMChainPoller(
	ethClient *ethereum.Client,
	chainEventsChan chan *chainPoller.LogWithBlock,
	logParser *transactionLogParser.TransactionLogParser,
	config *EVMChainPollerConfig,
	logger *zap.Logger,
) *EVMChainPoller {
	for i, contract := range config.EigenLayerCoreContracts {
		fmt.Printf("Contract %d: %s\n", i, contract)
	}
	return &EVMChainPoller{
		ethClient:       ethClient,
		logger:          logger,
		chainEventsChan: chainEventsChan,
		logParser:       logParser,
		config:          config,
	}
}

func (ecp *EVMChainPoller) Start(ctx context.Context) error {
	sugar := ecp.logger.Sugar()
	sugar.Infow("Starting Ethereum Chain Listener",
		zap.Any("chainId", ecp.config.ChainId),
		zap.Duration("pollingInterval", ecp.config.PollingInterval),
	)
	go ecp.pollForBlocks(ctx)
	return nil
}

func (ecp *EVMChainPoller) pollForBlocks(ctx context.Context) {
	ecp.logger.Sugar().Infow("Starting Ethereum Chain Listener poll loop")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shouldStop := atomic.Bool{}

	go func() {
		for !shouldStop.Load() {
			ecp.logger.Sugar().Infow("Tick")
			err := ecp.processNextBlock(ctx)
			if err != nil {
				ecp.logger.Sugar().Errorw("Error processing Ethereum block.", err)
				cancel()
				return
			}
			time.Sleep(ecp.config.PollingInterval)
		}
	}()

	<-ctx.Done()
	shouldStop.Store(true)
	ecp.logger.Sugar().Infow("Ethereum Chain Listener context cancelled, exiting poll loop")
}

func (ecp *EVMChainPoller) isInterestingLog(log *ethereum.EthereumEventLog) bool {
	logAddr := strings.ToLower(log.Address.Value())
	if slices.Contains(ecp.config.InterestingContracts, logAddr) {
		return true
	}
	if config.IsL1Chain(ecp.config.ChainId) && slices.Contains(ecp.config.EigenLayerCoreContracts, logAddr) {
		return true
	}
	return false
}

func (ecp *EVMChainPoller) processNextBlock(ctx context.Context) error {
	latestBlockNum, err := ecp.ethClient.GetLatestBlock(ctx)
	if err != nil {
		return nil
	}

	if ecp.lastObservedBlock == nil {
		ecp.logger.Sugar().Infow("no lastObservedBlock set, initializing last observed block to latest - 1")
		ecp.lastObservedBlock = &ethereum.EthereumBlock{
			Number: ethereum.EthereumQuantity(latestBlockNum - 1),
		}
	} else {
		ecp.logger.Sugar().Infow("latest on chain block",
			zap.Uint64("blockNumber", latestBlockNum),
			zap.Uint64("lastObservedBlock", ecp.lastObservedBlock.Number.Value()),
		)
	}

	// if the latest observed block is the same as the latest block, skip processing
	if ecp.lastObservedBlock != nil && ecp.lastObservedBlock.Number.Value() == latestBlockNum {
		ecp.logger.Sugar().Infow("Skipping block processing as the last observed block is the same as the latest block",
			zap.Uint64("lastObservedBlock", ecp.lastObservedBlock.Number.Value()),
			zap.Uint64("latestBlock", latestBlockNum),
		)
		return nil
	}

	// if the latest observed block is greater than the latest block, skip processing since the chain is lagging behind
	if ecp.lastObservedBlock != nil && ecp.lastObservedBlock.Number.Value() > latestBlockNum {
		ecp.logger.Sugar().Infow("Skipping block processing as the last observed block is greater than the latest block",
			zap.Uint64("lastObservedBlock", ecp.lastObservedBlock.Number.Value()),
			zap.Uint64("latestBlock", latestBlockNum),
		)
		return nil
	}

	var blocksToFetch []uint64
	if latestBlockNum >= ecp.lastObservedBlock.Number.Value()+1 {
		for i := ecp.lastObservedBlock.Number.Value() + 1; i <= latestBlockNum; i++ {
			blocksToFetch = append(blocksToFetch, i)
		}
	}
	ecp.logger.Sugar().Infow("Fetching blocks with logs",
		zap.Any("blocksToFetch", blocksToFetch),
	)

	for _, blockNum := range blocksToFetch {
		_, _, err = ecp.getBlockWithLogs(ctx, blockNum)
		if err != nil {
			ecp.logger.Sugar().Errorw("Error fetching block with logs",
				zap.Uint64("blockNumber", blockNum),
				zap.Error(err),
			)
			return err
		}
	}
	ecp.logger.Sugar().Infow("All blocks processed",
		zap.Any("blocksToFetch", blocksToFetch),
	)

	return nil
}

func (ecp *EVMChainPoller) getBlockWithLogs(ctx context.Context, blockNum uint64) (*ethereum.EthereumBlock, []*ethereum.EthereumEventLog, error) {
	ecp.logger.Sugar().Infow("Fetching Ethereum block with logs",
		zap.Uint64("blockNumber", blockNum),
	)
	block, err := ecp.ethClient.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		return nil, nil, err
	}

	logs, err := ecp.fetchLogsForInterestingContractsForBlock(block.Number.Value())
	if err != nil {
		ecp.logger.Sugar().Errorw("Error fetching logs for block",
			zap.Uint64("blockNumber", block.Number.Value()),
			zap.Error(err),
		)
		return nil, nil, err
	}

	if block == nil {
		return nil, nil, nil
	}
	block.ChainId = ecp.config.ChainId

	ecp.logger.Sugar().Infow("Block fetched with logs",
		"latestBlockNum", block.Number.Value(),
		"blockHash", block.Hash.Value(),
		"logCount", len(logs),
	)

	for _, l := range logs {
		if !ecp.isInterestingLog(l) {
			continue
		}

		decodedLog, err := ecp.logParser.DecodeLog(nil, l)
		if err != nil {
			ecp.logger.Sugar().Errorw("Failed to decode log",
				zap.String("transactionHash", l.TransactionHash.Value()),
				zap.String("logAddress", l.Address.Value()),
				zap.Uint64("logIndex", l.LogIndex.Value()),
				zap.Error(err),
			)
			return nil, nil, err
		}

		lwb := &chainPoller.LogWithBlock{
			Block: block,
			Log:   decodedLog,
		}
		select {
		case ecp.chainEventsChan <- lwb:
			ecp.logger.Sugar().Infow("Enqueued log for processing",
				zap.Uint64("blockNumber", block.Number.Value()),
				zap.String("transactionHash", l.TransactionHash.Value()),
				zap.String("logAddress", l.Address.Value()),
				zap.Uint64("logIndex", l.LogIndex.Value()),
			)
		case <-time.After(100 * time.Millisecond):
			ecp.logger.Sugar().Warnw("Failed to enqueue log (channel full or closed)",
				zap.Uint64("blockNumber", block.Number.Value()),
				zap.String("transactionHash", l.TransactionHash.Value()),
				zap.String("logAddress", l.Address.Value()),
				zap.Uint64("logIndex", l.LogIndex.Value()),
			)
		}
	}
	ecp.logger.Sugar().Infow("Processed logs",
		zap.Uint64("blockNumber", block.Number.Value()),
	)
	ecp.lastObservedBlock = block
	return block, logs, nil
}

func (ecp *EVMChainPoller) listAllInterestingContracts() []string {
	contracts := make([]string, 0)
	for _, contract := range ecp.config.InterestingContracts {
		if contract != "" {
			contracts = append(contracts, strings.ToLower(contract))
		}
	}
	for _, contract := range ecp.config.EigenLayerCoreContracts {
		if contract != "" {
			contracts = append(contracts, strings.ToLower(contract))
		}
	}
	return contracts
}

func (ecp *EVMChainPoller) fetchLogsForInterestingContractsForBlock(blockNumber uint64) ([]*ethereum.EthereumEventLog, error) {
	var wg sync.WaitGroup

	allContracts := ecp.listAllInterestingContracts()
	ecp.logger.Sugar().Infow("Fetching logs for interesting contracts",
		zap.Any("contracts", allContracts),
	)
	logResultsChan := make(chan []*ethereum.EthereumEventLog, len(allContracts))
	errorsChan := make(chan error, len(allContracts))

	for _, contract := range allContracts {
		wg.Add(1)
		go func(contract string, wg *sync.WaitGroup) {
			defer wg.Done()

			ecp.logger.Sugar().Infow("Fetching logs for contract",
				zap.String("contract", contract),
				zap.Uint64("blockNumber", blockNumber),
			)
			logs, err := ecp.ethClient.GetLogs(context.Background(), contract, blockNumber, blockNumber)
			if err != nil {
				ecp.logger.Sugar().Errorw("Failed to fetch logs for contract",
					zap.String("contract", contract),
					zap.Uint64("blockNumber", blockNumber),
					zap.Error(err),
				)
				errorsChan <- fmt.Errorf("failed to fetch logs for contract %s: %w", contract, err)
				return
			}
			if len(logs) == 0 {
				ecp.logger.Sugar().Infow("No logs found for contract",
					zap.String("contract", contract),
					zap.Uint64("blockNumber", blockNumber),
				)
				logResultsChan <- []*ethereum.EthereumEventLog{}
				return
			}
			ecp.logger.Sugar().Infow("Fetched logs for contract",
				zap.String("contract", contract),
				zap.Uint64("blockNumber", blockNumber),
				zap.Int("logCount", len(logs)),
			)
			logResultsChan <- logs
		}(contract, &wg)
	}
	wg.Wait()
	close(logResultsChan)
	close(errorsChan)
	ecp.logger.Sugar().Infow("All logs fetched for contracts",
		zap.Uint64("blockNumber", blockNumber),
	)

	allErrors := make([]error, 0)
	for err := range errorsChan {
		allErrors = append(allErrors, err)
	}
	if len(allErrors) > 0 {
		return nil, fmt.Errorf("failed to fetch logs for contracts: %v", allErrors)
	}

	allLogs := make([]*ethereum.EthereumEventLog, 0)
	for contractLogs := range logResultsChan {
		allLogs = append(allLogs, contractLogs...)
	}
	ecp.logger.Sugar().Infow("All logs fetched for contracts",
		zap.Uint64("blockNumber", blockNumber),
		zap.Int("logCount", len(allLogs)),
	)

	return allLogs, nil
}
