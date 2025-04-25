package ethereumChainListener

import (
	"context"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

type EthereumChainListenerConfig struct {
	ChainId                  config.ChainId
	PollingInterval          time.Duration
	InboxAddr                string
	MaxConsecutiveErrorCount int
}

type EthereumChainListener struct {
	ethClient         *ethereum.Client
	lastObservedBlock *ethereum.EthereumBlock
	logger            *zap.Logger
	taskQueue         chan *types.Task
	errorCount        int
	config            *EthereumChainListenerConfig
}

func NewEthereumChainListenerDefaultConfig(
	chainId config.ChainId,
	inboxAddr string,
) *EthereumChainListenerConfig {
	return &EthereumChainListenerConfig{
		ChainId:                  chainId,
		InboxAddr:                inboxAddr,
		PollingInterval:          10 * time.Millisecond,
		MaxConsecutiveErrorCount: 5,
	}
}

func NewEthereumChainListener(
	ethClient *ethereum.Client,
	logger *zap.Logger,
	taskQueue chan *types.Task,
	config *EthereumChainListenerConfig,
) *EthereumChainListener {
	return &EthereumChainListener{
		ethClient:  ethClient,
		logger:     logger,
		taskQueue:  taskQueue,
		config:     config,
		errorCount: 0,
	}
}

func (ecl *EthereumChainListener) Start(ctx context.Context) error {
	sugar := ecl.logger.Sugar()
	sugar.Infow("Starting Ethereum Chain Listener",
		"chainId", ecl.config.ChainId,
		"inboxAddr", ecl.config.InboxAddr,
		"pollingInterval", ecl.config.PollingInterval,
	)
	go ecl.pollForBlocks(ctx)
	return nil
}

func (ecl *EthereumChainListener) pollForBlocks(ctx context.Context) {
	ticker := time.NewTicker(ecl.config.PollingInterval)
	defer ticker.Stop()

	sugar := ecl.logger.Sugar()

	for {
		select {
		case <-ctx.Done():
			sugar.Infow("Ethereum Chain Listener context cancelled, exiting poll loop")
			return
		case <-ticker.C:
			shouldContinue := ecl.processNextBlock(ctx)

			if !shouldContinue {
				return
			}
		}
	}
}

func (ecl *EthereumChainListener) processNextBlock(ctx context.Context) bool {
	sugar := ecl.logger.Sugar()

	block, logs, err := ecl.getNextBlockWithLogs(ctx)
	if err != nil {
		sugar.Errorw("Failed to get next block", "error", err)
		ecl.errorCount++
		if ecl.errorCount > ecl.config.MaxConsecutiveErrorCount {
			sugar.Errorw("Too many consecutive errors, stopping poll loop",
				"errorCount", ecl.errorCount,
				"maxErrorCount", ecl.config.MaxConsecutiveErrorCount,
			)
			return false
		}
		return true
	}

	ecl.errorCount = 0
	if block == nil {
		return true
	}

	sugar.Infow("New Ethereum Block:",
		"blockNum", block.Number.Value(),
		"blockHash", block.Hash.Value(),
		"logCount", len(logs),
	)

	return ecl.processLogs(ctx, block, logs)
}

func (ecl *EthereumChainListener) processLogs(
	ctx context.Context,
	block *ethereum.EthereumBlock,
	logs []*ethereum.EthereumEventLog,
) bool {
	for range logs {
		task := &types.Task{
			ChainId:      ecl.config.ChainId,
			BlockNumber:  block.Number.Value(),
			BlockHash:    block.Hash.Value(),
			CallbackAddr: ecl.config.InboxAddr,
		}

		select {
		case ecl.taskQueue <- task:
			ecl.logger.Sugar().Debugw("Enqueued task for processing",
				"blockNumber", task.BlockNumber,
				"blockHash", task.BlockHash,
			)
		case <-time.After(100 * time.Millisecond):
			ecl.logger.Sugar().Warnw("Failed to enqueue task (channel full or closed)",
				"blockNumber", task.BlockNumber,
				"blockHash", task.BlockHash,
			)
		case <-ctx.Done():
			return false
		}
	}
	return true
}

func (ecl *EthereumChainListener) getNextBlockWithLogs(ctx context.Context) (*ethereum.EthereumBlock, []*ethereum.EthereumEventLog, error) {
	blockNum, err := ecl.ethClient.GetLatestBlock(ctx)
	if err != nil {
		return nil, nil, err
	}

	if ecl.lastObservedBlock != nil && ecl.lastObservedBlock.Number.Value() == blockNum {
		return nil, nil, nil
	}

	block, err := ecl.ethClient.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		return nil, nil, err
	}
	ecl.lastObservedBlock = block

	logs, err := ecl.ethClient.GetLogs(ctx, ecl.config.InboxAddr, block.Number.Value(), block.Number.Value())
	if err != nil {
		return nil, nil, err
	}
	return block, logs, nil
}
