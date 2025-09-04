package avsExecutionManager

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/taskSession"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

type AvsExecutionManagerConfig struct {
	AvsAddress               string
	SupportedChainIds        []config.ChainId
	MailboxContractAddresses map[config.ChainId]string
	AggregatorAddress        string
	L1ChainId                config.ChainId
	TlsEnabled               bool
}

type OperatorSet struct {
	Avs common.Address
	Id  uint32
}

type ConsensusType uint8

type OperatorSetTaskConsensus struct {
	ConsensusType ConsensusType
	Threshold     uint16
}

type OperatorSetTaskConfig struct {
	TaskSLA                *big.Int
	CurveType              config.CurveType
	TaskMetadata           []byte
	Consensus              OperatorSetTaskConsensus
	L1ReferenceBlockNumber uint64
}

type AvsConfig struct {
	contractCaller.AVSConfig

	curveType config.CurveType
}

type AvsExecutionManager struct {
	logger *zap.Logger

	avsConfig *AvsConfig

	config *AvsExecutionManagerConfig

	chainContractCallers map[config.ChainId]contractCaller.IContractCaller

	signers signer.Signers

	operatorManager *operatorManager.OperatorManager

	contractStore contractStore.IContractStore

	taskQueue chan *types.Task

	inflightTasks sync.Map

	store storage.AggregatorStore

	chainPollers map[config.ChainId]*EVMChainPoller.EVMChainPoller

	avsConfigMutex sync.Mutex

	pollersMutex sync.RWMutex
}

func NewAvsExecutionManager(
	config *AvsExecutionManagerConfig,
	chainContractCallers map[config.ChainId]contractCaller.IContractCaller,
	signers signer.Signers,
	cs contractStore.IContractStore,
	om *operatorManager.OperatorManager,
	taskQueue chan *types.Task,
	chainPollers map[config.ChainId]*EVMChainPoller.EVMChainPoller,
	store storage.AggregatorStore,
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
	if store == nil {
		return nil, fmt.Errorf("store is required")
	}

	manager := &AvsExecutionManager{
		config:               config,
		logger:               logger,
		chainContractCallers: chainContractCallers,
		signers:              signers,
		contractStore:        cs,
		operatorManager:      om,
		store:                store,
		inflightTasks:        sync.Map{},
		taskQueue:            taskQueue,
		chainPollers:         chainPollers,
	}
	return manager, nil
}

func (em *AvsExecutionManager) Start(ctx context.Context) error {

	em.logger.Sugar().Infow("Starting AvsExecutionManager",
		zap.String("contractAddress", em.config.AvsAddress),
		zap.Any("supportedChainIds", em.config.SupportedChainIds),
		zap.String("avsAddress", em.config.AvsAddress),
	)

	if err := em.recoverPendingTasks(ctx); err != nil {
		em.logger.Sugar().Warnw("Failed to recover pending tasks",
			"error", err,
			"avsAddress", em.config.AvsAddress)
		// Continue anyway - this is not a fatal error
	}

	if err := em.startPollers(ctx); err != nil {
		return fmt.Errorf("failed to start pollers: %w", err)
	}

	go func() {
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
				if ctx.Err() != nil {
					em.logger.Sugar().Errorw("Error stopping AvsExecutionManager")
				}
				return
			}
		}
	}()

	return nil
}

// recoverPendingTasks loads pending tasks from storage and re-queues them
func (em *AvsExecutionManager) recoverPendingTasks(ctx context.Context) error {
	pendingTasks, err := em.store.ListPendingTasksForAVS(ctx, em.config.AvsAddress)
	if err != nil {
		return fmt.Errorf("failed to list pending tasks: %w", err)
	}

	if len(pendingTasks) == 0 {
		return nil
	}

	em.logger.Sugar().Infow("Recovering pending tasks from storage",
		"count", len(pendingTasks),
		"avsAddress", em.config.AvsAddress)

	recovered := 0
	for _, task := range pendingTasks {

		if task.DeadlineUnixSeconds != nil && time.Now().After(*task.DeadlineUnixSeconds) {
			em.logger.Sugar().Warnw("Skipping expired task during recovery",
				"taskId", task.TaskId,
				"deadline", task.DeadlineUnixSeconds.Unix(),
				"currentTime", time.Now().Unix())

			if err := em.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusFailed); err != nil {
				em.logger.Sugar().Warnw("Failed to mark expired task as failed",
					"error", err,
					"taskId", task.TaskId)
			}
			continue
		}

		if _, exists := em.inflightTasks.Load(task.TaskId); exists {
			em.logger.Sugar().Warnw("Task already in flight, skipping recovery",
				"taskId", task.TaskId)
			continue
		}

		select {
		case em.taskQueue <- task:
			recovered++
			em.logger.Sugar().Infow("Re-queued recovered task",
				"taskId", task.TaskId,
				"avsAddress", task.AVSAddress)
		default:
			em.logger.Sugar().Warnw("Task queue full, cannot recover task",
				"taskId", task.TaskId)
			// If we can't queue it now, it will be picked up on next restart
			break
		}
	}

	em.logger.Sugar().Infow("Task recovery completed",
		"totalPending", len(pendingTasks),
		"recovered", recovered,
		"avsAddress", em.config.AvsAddress)

	return nil
}

// startPollers starts all chain pollers for this AVS
func (em *AvsExecutionManager) startPollers(ctx context.Context) error {
	em.pollersMutex.RLock()
	defer em.pollersMutex.RUnlock()

	if len(em.chainPollers) == 0 {
		em.logger.Sugar().Infow("No chain pollers configured for AVS",
			zap.String("avsAddress", em.config.AvsAddress))
		return nil
	}

	wg := sync.WaitGroup{}
	errChan := make(chan error, len(em.chainPollers))

	for chainId, poller := range em.chainPollers {
		wg.Add(1)
		em.logger.Sugar().Infow("Starting poller for chain",
			zap.Uint("chainId", uint(chainId)),
			zap.String("avsAddress", em.config.AvsAddress))

		go func(p *EVMChainPoller.EVMChainPoller, cId config.ChainId) {
			if err := p.Start(ctx); err != nil {
				em.logger.Sugar().Errorw("Poller stopped with error",
					zap.Error(err),
					zap.Uint("chainId", uint(cId)),
					zap.String("avsAddress", em.config.AvsAddress))
				errChan <- err
			}
			wg.Done()
		}(poller, chainId)
	}

	wg.Wait()
	if err := <-errChan; err != nil {
		return fmt.Errorf("failed to start pollers: %w", err)
	}

	em.logger.Sugar().Infow("Started all chain pollers",
		zap.Int("count", len(em.chainPollers)),
		zap.String("avsAddress", em.config.AvsAddress))

	return nil
}

func (em *AvsExecutionManager) getOperatorSetTaskConfig(
	ctx context.Context,
	task *types.Task,
) (*OperatorSetTaskConfig, error) {

	taskChainId := task.ChainId
	cc, err := em.getContractCallerForChain(taskChainId)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract caller for chain %d: %w", taskChainId, err)
	}

	l1ReferenceBlockNumber, err := em.getL1BlockForChainBlock(ctx, taskChainId, task.SourceBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get l1 reference block number for chain %d: %w", taskChainId, err)
	}

	var opsetConfig *contractCaller.TaskMailboxExecutorOperatorSetConfig
	executorOpSetBlockNumber := task.SourceBlockNumber

	if task.ChainId == em.config.L1ChainId {
		executorOpSetBlockNumber = l1ReferenceBlockNumber
	}
	opsetConfig, err = cc.GetExecutorOperatorSetTaskConfig(ctx, common.HexToAddress(task.AVSAddress), task.OperatorSetId, executorOpSetBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator set config for chain %d: %w", taskChainId, err)
	}

	curveType, err := opsetConfig.GetCurveType()
	if err != nil {
		return nil, fmt.Errorf("failed to get curve type from operator set config: %w", err)
	}

	consensusValue, err := opsetConfig.GetConsensusValue()
	if err != nil {
		return nil, fmt.Errorf("failed to get consensus value from operator set config: %w", err)
	}

	taskConfig := &OperatorSetTaskConfig{
		TaskSLA:      opsetConfig.TaskSLA,
		CurveType:    curveType,
		TaskMetadata: opsetConfig.TaskMetadata,
		Consensus: OperatorSetTaskConsensus{
			ConsensusType: ConsensusType(opsetConfig.Consensus.ConsensusType),
			Threshold:     consensusValue,
		},
		L1ReferenceBlockNumber: l1ReferenceBlockNumber,
	}

	return taskConfig, nil
}

func (em *AvsExecutionManager) getAggregatorTaskConfig(_ context.Context, blockNumber uint64) (*AvsConfig, error) {
	em.avsConfigMutex.Lock()
	defer em.avsConfigMutex.Unlock()

	cc, ok := em.chainContractCallers[em.config.L1ChainId]
	if !ok {
		return nil, fmt.Errorf("no contract caller found for L1ChainId: %d", em.config.L1ChainId)
	}
	avsConfig, err := cc.GetAVSConfig(em.config.AvsAddress, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get AVS config: %w", err)
	}

	curveType, err := cc.GetOperatorSetCurveType(em.config.AvsAddress, avsConfig.AggregatorOperatorSetId, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get curve type for operator set: %w", err)
	}
	em.avsConfig = &AvsConfig{
		AVSConfig: contractCaller.AVSConfig{
			AggregatorOperatorSetId: avsConfig.AggregatorOperatorSetId,
			ExecutorOperatorSetIds:  avsConfig.ExecutorOperatorSetIds,
		},
		curveType: curveType,
	}

	return em.avsConfig, nil
}

func (em *AvsExecutionManager) handleTask(ctx context.Context, task *types.Task) error {
	em.logger.Sugar().Infow("Handling task",
		zap.String("taskId", task.TaskId),
	)
	if _, ok := em.inflightTasks.Load(task.TaskId); ok {
		return fmt.Errorf("task %s is already being processed", task.TaskId)
	}

	// Update task status to processing
	if err := em.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusProcessing); err != nil {
		em.logger.Sugar().Warnw("Failed to update task status to processing",
			"error", err,
			"taskId", task.TaskId,
		)
	}
	ctx, cancel := context.WithDeadline(ctx, *task.DeadlineUnixSeconds)
	defer cancel()

	executorTaskConfig, err := em.getOperatorSetTaskConfig(ctx, task)
	if err != nil {
		em.logger.Sugar().Errorw("Failed to get or set operator set task config",
			zap.String("taskId", task.TaskId),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get or set operator set task config: %w", err)
	}

	task.ThresholdBips = executorTaskConfig.Consensus.Threshold
	task.L1ReferenceBlockNumber = executorTaskConfig.L1ReferenceBlockNumber

	avsConfig, err := em.getAggregatorTaskConfig(ctx, task.L1ReferenceBlockNumber)
	if err != nil {
		em.logger.Sugar().Errorw("Failed to get or set aggregator task config",
			zap.String("taskId", task.TaskId),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get or set aggregator task config: %w", err)
	}

	var signerToUse signer.ISigner
	switch avsConfig.curveType {
	case config.CurveTypeBN254:
		signerToUse = em.signers.BLSSigner
	case config.CurveTypeECDSA:
		signerToUse = em.signers.ECDSASigner
	default:
		em.logger.Sugar().Errorw("Unsupported curve type for task",
			zap.String("taskId", task.TaskId),
			zap.String("curveType", avsConfig.curveType.String()),
		)
		return fmt.Errorf("unsupported curve type: %s", avsConfig.curveType)
	}

	chainCC, err := em.getContractCallerForChain(task.ChainId)
	if err != nil {
		em.logger.Sugar().Errorw("Failed to get contract caller for chain",
			zap.Uint("chainId", uint(task.ChainId)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get contract caller for chain: %w", err)
	}

	// TODO (FromTheRain):  reuse the calculated L1ReferenceBlockNumber
	operatorPeersWeight, err := em.operatorManager.GetExecutorPeersAndWeightsForBlock(
		ctx,
		task.ChainId,
		// This must be source block number so this method can translate to reference block.
		task.SourceBlockNumber,
		task.OperatorSetId,
	)
	if err != nil {
		em.logger.Sugar().Errorw("Failed to get operator peers and weights",
			zap.Uint("chainId", uint(task.ChainId)),
			zap.Uint64("blockNumber", task.L1ReferenceBlockNumber),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get operator peers and weights: %w", err)
	}
	fmt.Printf("Operator peers and weights: %v\n", operatorPeersWeight)

	opsetCurveType, err := em.operatorManager.GetCurveTypeForOperatorSet(ctx, task.AVSAddress, task.OperatorSetId, task.L1ReferenceBlockNumber)
	if err != nil {
		em.logger.Sugar().Errorw("Failed to get curve type for operator set",
			zap.String("avsAddress", task.AVSAddress),
			zap.Uint32("operatorSetId", task.OperatorSetId),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get curve type for operator set: %w", err)
	}

	// Get the L1 contract caller
	l1Cc, ok := em.chainContractCallers[em.config.L1ChainId]
	if !ok {
		return fmt.Errorf("no L1 contract caller found")
	}

	// TODO: pass in the known values we indexed to verify against the response during aggregation
	if opsetCurveType == config.CurveTypeBN254 {
		ts, err := taskSession.NewBN254TaskSession(
			ctx,
			cancel,
			task,
			l1Cc,
			em.config.AggregatorAddress,
			signerToUse,
			operatorPeersWeight,
			em.config.TlsEnabled,
			em.logger,
		)
		if err != nil {
			em.logger.Sugar().Errorw("Failed to create task session",
				zap.String("taskId", task.TaskId),
				zap.Error(err),
			)
			return fmt.Errorf("failed to create task session: %w", err)
		}
		return em.processBN254Task(ctx, task, ts, chainCC, operatorPeersWeight)
	} else if opsetCurveType == config.CurveTypeECDSA {
		ts, err := taskSession.NewECDSATaskSession(
			ctx,
			cancel,
			task,
			l1Cc,
			em.config.AggregatorAddress,
			signerToUse,
			operatorPeersWeight,
			em.config.TlsEnabled,
			em.logger,
		)
		if err != nil {
			em.logger.Sugar().Errorw("Failed to create task session",
				zap.String("taskId", task.TaskId),
				zap.Error(err),
			)
			return fmt.Errorf("failed to create task session: %w", err)
		}
		return em.processECDSATask(ctx, task, ts, chainCC, operatorPeersWeight)
	}
	em.logger.Sugar().Errorw("Unsupported curve type for task",
		zap.String("taskId", task.TaskId),
		zap.String("curveType", opsetCurveType.String()),
	)
	return fmt.Errorf("unsupported curve type: %s", opsetCurveType)
}

func (em *AvsExecutionManager) processBN254Task(
	ctx context.Context,
	task *types.Task,
	ts *taskSession.TaskSession[bn254.Signature, aggregation.AggregatedBN254Certificate, signing.PublicKey],
	chainCC contractCaller.IContractCaller,
	operatorPeersWeight *operatorManager.PeerWeight,
) error {
	em.logger.Sugar().Infow("Created BN254 task session",
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
			zap.String("taskResponseDigest", hexutil.Encode(cert.TaskResponseDigest[:])),
		)

		// Convert certificate to submission parameters
		params := cert.ToSubmitParams()
		receipt, err := chainCC.SubmitBN254TaskResultRetryable(ctx, params, operatorPeersWeight.RootReferenceTimestamp)
		if err != nil {
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
		// Update task status to completed
		if err := em.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted); err != nil {
			em.logger.Sugar().Warnw("Failed to update task status to completed",
				"error", err,
				"taskId", task.TaskId,
			)
		}
		// Remove from inflight tasks
		em.inflightTasks.Delete(task.TaskId)
		return nil
	case err := <-errorsChan:
		em.logger.Sugar().Errorw("Task session failed", zap.Error(err))
		// Update task status to failed
		if updateErr := em.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusFailed); updateErr != nil {
			em.logger.Sugar().Warnw("Failed to update task status to failed",
				"error", updateErr,
				"taskId", task.TaskId,
			)
		}
		// Remove from inflight tasks
		em.inflightTasks.Delete(task.TaskId)
		return err
	case <-ctx.Done():
		switch err := ctx.Err(); {
		case errors.Is(err, context.Canceled):
			em.logger.Sugar().Errorw("task session context done",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
		case errors.Is(err, context.DeadlineExceeded):
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

func (em *AvsExecutionManager) processECDSATask(
	ctx context.Context,
	task *types.Task,
	ts *taskSession.TaskSession[ecdsa.Signature, aggregation.AggregatedECDSACertificate, common.Address],
	chainCC contractCaller.IContractCaller,
	operatorPeersWeight *operatorManager.PeerWeight,
) error {
	em.logger.Sugar().Infow("Created ECDSA task session",
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
			zap.String("taskResponseDigest", hexutil.Encode(cert.TaskResponseDigest[:])),
		)

		// Convert certificate to submission parameters
		params := cert.ToSubmitParams()
		receipt, err := chainCC.SubmitECDSATaskResultRetryable(ctx, params, operatorPeersWeight.RootReferenceTimestamp)
		if err != nil {
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
		// Update task status to completed
		if err := em.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusCompleted); err != nil {
			em.logger.Sugar().Warnw("Failed to update task status to completed",
				"error", err,
				"taskId", task.TaskId,
			)
		}
		// Remove from inflight tasks
		em.inflightTasks.Delete(task.TaskId)
		return nil
	case err := <-errorsChan:
		em.logger.Sugar().Errorw("Task session failed", zap.Error(err))
		// Update task status to failed
		if updateErr := em.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusFailed); updateErr != nil {
			em.logger.Sugar().Warnw("Failed to update task status to failed",
				"error", updateErr,
				"taskId", task.TaskId,
			)
		}
		// Remove from inflight tasks
		em.inflightTasks.Delete(task.TaskId)
		return err
	case <-ctx.Done():
		switch err := ctx.Err(); {
		case errors.Is(err, context.Canceled):
			em.logger.Sugar().Errorw("task session context done",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
		case errors.Is(err, context.DeadlineExceeded):
			em.logger.Sugar().Errorw("task session context deadline exceeded",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
			if updateErr := em.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusFailed); updateErr != nil {
				em.logger.Sugar().Warnw("Failed to update task status to failed",
					"error", updateErr,
					"taskId", task.TaskId,
				)
			}
			return fmt.Errorf("task session context deadline exceeded: %w", ctx.Err())
		default:
			if updateErr := em.store.UpdateTaskStatus(ctx, task.TaskId, storage.TaskStatusFailed); updateErr != nil {
				em.logger.Sugar().Warnw("Failed to update task status to failed",
					"error", updateErr,
					"taskId", task.TaskId,
				)
			}
			em.logger.Sugar().Errorw("task session encountered an error",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
			return fmt.Errorf("task session encountered an error: %w", ctx.Err())
		}

		return nil
	}
}

func (em *AvsExecutionManager) getContractCallerForChain(chainId config.ChainId) (contractCaller.IContractCaller, error) {
	caller, ok := em.chainContractCallers[chainId]
	if !ok {
		return nil, fmt.Errorf("no contract caller found for chain ID: %d", chainId)
	}
	return caller, nil
}

func (em *AvsExecutionManager) getL1BlockForChainBlock(ctx context.Context, chainId config.ChainId, blockNumber uint64) (uint64, error) {
	// If this is L1, return the block number directly
	if chainId == em.config.L1ChainId {
		return blockNumber, nil
	}

	// Get the L1 contract caller
	l1Cc, ok := em.chainContractCallers[em.config.L1ChainId]
	if !ok {
		return 0, fmt.Errorf("no L1 contract caller found")
	}

	// Get the target chain contract caller
	targetChainCc, ok := em.chainContractCallers[chainId]
	if !ok {
		return 0, fmt.Errorf("no contract caller found for chain ID: %d", chainId)
	}

	// Get supported chains from L1 to find the table updater address using -1 (latest) for L2 chains
	destChainIds, tableUpdaterAddresses, err := l1Cc.GetSupportedChainsForMultichain(ctx, -1)
	if err != nil {
		return 0, fmt.Errorf("failed to get supported chains: %w", err)
	}

	// Find the table updater address for the target chain
	var destTableUpdaterAddress common.Address
	for i, destChainId := range destChainIds {
		if destChainId.Uint64() == uint64(chainId) {
			destTableUpdaterAddress = tableUpdaterAddresses[i]
			break
		}
	}

	if destTableUpdaterAddress == (common.Address{}) {
		return 0, fmt.Errorf("no table updater address found for chain ID %d", chainId)
	}

	// Get the reference time and block from the target chain
	latestReferenceTimeAndBlock, err := targetChainCc.GetTableUpdaterReferenceTimeAndBlock(
		ctx,
		destTableUpdaterAddress,
		blockNumber,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get reference time and block: %w", err)
	}

	return uint64(latestReferenceTimeAndBlock.LatestReferenceBlockNumber), nil
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
