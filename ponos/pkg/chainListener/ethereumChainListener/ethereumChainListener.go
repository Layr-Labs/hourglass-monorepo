package ethereumChainListener

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"time"
)

type EthereumChainListener struct {
	ethClient         *ethereum.Client
	logger            *zap.Logger
	lastObservedBlock *ethereum.EthereumBlock
}

func NewEthereumChainListener(ethClient *ethereum.Client, logger *zap.Logger) *EthereumChainListener {
	return &EthereumChainListener{
		ethClient: ethClient,
		logger:    logger,
	}
}

func (ecl *EthereumChainListener) ListenForInboxEvents(
	ctx context.Context,
	queue chan *chainListener.Event,
	inboxAddress string,
) error {
	ecl.logger.Info("Starting Ethereum Chain Listener")
	// Implement the logic to start listening to Ethereum chain events

	// TODO(seanmcgary): actually pass the inbox address
	err := ecl.pollForBlocks(ctx, queue, []string{})
	return err
}

func (ecl *EthereumChainListener) pollForBlocks(
	ctx context.Context,
	queue chan *chainListener.Event,
	interestingAddresses []string,
) error {
	done := atomic.NewBool(false)

	go func() {
		errorCount := 0
		for {
			time.Sleep(1 * time.Second)
			if done.Load() {
				return
			}
			// if we've encountered 5 errors in a row, something is wrong
			if errorCount > 5 {
				ecl.logger.Sugar().Error("Too many errors, stopping Ethereum Chain Listener")
				return
			}
			block, logs, err := ecl.getNextBlockWithLogs(ctx, interestingAddresses)
			if err != nil {
				ecl.logger.Sugar().Errorw("Failed to get next block",
					zap.Error(err),
				)
				errorCount++
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
		}
	}()
	<-ctx.Done()
	ecl.logger.Sugar().Infow("Context done, stopping Ethereum Chain Listener")
	done.Store(true)
	return nil
}

func (ecl *EthereumChainListener) getNextBlockWithLogs(
	ctx context.Context,
	interestingAddresses []string,
) (*ethereum.EthereumBlock, []*ethereum.EthereumEventLog, error) {
	blockNum, err := ecl.ethClient.GetLatestBlock(ctx)
	if err != nil {
		ecl.logger.Sugar().Errorw("Failed to get latest block",
			zap.Error(err),
		)
		return nil, nil, err
	}
	if ecl.lastObservedBlock != nil && ecl.lastObservedBlock.Number.Value() == blockNum {
		ecl.logger.Sugar().Debugw("No new block found")
		return nil, nil, nil
	}

	block, err := ecl.ethClient.GetBlockByNumber(ctx, blockNum)
	if err != nil {
		ecl.logger.Sugar().Errorw("Failed to get block by number",
			zap.Uint64("blockNum", blockNum),
			zap.Error(err),
		)
		return nil, nil, err
	}
	ecl.lastObservedBlock = block

	if len(interestingAddresses) == 0 {
		ecl.logger.Sugar().Debugw("No interesting addresses, returning block",
			zap.Uint64("blockNum", block.Number.Value()),
		)
		return block, nil, nil
	}
	logs, err := ecl.ethClient.GetLogs(ctx, "", block.Number.Value(), block.Number.Value())
	if err != nil {
		ecl.logger.Sugar().Errorw("Failed to get logs",
			zap.Uint64("blockNum", block.Number.Value()),
			zap.Error(err),
		)
		return nil, nil, err
	}
	return block, logs, err
}
