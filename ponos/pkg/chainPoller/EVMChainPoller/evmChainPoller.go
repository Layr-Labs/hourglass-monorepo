package EVMChainPoller

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type EVNChainPollerConfig struct {
	ChainId                 config.ChainId
	PollingInterval         time.Duration
	InboxAddr               string
	EigenLayerCoreContracts []string
	InterestingContracts    []string
}

type EVMChainPoller struct {
	ethClient         *ethereum.Client
	lastObservedBlock *ethereum.EthereumBlock
	logSequencer      chan *chainPoller.LogWithBlock
	logParser         *transactionLogParser.TransactionLogParser
	config            *EVNChainPollerConfig
	logger            *zap.Logger
}

func NewEVMChainPollerDefaultConfig(chainId config.ChainId, inboxAddr string) *EVNChainPollerConfig {
	return &EVNChainPollerConfig{
		ChainId:         chainId,
		InboxAddr:       inboxAddr,
		PollingInterval: 10 * time.Millisecond,
	}
}

func NewEVMChainPoller(
	ethClient *ethereum.Client,
	sequencer chan *chainPoller.LogWithBlock,
	logParser *transactionLogParser.TransactionLogParser,
	config *EVNChainPollerConfig,
	logger *zap.Logger,
) *EVMChainPoller {
	return &EVMChainPoller{
		ethClient:    ethClient,
		logger:       logger,
		logSequencer: sequencer,
		logParser:    logParser,
		config:       config,
	}
}

func (ecp *EVMChainPoller) Start(ctx context.Context) error {
	sugar := ecp.logger.Sugar()
	sugar.Infow("Starting Ethereum Chain Listener",
		"chainId", ecp.config.ChainId,
		"inboxAddr", ecp.config.InboxAddr,
		"pollingInterval", ecp.config.PollingInterval,
	)
	go ecp.pollForBlocks(ctx)
	return nil
}

func (ecp *EVMChainPoller) pollForBlocks(ctx context.Context) {
	ticker := time.NewTicker(ecp.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ecp.logger.Sugar().Infow("Ethereum Chain Listener context cancelled, exiting poll loop")
			return
		case <-ticker.C:
			err := ecp.processNextBlock(ctx)
			if err != nil {
				ecp.logger.Sugar().Errorw("Error processing Ethereum block.", err)
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

	for _, log := range logs {
		if !ecp.isInterestingLog(log) {
			continue
		}

		lwb := &chainPoller.LogWithBlock{
			Block: block,
			Log:   log,
		}
		select {
		case ecp.logSequencer <- lwb:
			ecp.logger.Sugar().Infow("Enqueued log for processing",
				zap.Uint64("blockNumber", block.Number.Value()),
				zap.String("transactionHash", log.TransactionHash.Value()),
				zap.String("logAddress", log.Address.Value()),
				zap.Uint64("logIndex", log.LogIndex.Value()),
			)
		case <-time.After(100 * time.Millisecond):
			ecp.logger.Sugar().Warnw("Failed to enqueue log (channel full or closed)",
				zap.Uint64("blockNumber", block.Number.Value()),
				zap.String("transactionHash", log.TransactionHash.Value()),
				zap.String("logAddress", log.Address.Value()),
				zap.Uint64("logIndex", log.LogIndex.Value()),
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

	logs, err := ecp.ethClient.GetLogs(ctx, ecp.config.InboxAddr, block.Number.Value(), block.Number.Value())
	if err != nil {
		return nil, nil, err
	}
	return block, logs, nil
}

func convertTask(log *log.DecodedLog, block *ethereum.EthereumBlock, inboxAddress string) (*types.Task, error) {
	var avsAddress common.Address
	var operatorSetId uint32
	var parsedTaskDeadline *big.Int
	var taskId string
	var payload []byte

	taskId, ok := log.Arguments[1].Value.(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse task id")
	}
	avsAddress, ok = log.Arguments[2].Value.(common.Address)
	if !ok {
		return nil, fmt.Errorf("failed to parse task event address")
	}
	operatorSetId, ok = log.OutputData["executorOperatorSetId"].(uint32)
	if !ok {
		return nil, fmt.Errorf("failed to parse operator set id")
	}
	parsedTaskDeadline, ok = log.OutputData["taskDeadline"].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to parse task event deadline")
	}
	taskDeadlineTime := time.Now().Add(time.Duration(parsedTaskDeadline.Int64()) * time.Second)
	payload, ok = log.OutputData["payload"].([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to parse task payload")
	}

	return &types.Task{
		TaskId:              taskId,
		AVSAddress:          avsAddress.String(),
		OperatorSetId:       operatorSetId,
		CallbackAddr:        inboxAddress,
		DeadlineUnixSeconds: &taskDeadlineTime,
		Payload:             payload,
		ChainId:             block.ChainId,
		BlockNumber:         block.Number.Value(),
		BlockHash:           block.Hash.Value(),
	}, nil
}
