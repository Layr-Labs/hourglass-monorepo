package ethereumChainListener

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/workQueue"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"go.uber.org/zap"
)

type EthereumChainListener struct {
	ethClient         *ethereum.Client
	logger            *zap.Logger
	lastObservedBlock *ethereum.EthereumBlock
	ctx               context.Context
	cancel            context.CancelFunc
	queue             workQueue.IInputQueue[types.Task]
	inboxAddr         string
	chainId           config.ChainId
}

func NewEthereumChainListener(
	ethClient *ethereum.Client,
	logger *zap.Logger,
	queue workQueue.IInputQueue[types.Task],
	inboxAddr string,
	chainId config.ChainId,
) *EthereumChainListener {
	return &EthereumChainListener{
		ethClient: ethClient,
		logger:    logger,
		queue:     queue,
		inboxAddr: inboxAddr,
		chainId:   chainId,
	}
}

func (ecl *EthereumChainListener) Start(ctx context.Context) error {
	ecl.logger.Info("Starting Ethereum Chain Listener")
	ecl.ctx, ecl.cancel = context.WithCancel(ctx)

	go ecl.pollForBlocks()
	return nil
}

func (ecl *EthereumChainListener) Close() error {
	ecl.logger.Info("Stopping Ethereum Chain Listener")
	if ecl.cancel != nil {
		ecl.cancel()
	}
	return nil
}

func (ecl *EthereumChainListener) pollForBlocks() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	errorCount := 0
	for {
		select {
		case <-ecl.ctx.Done():
			ecl.logger.Info("Ethereum Chain Listener context cancelled, exiting poll loop")
			return
		case <-ticker.C:
			block, logs, err := ecl.getNextBlockWithLogs(ecl.ctx)
			if err != nil {
				ecl.logger.Sugar().Errorw("Failed to get next block", zap.Error(err))
				errorCount++
				if errorCount > 5 {
					ecl.logger.Error("Too many errors, stopping poll loop")
					return
				}
				continue
			}
			errorCount = 0

			if block == nil {
				continue
			}

			ecl.logger.Sugar().Infow("Got new block",
				zap.Uint64("blockNum", block.Number.Value()),
				zap.String("blockHash", block.Hash.Value()),
				zap.Int("logCount", len(logs)),
			)

			for range logs {
				task := &types.Task{
					ChainId:      ecl.chainId,
					BlockNumber:  block.Number.Value(),
					BlockHash:    block.Hash.Value(),
					CallbackAddr: ecl.inboxAddr,
				}
				_ = ecl.queue.Enqueue(task)
			}
		}
	}
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

	logs, err := ecl.ethClient.GetLogs(ctx, ecl.inboxAddr, block.Number.Value(), block.Number.Value())
	if err != nil {
		return nil, nil, err
	}
	return block, logs, nil
}
