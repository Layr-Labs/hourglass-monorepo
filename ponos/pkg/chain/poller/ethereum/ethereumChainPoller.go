package ethereum

import (
	"context"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"go.uber.org/zap"
)

type EthereumChainPollerConfig struct {
	ChainId                  config.ChainId
	PollingInterval          time.Duration
	InboxAddr                string
	MaxConsecutiveErrorCount int
}

type EthereumChainPoller struct {
	ethClient         *ethereum.Client
	lastObservedBlock *ethereum.EthereumBlock
	taskQueue         chan *types.Task
	logParser         *transactionLogParser.TransactionLogParser
	contractABI       *abi.ABI
	config            *EthereumChainPollerConfig
	logger            *zap.Logger
	errorCount        int
}

func NewEthereumChainPollerDefaultConfig(
	chainId config.ChainId,
	inboxAddr string,
) *EthereumChainPollerConfig {
	return &EthereumChainPollerConfig{
		ChainId:                  chainId,
		InboxAddr:                inboxAddr,
		PollingInterval:          10 * time.Millisecond,
		MaxConsecutiveErrorCount: 5,
	}
}

func NewEthereumChainPoller(
	ethClient *ethereum.Client,
	taskQueue chan *types.Task,
	logParser *transactionLogParser.TransactionLogParser,
	lastObservedBlock *ethereum.EthereumBlock,
	abi *abi.ABI,
	config *EthereumChainPollerConfig,
	logger *zap.Logger,
) *EthereumChainPoller {
	return &EthereumChainPoller{
		ethClient:         ethClient,
		taskQueue:         taskQueue,
		logParser:         logParser,
		lastObservedBlock: lastObservedBlock,
		contractABI:       abi,
		config:            config,
		logger:            logger,
		errorCount:        0,
	}
}

func (ecp *EthereumChainPoller) Start(ctx context.Context) error {
	sugar := ecp.logger.Sugar()
	sugar.Infow("Starting Ethereum Chain Listener",
		"chainId", ecp.config.ChainId,
		"inboxAddr", ecp.config.InboxAddr,
		"pollingInterval", ecp.config.PollingInterval,
	)
	go ecp.pollForBlocks(ctx)
	return nil
}

func (ecp *EthereumChainPoller) pollForBlocks(ctx context.Context) {
	ticker := time.NewTicker(ecp.config.PollingInterval)
	defer ticker.Stop()

	sugar := ecp.logger.Sugar()

	for {
		select {
		case <-ctx.Done():
			sugar.Infow("Ethereum Chain Listener context cancelled, exiting poll loop")
			return
		case <-ticker.C:
			shouldContinue := ecp.processNextBlock(ctx)

			if !shouldContinue {
				return
			}
		}
	}
}

func (ecp *EthereumChainPoller) processNextBlock(ctx context.Context) bool {
	sugar := ecp.logger.Sugar()

	block, logs, err := ecp.getNextBlockWithLogs(ctx)
	if err != nil {
		sugar.Errorw("Failed to get next block", "error", err)
		ecp.errorCount++
		if ecp.errorCount > ecp.config.MaxConsecutiveErrorCount {
			sugar.Errorw("Too many consecutive errors, stopping poll loop",
				"errorCount", ecp.errorCount,
				"maxErrorCount", ecp.config.MaxConsecutiveErrorCount,
			)
			return false
		}
		return true
	}

	ecp.errorCount = 0
	if block == nil {
		return true
	}

	return ecp.processLogs(ctx, block, logs)
}

func (ecp *EthereumChainPoller) processLogs(
	ctx context.Context,
	block *ethereum.EthereumBlock,
	logs []*ethereum.EthereumEventLog,
) bool {
	sugar := ecp.logger.Sugar()

	for _, log := range logs {
		decodedLog, err := ecp.logParser.DecodeLog(ecp.contractABI, log)
		if err != nil {
			sugar.Errorw("Failed to parse transaction logs",
				"txHash", log.TransactionHash.Value(),
				"error", err,
			)
			continue
		}

		task, err := ecp.convertDecodedLogToTask(decodedLog, block)
		if err != nil {
			sugar.Errorw("Failed to convert decoded log to task",
				"txHash", log.TransactionHash.Value(),
				"eventName", decodedLog.EventName,
				"error", err,
			)
			continue
		}

		select {
		case ecp.taskQueue <- task:
			sugar.Infow("Enqueued task for processing",
				"blockNumber", task.BlockNumber,
				"blockHash", task.BlockHash,
				"taskId", task.TaskId,
			)
		case <-time.After(100 * time.Millisecond):
			sugar.Warnw("Failed to enqueue task (channel full or closed)",
				"blockNumber", task.BlockNumber,
				"blockHash", task.BlockHash,
				"taskId", task.TaskId,
			)
		case <-ctx.Done():
			return false
		}
	}
	return true
}

func (ecp *EthereumChainPoller) convertDecodedLogToTask(
	decodedLog *transactionLogParser.DecodedLog,
	block *ethereum.EthereumBlock,
) (*types.Task, error) {
	if decodedLog.EventName != "TaskCreated" {
		return nil, nil
	}

	return types.NewTask(decodedLog, block, ecp.config.InboxAddr, ecp.config.ChainId)
}

func (ecp *EthereumChainPoller) getNextBlockWithLogs(
	ctx context.Context,
) (*ethereum.EthereumBlock, []*ethereum.EthereumEventLog, error) {
	blockNum, err := ecp.ethClient.GetLatestBlock(ctx)
	if err != nil {
		return nil, nil, err
	}

	if ecp.lastObservedBlock != nil && ecp.lastObservedBlock.Number.Value() == blockNum {
		return nil, nil, nil
	}

	block, err := ecp.ethClient.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		return nil, nil, err
	}
	ecp.lastObservedBlock = block

	logs, err := ecp.ethClient.GetLogs(ctx, ecp.config.InboxAddr, block.Number.Value(), block.Number.Value())
	if err != nil {
		return nil, nil, err
	}
	return block, logs, nil
}
