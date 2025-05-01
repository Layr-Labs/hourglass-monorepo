package ethereumChainPoller

import (
	"context"
	"fmt"
	"math/big"
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

type EthereumChainPollerConfig struct {
	ChainId         config.ChainId
	PollingInterval time.Duration
	InboxAddr       string
}

type EthereumChainPoller struct {
	ethClient         *ethereum.Client
	lastObservedBlock *ethereum.EthereumBlock
	taskQueue         chan *types.Task
	logParser         *transactionLogParser.TransactionLogParser
	config            *EthereumChainPollerConfig
	logger            *zap.Logger
}

func NewEthereumChainPollerDefaultConfig(chainId config.ChainId, inboxAddr string) *EthereumChainPollerConfig {
	return &EthereumChainPollerConfig{
		ChainId:         chainId,
		InboxAddr:       inboxAddr,
		PollingInterval: 10 * time.Millisecond,
	}
}

func NewEthereumChainPoller(
	ethClient *ethereum.Client,
	taskQueue chan *types.Task,
	logParser *transactionLogParser.TransactionLogParser,
	config *EthereumChainPollerConfig,
	logger *zap.Logger,
) *EthereumChainPoller {
	return &EthereumChainPoller{
		ethClient: ethClient,
		logger:    logger,
		taskQueue: taskQueue,
		logParser: logParser,
		config:    config,
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

func (ecp *EthereumChainPoller) processNextBlock(ctx context.Context) error {
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

	return ecp.processLogs(ctx, block, logs)
}

func (ecp *EthereumChainPoller) processLogs(
	ctx context.Context,
	block *ethereum.EthereumBlock,
	logs []*ethereum.EthereumEventLog,
) error {
	for _, log := range logs {
		if !strings.EqualFold(log.Address.Value(), ecp.config.InboxAddr) {
			continue
		}

		parsedLog, err := ecp.logParser.ParseLog(log)
		if err != nil {
			// TODO: emit metric
			return fmt.Errorf("failed to parse log: %w", err)
		}
		task, err := convertTask(parsedLog, block, ecp.config.InboxAddr)
		if err != nil {
			// TODO: emit metric
			return fmt.Errorf("failed to convert task: %w", err)
		}

		select {
		case ecp.taskQueue <- task:
			ecp.logger.Sugar().Infow("Enqueued task for processing",
				"blockNumber", task.BlockNumber,
				"blockHash", task.BlockHash,
			)
		case <-time.After(100 * time.Millisecond):
			ecp.logger.Sugar().Warnw("Failed to enqueue task (channel full or closed)",
				"blockNumber", task.BlockNumber,
				"blockHash", task.BlockHash,
			)
		case <-ctx.Done():
			return nil
		}
	}
	return nil
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
