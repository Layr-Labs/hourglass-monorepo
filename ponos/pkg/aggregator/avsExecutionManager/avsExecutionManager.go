package avsExecutionManager

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/taskSession"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
	"strings"
	"sync"
)

type AggregationStrategy uint

const (
	AggregationStrategyNone          AggregationStrategy = 0
	AggregationStrategyStakeWeighted AggregationStrategy = 1
)

type AvsExecutionManagerConfig struct {
	AvsAddress               string
	SupportedChainIds        []config.ChainId
	MailboxContractAddresses map[config.ChainId]string
	AggregatorAddress        string
	AggregatorUrl            string
	L1ChainId                config.ChainId
	AggregationStrategy      AggregationStrategy
}

type AvsExecutionManager struct {
	logger *zap.Logger
	config *AvsExecutionManagerConfig

	// will be a proper type when another PR is merged
	chainContractCallers map[config.ChainId]contractCaller.IContractCaller

	signer signer.ISigner

	operatorManager *operatorManager.OperatorManager

	contractStore contractStore.IContractStore

	taskQueue chan *types.Task

	inflightTasks sync.Map

	aggregationStrategy IAggregationStrategy
}

func NewAvsExecutionManager(
	config *AvsExecutionManagerConfig,
	chainContractCallers map[config.ChainId]contractCaller.IContractCaller,
	signer signer.ISigner,
	cs contractStore.IContractStore,
	om *operatorManager.OperatorManager,
	logger *zap.Logger,
) (*AvsExecutionManager, error) {
	logger.Sugar().Infow("Creating AvsExecutionManager",
		zap.String("avsAddress", config.AvsAddress),
		zap.Any("supportedChainIds", config.SupportedChainIds),
	)
	if config.L1ChainId == 0 {
		return nil, fmt.Errorf("L1ChainId must be set in AvsExecutionManagerConfig")
	}
	if err := hasExpectedMailboxContractsForChains(config.SupportedChainIds, config.MailboxContractAddresses); err != nil {
		return nil, fmt.Errorf("invalid mailbox contract addresses: %w", err)
	}
	if err := hasExpectedContractCallersForChains(config.SupportedChainIds, chainContractCallers); err != nil {
		return nil, fmt.Errorf("invalid contract callers: %w", err)
	}
	if _, ok := chainContractCallers[config.L1ChainId]; !ok {
		return nil, fmt.Errorf("chainContractCallers must contain L1ChainId: %d", config.L1ChainId)
	}

	var aggStrat IAggregationStrategy
	switch config.AggregationStrategy {
	case AggregationStrategyNone:
		aggStrat = NewSingleAvsAggregationStrategy()
	case AggregationStrategyStakeWeighted:
		aggStrat = NewStakeWeightedAggregationStrategy(
			config.AvsAddress,
			om,
			logger,
		)
	default:
		logger.Sugar().Errorw("Unknown aggregation strategy",
			zap.String("avsAddress", config.AvsAddress),
			zap.Any("aggregationStrategy", config.AggregationStrategy),
		)
		return nil, fmt.Errorf("unknown aggregation strategy: %d", config.AggregationStrategy)
	}

	manager := &AvsExecutionManager{
		config:               config,
		logger:               logger,
		chainContractCallers: chainContractCallers,
		signer:               signer,
		contractStore:        cs,
		operatorManager:      om,
		inflightTasks:        sync.Map{},
		taskQueue:            make(chan *types.Task, 10000),
		aggregationStrategy:  aggStrat,
	}
	return manager, nil
}

func hasExpectedMailboxContractsForChains(supportedChains []config.ChainId, mailboxAddresses map[config.ChainId]string) error {
	for _, chainId := range supportedChains {
		if _, ok := mailboxAddresses[chainId]; !ok {
			return fmt.Errorf("missing mailbox contract address for chain ID: %d", chainId)
		}
	}
	return nil
}

func hasExpectedContractCallersForChains(supportedChains []config.ChainId, contractCallers map[config.ChainId]contractCaller.IContractCaller) error {
	for _, chainId := range supportedChains {
		if _, ok := contractCallers[chainId]; !ok {
			return fmt.Errorf("missing contract caller for chain ID: %d", chainId)
		}
	}
	return nil
}

func (em *AvsExecutionManager) getListOfContractAddresses() []string {
	addrs := make([]string, 0, len(em.config.MailboxContractAddresses))
	for _, addr := range em.config.MailboxContractAddresses {
		addrs = append(addrs, strings.ToLower(addr))
	}
	return addrs
}

// Init initializes the AvsExecutionManager before starting
func (em *AvsExecutionManager) Init(ctx context.Context) error {
	em.logger.Sugar().Infow("Initializing AvsExecutionManager",
		zap.String("avsAddress", em.config.AvsAddress),
	)
	return nil
}

// Start starts the AvsExecutionManager
func (em *AvsExecutionManager) Start(ctx context.Context) error {
	em.logger.Sugar().Infow("Starting AvsExecutionManager",
		zap.String("contractAddress", em.config.AvsAddress),
		zap.Any("supportedChainIds", em.config.SupportedChainIds),
		zap.String("avsAddress", em.config.AvsAddress),
	)
	for {
		select {
		case task := <-em.taskQueue:
			em.logger.Sugar().Infow("Received task from queue",
				zap.String("taskId", task.TaskId),
			)
			if err := em.handleTask(ctx, task); err != nil {
				em.logger.Sugar().Errorw("Failed to handle task",
					"taskId", task.TaskId,
					"error", err,
				)
			}
		case <-ctx.Done():
			em.logger.Sugar().Infow("AvsExecutionManager context cancelled, exiting")
			return ctx.Err()
		}
	}
}

// HandleLog processes logs from the chain poller
func (em *AvsExecutionManager) HandleLog(lwb *chainPoller.LogWithBlock) error {
	em.logger.Sugar().Infow("Received log from chain poller",
		zap.Any("log", lwb),
	)
	lg := lwb.Log

	mailboxContract, _ := em.contractStore.GetContractByNameForChainId(config.ContractName_TaskMailbox, lwb.Block.ChainId)

	// Handle new task created
	if lg.EventName == "TaskCreated" {
		if mailboxContract == nil {
			em.logger.Sugar().Errorw("Mailbox contract not found for TaskCreated event",
				zap.String("eventName", lg.EventName),
				zap.String("contractAddress", lg.Address),
				zap.Uint("chainId", uint(lwb.Block.ChainId)),
				zap.Uint64("blockNumber", lwb.Block.Number.Value()),
				zap.String("transactionHash", lwb.RawLog.TransactionHash.Value()),
			)
			return nil
		}
		if strings.EqualFold(lwb.Log.Address, mailboxContract.Address) {
			isLeader, err := em.aggregationStrategy.IsLeaderForBlock(context.Background(), lwb.Block)
			if err != nil {
				em.logger.Sugar().Errorw("Failed to check if leader for block",
					zap.Uint64("blockNumber", lwb.Block.Number.Value()),
					zap.Error(err),
				)
				return fmt.Errorf("failed to check if leader for block: %w", err)
			}
			if isLeader {
				return em.processTask(lwb)
			}
			em.logger.Sugar().Infow("Not leader for block, ignoring TaskCreated event",
				zap.String("eventName", lg.EventName),
				zap.String("contractAddress", lg.Address),
				zap.Uint64("blockNumber", lwb.Block.Number.Value()),
				zap.String("transactionHash", lwb.RawLog.TransactionHash.Value()),
				zap.String("avsAddress", em.config.AvsAddress),
			)
			return nil
		}
	}

	em.logger.Sugar().Infow("Ignoring log",
		zap.String("eventName", lg.EventName),
		zap.String("contractAddress", lg.Address),
		zap.Strings("addresses", em.getListOfContractAddresses()),
	)
	return nil
}

func (em *AvsExecutionManager) handleTask(ctx context.Context, task *types.Task) error {
	em.logger.Sugar().Infow("Handling task",
		zap.String("taskId", task.TaskId),
	)
	if _, ok := em.inflightTasks.Load(task.TaskId); ok {
		return fmt.Errorf("task %s is already being processed", task.TaskId)
	}
	ctx, cancel := context.WithDeadline(ctx, *task.DeadlineUnixSeconds)
	defer cancel()

	sig, err := em.signer.SignMessage(task.Payload)
	if err != nil {
		return fmt.Errorf("failed to sign task payload: %w", err)
	}

	chainCC, err := em.getContractCallerForChain(task.ChainId)
	if err != nil {
		em.logger.Sugar().Errorw("Failed to get contract caller for chain",
			zap.Uint("chainId", uint(task.ChainId)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get contract caller for chain: %w", err)
	}

	operatorPeersWeight, err := em.operatorManager.GetExecutorPeersAndWeightsForBlock(
		ctx,
		task.ChainId,
		task.BlockNumber,
		task.OperatorSetId,
	)
	if err != nil {
		em.logger.Sugar().Errorw("Failed to get operator peers and weights",
			zap.Uint("chainId", uint(task.ChainId)),
			zap.Uint64("blockNumber", task.BlockNumber),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get operator peers and weights: %w", err)
	}

	ts, err := taskSession.NewTaskSession(
		ctx,
		cancel,
		task,
		em.config.AggregatorAddress,
		sig,
		operatorPeersWeight,
		em.logger,
	)
	if err != nil {
		em.logger.Sugar().Errorw("Failed to create task session",
			zap.String("taskId", task.TaskId),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create task session: %w", err)
	}

	em.logger.Sugar().Infow("Created task session",
		zap.Any("taskSession", ts),
	)
	em.inflightTasks.Store(task.TaskId, ts)

	doneChan := make(chan bool, 1)
	errorsChan := make(chan error, 1)

	// Process the task
	// - Distributed the task to operators in the set
	// - Wait for responses
	// - Aggregate the results
	go func(chainCC contractCaller.IContractCaller) {
		em.logger.Sugar().Infow("Processing task session",
			zap.String("taskId", task.TaskId),
		)
		cert, err := ts.Process()
		if err != nil {
			em.logger.Sugar().Errorw("Failed to process task",
				zap.String("taskId", task.TaskId),
				zap.Error(err),
			)
			errorsChan <- fmt.Errorf("failed to process task: %w", err)
			return
		}
		if cert == nil {
			em.logger.Sugar().Errorw("Received nil aggregate certificate",
				zap.String("taskId", task.TaskId),
			)
			errorsChan <- fmt.Errorf("received nil aggregate certificate")
			return
		}
		em.logger.Sugar().Infow("Received task response and certificate",
			zap.String("taskId", task.TaskId),
			zap.String("taskResponseDigest", string(cert.TaskResponseDigest)),
		)

		em.logger.Sugar().Infow("Calling chain contract", zap.Uint("chainId", uint(ts.Task.ChainId)))

		receipt, err := chainCC.SubmitTaskResultRetryable(ctx, cert, operatorPeersWeight.RootReferenceTimestamp)
		if err != nil {
			// TODO: emit metric
			em.logger.Sugar().Errorw("Failed to submit task result", "error", err)
			errorsChan <- fmt.Errorf("failed to submit task result: %w", err)
			return
		} else {
			em.logger.Sugar().Infow("Successfully submitted task result",
				zap.String("taskId", ts.Task.TaskId),
				zap.String("transactionHash", receipt.TxHash.String()),
			)
		}
		doneChan <- true
	}(chainCC)

	select {
	case <-doneChan:
		em.logger.Sugar().Infow("Task session completed",
			zap.String("taskId", task.TaskId),
		)
		return nil
	case <-errorsChan:
		em.logger.Sugar().Errorw("Task session failed", zap.Error(err))
		return err
	case <-ctx.Done():
		switch ctx.Err() {
		case context.Canceled:
			em.logger.Sugar().Errorw("task session context done",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
		case context.DeadlineExceeded:
			em.logger.Sugar().Errorw("task session context deadline exceeded",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
			return fmt.Errorf("task session context deadline exceeded: %w", ctx.Err())
		default:
			em.logger.Sugar().Errorw("task session encountered an error",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
			return fmt.Errorf("task session encountered an error: %w", ctx.Err())
		}

		return nil
	}
}

func (em *AvsExecutionManager) processTask(lwb *chainPoller.LogWithBlock) error {
	lg := lwb.Log
	em.logger.Sugar().Infow("Received TaskCreated event",
		zap.String("eventName", lg.EventName),
		zap.String("contractAddress", lg.Address),
	)
	task, err := types.NewTaskFromLog(lg, lwb.Block, lg.Address)
	if err != nil {
		return fmt.Errorf("failed to convert task: %w", err)
	}
	em.logger.Sugar().Infow("Converted task",
		zap.Any("task", task),
	)

	if task.AVSAddress != strings.ToLower(em.config.AvsAddress) {
		em.logger.Sugar().Infow("Ignoring task for different AVS address",
			zap.String("taskAvsAddress", task.AVSAddress),
			zap.String("currentAvsAddress", em.config.AvsAddress),
		)
		return nil
	}
	em.taskQueue <- task
	em.logger.Sugar().Infow("Added task to queue")
	return nil
}

func (em *AvsExecutionManager) getContractCallerForChain(chainId config.ChainId) (contractCaller.IContractCaller, error) {
	caller, ok := em.chainContractCallers[chainId]
	if !ok {
		return nil, fmt.Errorf("no contract caller found for chain ID: %d", chainId)
	}
	return caller, nil
}
