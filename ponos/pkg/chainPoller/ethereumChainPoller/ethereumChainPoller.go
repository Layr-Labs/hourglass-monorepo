package ethereumChainPoller

import (
	"context"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
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
	config *EthereumChainPollerConfig,
	logger *zap.Logger,
) *EthereumChainPoller {
	return &EthereumChainPoller{
		ethClient:  ethClient,
		logger:     logger,
		taskQueue:  taskQueue,
		config:     config,
		errorCount: 0,
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

	sugar.Infow("New Ethereum Block:",
		"blockNum", block.Number.Value(),
		"blockHash", block.Hash.Value(),
		"logCount", len(logs),
	)

	return ecp.processLogs(ctx, block, logs)
}

func (ecp *EthereumChainPoller) processLogs(
	ctx context.Context,
	block *ethereum.EthereumBlock,
	logs []*ethereum.EthereumEventLog,
) bool {
	for range logs {
		task := &types.Task{
			ChainId:      ecp.config.ChainId,
			BlockNumber:  block.Number.Value(),
			BlockHash:    block.Hash.Value(),
			CallbackAddr: ecp.config.InboxAddr,
		}

		select {
		case ecp.taskQueue <- task:
			ecp.logger.Sugar().Debugw("Enqueued task for processing",
				"blockNumber", task.BlockNumber,
				"blockHash", task.BlockHash,
			)
		case <-time.After(100 * time.Millisecond):
			ecp.logger.Sugar().Warnw("Failed to enqueue task (channel full or closed)",
				"blockNumber", task.BlockNumber,
				"blockHash", task.BlockHash,
			)
		case <-ctx.Done():
			return false
		}
	}
	return true
}

func (ecp *EthereumChainPoller) getNextBlockWithLogs(ctx context.Context) (*ethereum.EthereumBlock, []*ethereum.EthereumEventLog, error) {
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
