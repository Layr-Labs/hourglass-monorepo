package EVMChainPoller

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"go.uber.org/zap"
)

type EVNChainPollerConfig struct {
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
	config            *EVNChainPollerConfig
	logger            *zap.Logger
}

func NewEVMChainPollerDefaultConfig(chainId config.ChainId, inboxAddr string) *EVNChainPollerConfig {
	return &EVNChainPollerConfig{
		ChainId:         chainId,
		PollingInterval: 10 * time.Millisecond,
	}
}

func NewEVMChainPoller(
	ethClient *ethereum.Client,
	chainEventsChan chan *chainPoller.LogWithBlock,
	logParser *transactionLogParser.TransactionLogParser,
	config *EVNChainPollerConfig,
	logger *zap.Logger,
) *EVMChainPoller {
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
	ctx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(ecp.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ecp.logger.Sugar().Infow("Ethereum Chain Listener context cancelled, exiting poll loop")
			cancel()
			return
		case <-ticker.C:
			err := ecp.processNextBlock(ctx)
			if err != nil {
				ecp.logger.Sugar().Errorw("Error processing Ethereum block.", err)
				cancel()
				return
			}
		}
	}
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
	block, logs, err := ecp.getNextBlockWithLogs(ctx)
	if err != nil {
		return err
	}

	if block == nil {
		return nil
	}
	block.ChainId = ecp.config.ChainId

	ecp.logger.Sugar().Infow("New Ethereum Block:",
		"blockNum", block.Number.Value(),
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
			return err
		}

		lwb := &chainPoller.LogWithBlock{
			Block: block,
			Log:   decodedLog,
		}
		select {
		case ecp.chainEventsChan <- lwb:
			ecp.logger.Sugar().Infow("Enqueued l for processing",
				zap.Uint64("blockNumber", block.Number.Value()),
				zap.String("transactionHash", l.TransactionHash.Value()),
				zap.String("logAddress", l.Address.Value()),
				zap.Uint64("logIndex", l.LogIndex.Value()),
			)
		case <-time.After(100 * time.Millisecond):
			ecp.logger.Sugar().Warnw("Failed to enqueue l (channel full or closed)",
				zap.Uint64("blockNumber", block.Number.Value()),
				zap.String("transactionHash", l.TransactionHash.Value()),
				zap.String("logAddress", l.Address.Value()),
				zap.Uint64("logIndex", l.LogIndex.Value()),
			)
		}
	}
	return nil
}

func (ecp *EVMChainPoller) getNextBlockWithLogs(ctx context.Context) (*ethereum.EthereumBlock, []*ethereum.EthereumEventLog, error) {
	blockNum, err := ecp.ethClient.GetLatestBlock(ctx)
	if err != nil {
		return nil, nil, err
	}

	// if the latest observed block is the same as the latest block, skip processing
	if ecp.lastObservedBlock != nil && ecp.lastObservedBlock.Number.Value() == blockNum {
		ecp.logger.Sugar().Infow("Skipping block processing as the last observed block is the same as the latest block",
			zap.Uint64("lastObservedBlock", ecp.lastObservedBlock.Number.Value()),
			zap.Uint64("latestBlock", blockNum),
		)
		return nil, nil, nil
	}

	// if the latest observed block is greater than the latest block, skip processing since the chain is lagging behind
	if ecp.lastObservedBlock != nil && ecp.lastObservedBlock.Number.Value() > blockNum {
		ecp.logger.Sugar().Infow("Skipping block processing as the last observed block is greater than the latest block",
			zap.Uint64("lastObservedBlock", ecp.lastObservedBlock.Number.Value()),
			zap.Uint64("latestBlock", blockNum),
		)
		return nil, nil, nil
	}

	// TODO: if blockNum > latestBlock + 1, we need to handle filling in the gap.
	block, err := ecp.ethClient.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		return nil, nil, err
	}
	ecp.lastObservedBlock = block

	logs, err := ecp.fetchLogsForInterestingContractsForBlock(block.Number.Value())
	if err != nil {
		return nil, nil, err
	}
	return block, logs, nil
}

func (ecp *EVMChainPoller) listAllInterestingContracts() []string {
	contracts := make([]string, 0)
	for _, contract := range ecp.config.InterestingContracts {
		contracts = append(contracts, strings.ToLower(contract))
	}
	for _, contract := range ecp.config.EigenLayerCoreContracts {
		contracts = append(contracts, strings.ToLower(contract))
	}
	return contracts
}

func (ecp *EVMChainPoller) fetchLogsForInterestingContractsForBlock(blockNumber uint64) ([]*ethereum.EthereumEventLog, error) {
	var wg sync.WaitGroup

	allContracts := ecp.listAllInterestingContracts()
	logResultsChan := make(chan []*ethereum.EthereumEventLog, len(allContracts))
	errorsChan := make(chan error, len(allContracts))

	for _, contract := range allContracts {
		wg.Add(1)
		go func(contract string, wg *sync.WaitGroup) {
			defer wg.Done()

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

	return allLogs, nil
}
