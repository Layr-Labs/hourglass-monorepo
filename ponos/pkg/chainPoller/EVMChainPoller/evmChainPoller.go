package EVMChainPoller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"go.uber.org/zap"
)

type EVMChainPollerConfig struct {
	ChainId              config.ChainId
	PollingInterval      time.Duration
	InterestingContracts []string
	AvsAddress           string

	MaxReorgDepth     int
	BlockHistorySize  int
	ReorgCheckEnabled bool
}

type EVMChainPoller struct {
	ethClient           ethereum.Client
	taskQueue           chan *types.Task
	logParser           transactionLogParser.LogParser
	config              *EVMChainPollerConfig
	contractStore       contractStore.IContractStore
	logger              *zap.Logger
	store               storage.AggregatorStore
	blockContextManager IBlockContextManager
}

func NewEVMChainPoller(
	ethClient ethereum.Client,
	taskQueue chan *types.Task,
	logParser transactionLogParser.LogParser,
	config *EVMChainPollerConfig,
	contractStore contractStore.IContractStore,
	store storage.AggregatorStore,
	blockContextManager IBlockContextManager,
	logger *zap.Logger,
) *EVMChainPoller {

	if store == nil {
		panic("store is required")
	}

	// Set default values for reorg configuration if not provided
	if config.MaxReorgDepth == 0 {
		config.MaxReorgDepth = 10
	}
	if config.BlockHistorySize == 0 {
		config.BlockHistorySize = 100
	}
	// ReorgCheckEnabled defaults to true unless explicitly set to false
	if !config.ReorgCheckEnabled && config.MaxReorgDepth > 0 {
		config.ReorgCheckEnabled = true
	}

	for i, contract := range config.InterestingContracts {
		logger.Sugar().Infof("InterestingContracts %d: %s\n", i, contract)
	}
	pollerLogger := logger.With(
		zap.Uint("chainId", uint(config.ChainId)),
	)
	return &EVMChainPoller{
		ethClient:           ethClient,
		logger:              pollerLogger,
		taskQueue:           taskQueue,
		logParser:           logParser,
		config:              config,
		contractStore:       contractStore,
		store:               store,
		blockContextManager: blockContextManager,
	}
}

func (ecp *EVMChainPoller) Start(ctx context.Context) error {

	ecp.logger.Sugar().Infow("Starting Ethereum L1Chain Listener",
		zap.Any("chainId", ecp.config.ChainId),
		zap.Duration("pollingInterval", ecp.config.PollingInterval),
	)

	lastBlockRecord, err := ecp.store.GetLastProcessedBlock(ctx, ecp.config.AvsAddress, ecp.config.ChainId)

	if err != nil {
		ecp.logger.Sugar().Infow("Poller could not get last processed block so using latest block")
		block, err := ecp.ethClient.GetLatestBlock(ctx)
		if err != nil {
			return fmt.Errorf("error getting latest block: %w", err)
		}

		lastCanonBlock, err := ecp.ethClient.GetBlockByNumber(ctx, block)
		if err != nil {
			return fmt.Errorf("couldn't get last canonical block: %w", err)
		}

		lastBlockRecord = &storage.BlockRecord{
			Number:     lastCanonBlock.Number.Value(),
			Hash:       lastCanonBlock.Hash.Value(),
			ParentHash: lastCanonBlock.ParentHash.Value(),
			Timestamp:  lastCanonBlock.Timestamp.Value(),
			ChainId:    ecp.config.ChainId,
		}

		err = ecp.store.SaveBlock(ctx, ecp.config.AvsAddress, lastBlockRecord)
		if err != nil {
			return fmt.Errorf("failed to save last processed block: %w", err)
		}
	}

	if lastBlockRecord == nil {
		return fmt.Errorf("last processed block must exist")
	}

	if err := ecp.recoverInProgressTasks(ctx); err != nil {
		ecp.logger.Sugar().Warnw("Failed to recover in-progress tasks",
			"error", err,
			"avsAddress", ecp.config.AvsAddress)
	}

	go ecp.pollForBlocks(ctx)

	return nil
}

func (ecp *EVMChainPoller) pollForBlocks(ctx context.Context) {

	ecp.logger.Sugar().Infow("Starting Ethereum Chain Listener poll loop")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(ecp.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ecp.logger.Sugar().Infow("Polling loop context cancelled, stopping")
			return
		case <-ticker.C:
			if err := ecp.processNextBlock(ctx); err != nil {
				ecp.logger.Sugar().Errorw("Error processing Ethereum block.", err)
				cancel()
				return
			}
		}
	}
}

func (ecp *EVMChainPoller) processNextBlock(ctx context.Context) error {

	latestBlockRecord, err := ecp.store.GetLastProcessedBlock(ctx, ecp.config.AvsAddress, ecp.config.ChainId)
	if err != nil {
		ecp.logger.Sugar().Errorw("Error getting last processed block", "error", err)
		return nil
	}

	latestBlockNum, err := ecp.ethClient.GetLatestBlock(ctx)
	if err != nil {
		ecp.logger.Sugar().Errorw("Error getting latest block number", "error", err)
		return nil
	}

	if latestBlockRecord.Number == latestBlockNum {
		ecp.logger.Sugar().Debugw("Skipping block processing as the last observed block is the same as the latest block",
			zap.Uint64("lastObservedBlock", latestBlockRecord.Number),
			zap.Uint64("latestBlock", latestBlockNum),
		)
		return nil
	}

	var blocksToFetch []uint64
	if latestBlockNum > latestBlockRecord.Number {
		for i := latestBlockRecord.Number + 1; i <= latestBlockNum; i++ {
			blocksToFetch = append(blocksToFetch, i)
		}
	}

	ecp.logger.Sugar().Debugw("Fetching blocks with logs",
		zap.Any("blocksToFetch", blocksToFetch),
	)

	for _, blockNum := range blocksToFetch {

		newCanonBlock, err := ecp.ethClient.GetBlockByNumber(ctx, blockNum)
		if err != nil {
			ecp.logger.Sugar().Errorw("Failed to fetch block for reorg check",
				zap.Uint64("blockNumber", blockNum),
				zap.Error(err),
			)
			return nil
		}

		if newCanonBlock.ParentHash.Value() != latestBlockRecord.Hash {
			ecp.logger.Sugar().Warnw("Blockchain reorganization detected",
				"blockNumber", blockNum,
				"expectedParent", latestBlockRecord.Hash,
				"actualParent", newCanonBlock.ParentHash.Value(),
				"chainId", ecp.config.ChainId)

			if err = ecp.reconcileReorg(ctx, newCanonBlock); err != nil {
				ecp.logger.Sugar().Errorw("Failed to reconcile reorg", "error", err)
			}
			return nil
		}

		latestBlockRecord, err = ecp.processBlockLogs(ctx, newCanonBlock)
		if err != nil {
			ecp.logger.Sugar().Errorw("Error fetching block with logs",
				zap.Uint64("blockNumber", blockNum),
				zap.Error(err),
			)
			return nil
		}
	}

	ecp.logger.Sugar().Debugw("All blocks processed", zap.Any("blocksToFetch", blocksToFetch))

	if len(blocksToFetch) > 0 && blocksToFetch[len(blocksToFetch)-1]%100 == 0 {
		ecp.logger.Sugar().Infow("Processed block",
			zap.Uint64("blockNumber", blocksToFetch[len(blocksToFetch)-1]),
		)
	}

	return nil
}

func (ecp *EVMChainPoller) processBlockLogs(ctx context.Context, block *ethereum.EthereumBlock) (*storage.BlockRecord, error) {

	logs, err := ecp.fetchLogsForInterestingContractsForBlock(block.Number.Value())
	if err != nil {
		ecp.logger.Sugar().Errorw("Error fetching logs for block",
			zap.Uint64("blockNumber", block.Number.Value()),
			zap.Error(err),
		)
		return nil, err
	}

	block.ChainId = ecp.config.ChainId

	ecp.logger.Sugar().Infow("Block fetched with logs",
		"latestBlockNum", block.Number.Value(),
		"blockHash", block.Hash.Value(),
		"logCount", len(logs),
	)

	for _, log := range logs {

		decodedLog, err := ecp.logParser.DecodeLog(nil, log)
		if err != nil {
			ecp.logger.Sugar().Errorw("Failed to decode log",
				zap.String("transactionHash", log.TransactionHash.Value()),
				zap.String("logAddress", log.Address.Value()),
				zap.Uint64("logIndex", log.LogIndex.Value()),
				zap.Error(err),
			)
			return nil, err
		}

		lwb := &chainPoller.LogWithBlock{
			Block:  block,
			RawLog: log,
			Log:    decodedLog,
		}
		err = ecp.handleLog(ctx, lwb)
		if err != nil {
			return nil, err
		}
	}
	ecp.logger.Sugar().Debugw("Processed logs",
		zap.Uint64("blockNumber", block.Number.Value()),
	)

	blockRecord := &storage.BlockRecord{
		Number:     block.Number.Value(),
		Hash:       block.Hash.Value(),
		ParentHash: block.ParentHash.Value(),
		Timestamp:  block.Timestamp.Value(),
		ChainId:    ecp.config.ChainId,
	}

	if err = ecp.store.SaveBlock(ctx, ecp.config.AvsAddress, blockRecord); err != nil {
		ecp.logger.Sugar().Warnw("Failed to save block info",
			"error", err,
			"blockNumber", blockRecord.Number)
	}

	if ecp.config.BlockHistorySize > 0 && blockRecord.Number > uint64(ecp.config.BlockHistorySize) {
		oldBlockNum := blockRecord.Number - uint64(ecp.config.BlockHistorySize)
		if err := ecp.store.DeleteBlock(ctx, ecp.config.AvsAddress, ecp.config.ChainId, oldBlockNum); err != nil {
			ecp.logger.Sugar().Debugw("Failed to prune old block",
				"blockNumber", oldBlockNum,
				"error", err)
			// TODO: non-fatal for now. Does run the (low) risk of orphaned storage usage growth
		}
	}

	return blockRecord, nil
}

func (ecp *EVMChainPoller) listAllInterestingContracts() []string {

	contracts := make([]string, 0)
	for _, contract := range ecp.config.InterestingContracts {
		if contract != "" {
			contracts = append(contracts, strings.ToLower(contract))
		}
	}
	return contracts
}

func (ecp *EVMChainPoller) fetchLogsForInterestingContractsForBlock(blockNumber uint64) ([]*ethereum.EthereumEventLog, error) {

	var wg sync.WaitGroup

	// TODO: make this configurable in the future
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

			ecp.logger.Sugar().Debugw("Fetching logs for contract",
				zap.String("contract", contract),
				zap.Uint64("blockNumber", blockNumber),
			)

			logs, err := ecp.ethClient.GetLogs(ctxWithTimeout, contract, blockNumber, blockNumber)
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
				ecp.logger.Sugar().Debugw("No logs found for contract",
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

	ecp.logger.Sugar().Debugw("All logs fetched for contracts",
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

// handleLog processes logs from the chain poller
func (ecp *EVMChainPoller) handleLog(ctx context.Context, lwb *chainPoller.LogWithBlock) error {

	ecp.logger.Sugar().Infow("Received log from chain poller",
		zap.Any("log", lwb),
	)
	lg := lwb.Log

	// Handle new task created
	if lg.EventName != "TaskCreated" {
		ecp.logger.Sugar().Debugw("Ignoring log",
			zap.String("eventName", lg.EventName),
			zap.String("contractAddress", lg.Address),
			zap.Strings("addresses", ecp.config.InterestingContracts),
		)
		return nil
	}

	mailboxContract, err := ecp.contractStore.GetContractByNameForChainId(config.ContractName_TaskMailbox, lwb.Block.ChainId)
	if err != nil {
		return err
	}

	if mailboxContract == nil {
		ecp.logger.Sugar().Errorw("Mailbox contract not found for TaskCreated event",
			zap.String("eventName", lg.EventName),
			zap.String("contractAddress", lg.Address),
			zap.Uint("chainId", uint(lwb.Block.ChainId)),
			zap.Uint64("blockNumber", lwb.Block.Number.Value()),
			zap.String("transactionHash", lwb.RawLog.TransactionHash.Value()),
		)
		return nil
	}

	if !strings.EqualFold(lwb.Log.Address, mailboxContract.Address) {
		return nil
	}

	return ecp.processTask(ctx, lwb)
}

func (ecp *EVMChainPoller) processTask(ctx context.Context, lwb *chainPoller.LogWithBlock) error {

	lg := lwb.Log

	ecp.logger.Sugar().Infow("Received TaskCreated event",
		zap.String("eventName", lg.EventName),
		zap.String("contractAddress", lg.Address),
	)
	task, err := types.NewTaskFromLog(lg, lwb.Block, lg.Address)
	if err != nil {
		return fmt.Errorf("failed to convert task: %w", err)
	}

	task.Context = ecp.blockContextManager.GetContext(lwb.Block.Number.Value(), task)

	ecp.logger.Sugar().Infow("Converted task",
		zap.Any("task", task),
	)

	if !strings.EqualFold(task.AVSAddress, ecp.config.AvsAddress) {
		ecp.logger.Sugar().Infow("Ignoring task for different AVS address",
			zap.String("taskAvsAddress", task.AVSAddress),
			zap.String("currentAvsAddress", ecp.config.AvsAddress),
		)
		return nil
	}

	existingTask, err := ecp.store.GetTask(ctx, task.TaskId)
	if err == nil && existingTask != nil {
		ecp.logger.Sugar().Infow("Task already exists in storage, skipping duplicate",
			"taskId", task.TaskId)

		return nil

	} else if err != nil && !errors.Is(err, storage.ErrNotFound) {
		ecp.logger.Sugar().Errorw("Failed to check existing task",
			"error", err,
			"taskId", task.TaskId)
	}

	if err := ecp.store.SavePendingTask(ctx, task); err != nil {
		ecp.logger.Sugar().Errorw("Failed to save task to storage",
			"error", err,
			"taskId", task.TaskId)
		// Continue processing even if initial save fails
	} else {
		ecp.logger.Sugar().Infow("Saved new task to storage as pending",
			"taskId", task.TaskId)
	}

	select {
	case ecp.taskQueue <- task:
		ecp.logger.Sugar().Infow("Task in queue for processing",
			"taskId", task.TaskId)

	case <-time.After(100 * time.Millisecond):
		ecp.logger.Sugar().Warnw("Failed to enqueue task (channel full or closed)",
			zap.String("taskId", task.TaskId),
			zap.Uint64("blockNumber", lwb.Block.Number.Value()),
			zap.String("transactionHash", lwb.RawLog.TransactionHash.Value()),
			zap.String("logAddress", lwb.RawLog.Address.Value()),
			zap.Uint64("logIndex", lwb.RawLog.LogIndex.Value()),
		)

		// Failed to queue, leave status as pending for retry
		return fmt.Errorf("failed to enqueue task (channel full or closed)")
	}

	ecp.logger.Sugar().Infow("Successfully enqueued task for processing",
		zap.String("taskId", task.TaskId),
		zap.Uint64("blockNumber", lwb.Block.Number.Value()),
		zap.String("transactionHash", lwb.RawLog.TransactionHash.Value()),
		zap.String("logAddress", lwb.RawLog.Address.Value()),
		zap.Uint64("logIndex", lwb.RawLog.LogIndex.Value()),
	)

	return nil
}

func (ecp *EVMChainPoller) recoverInProgressTasks(ctx context.Context) error {

	tasks, err := ecp.store.ListPendingTasksForAVS(ctx, ecp.config.AvsAddress)
	if err != nil {
		return fmt.Errorf("failed to list pending tasks for recovery: %w", err)
	}

	if len(tasks) == 0 {
		ecp.logger.Sugar().Debugw("No pending tasks to recover",
			"avsAddress", ecp.config.AvsAddress)
		return nil
	}

	recovered := 0
	expired := 0

	for _, task := range tasks {

		if task.DeadlineUnixSeconds != nil && time.Now().After(*task.DeadlineUnixSeconds) {
			ecp.logger.Sugar().Warnw("Skipping expired task during recovery",
				"taskId", task.TaskId,
				"deadline", task.DeadlineUnixSeconds.Unix())

			if err := ecp.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusFailed); err != nil {
				ecp.logger.Sugar().Warnw("Failed to mark expired task as failed",
					"error", err,
					"taskId", task.TaskId)
			}

			expired++
			continue
		}

		// Try to re-queue the task and mark as processing
		select {
		case ecp.taskQueue <- task:

			if err := ecp.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing); err != nil {
				ecp.logger.Sugar().Warnw("Failed to mark recovered task as processing",
					"error", err,
					"taskId", task.TaskId)
			}
			recovered++
			ecp.logger.Sugar().Infow("Re-queued recovered task",
				"taskId", task.TaskId,
				"avsAddress", ecp.config.AvsAddress)
		case <-time.After(100 * time.Millisecond):
			ecp.logger.Sugar().Warnw("Task queue full, cannot recover task",
				"taskId", task.TaskId)
			// Leave as pending for next recovery attempt
			break
		}
	}

	ecp.logger.Sugar().Infow("Task recovery completed",
		"totalPending", len(tasks),
		"recovered", recovered,
		"expired", expired,
		"avsAddress", ecp.config.AvsAddress)

	return nil
}

// reconcileReorg finds the common ancestor of the previously processed block head and the new blocks
func (ecp *EVMChainPoller) reconcileReorg(ctx context.Context, startBlock *ethereum.EthereumBlock) error {
	orphanedBlocks, err := ecp.findOrphanedBlocks(ctx, startBlock, ecp.config.MaxReorgDepth)

	if err != nil {
		return err
	}

	if len(orphanedBlocks) == 0 {
		return fmt.Errorf("no orphaned blocks found")
	}

	for _, orphanedBlock := range orphanedBlocks {

		ecp.blockContextManager.CancelBlock(orphanedBlock.Number)

		err = ecp.store.DeleteBlock(ctx, ecp.config.AvsAddress, orphanedBlock.ChainId, orphanedBlock.Number)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("failed to delete orphaned block: %w", err)
		}
	}

	return nil
}

func (ecp *EVMChainPoller) findOrphanedBlocks(ctx context.Context, startBlock *ethereum.EthereumBlock, maxDepth int) ([]*storage.BlockRecord, error) {
	var parentBlockRecord *storage.BlockRecord
	var orphanedBlocks []*storage.BlockRecord
	startBlockNumber := startBlock.Number.Value()

	for parentBlockNum := startBlockNumber - 1; startBlockNumber-parentBlockNum <= uint64(maxDepth) && parentBlockNum > 0; parentBlockNum-- {

		canonParentBlock, err := ecp.ethClient.GetBlockByNumber(ctx, parentBlockNum)
		if err != nil || canonParentBlock == nil {
			return nil, fmt.Errorf("failed to fetch block %d from chain: %w", parentBlockNum, err)
		}

		parentBlockRecord, err = ecp.store.GetBlock(
			ctx, ecp.config.AvsAddress,
			ecp.config.ChainId,
			parentBlockNum,
		)

		if err != nil || parentBlockRecord == nil {

			if errors.Is(err, storage.ErrNotFound) {
				ecp.logger.Sugar().Debugw("Block not found in storage",
					"blockNumber", parentBlockNum,
					"error", err)
				parentBlockRecord = &storage.BlockRecord{
					Number:     canonParentBlock.Number.Value(),
					Hash:       canonParentBlock.Hash.Value(),
					ParentHash: canonParentBlock.ParentHash.Value(),
					Timestamp:  canonParentBlock.Timestamp.Value(),
					ChainId:    canonParentBlock.ChainId,
				}
			} else {
				return nil, fmt.Errorf("failed to fetch block %d for : %w", parentBlockNum, err)
			}
		}

		if canonParentBlock.Hash.Value() != parentBlockRecord.Hash {
			ecp.logger.Sugar().Infow("Found orphaned block",
				"blockNumber", parentBlockNum,
				"storedBlockHash", parentBlockRecord.Hash,
				"canonChildBlockHash", canonParentBlock.Hash.Value(),
				"searchDepth", startBlockNumber-parentBlockNum)

			orphanedBlocks = append(orphanedBlocks, parentBlockRecord)
			continue
		}

		ecp.logger.Sugar().Infow("Block hash match, stopping reorg ancestry search",
			"blockNumber", parentBlockNum,
			"storedBlockHash", parentBlockRecord.Hash,
			"canonChildBlockHash", canonParentBlock.Hash.Value())

		return orphanedBlocks, ecp.store.SaveBlock(ctx, ecp.config.AvsAddress, parentBlockRecord)
	}

	ecp.logger.Sugar().Warn("Reached max reorg search depth")

	return orphanedBlocks, ecp.store.SaveBlock(ctx, ecp.config.AvsAddress, parentBlockRecord)
}
