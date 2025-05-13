package avsExecutionManager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/taskSession"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
	"slices"
	"strings"
	"sync"
	"time"
)

type AvsExecutionManagerConfig struct {
	AvsAddress               string
	SupportedChainIds        []config.ChainId
	MailboxContractAddresses map[config.ChainId]string
	AggregatorAddress        string
	AggregatorUrl            string
	WriteDelaySeconds        time.Duration
}

type operatorSetRegistrationData struct {
	AvsId           string
	OperatorAddress string
	OperatorSetId   uint32
}

type AvsExecutionManager struct {
	logger *zap.Logger
	config *AvsExecutionManagerConfig

	// will be a proper type when another PR is merged
	chainContractCallers map[config.ChainId]contractCaller.IContractCaller

	signer signer.ISigner

	peeringDataFetcher peering.IPeeringDataFetcher

	operatorPeers map[string]*peering.OperatorPeerInfo

	taskQueue chan *types.Task

	resultsQueue chan *taskSession.TaskSession

	inflightTasks sync.Map
}

func NewAvsExecutionManager(
	config *AvsExecutionManagerConfig,
	chainContractCallers map[config.ChainId]contractCaller.IContractCaller,
	signer signer.ISigner,
	peeringDataFetcher peering.IPeeringDataFetcher,
	logger *zap.Logger,
) *AvsExecutionManager {
	manager := &AvsExecutionManager{
		config:               config,
		logger:               logger,
		chainContractCallers: chainContractCallers,
		signer:               signer,
		peeringDataFetcher:   peeringDataFetcher,
		inflightTasks:        sync.Map{},
		taskQueue:            make(chan *types.Task, 10000),
		resultsQueue:         make(chan *taskSession.TaskSession, 10000),
	}
	return manager
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
	peers, err := em.peeringDataFetcher.ListExecutorOperators(ctx, em.config.AvsAddress)
	if err != nil {
		return fmt.Errorf("failed to fetch executor peers: %w", err)
	}
	operatorPeers := map[string]*peering.OperatorPeerInfo{}
	for _, peer := range peers {
		operatorPeers[peer.OperatorAddress] = peer
	}

	em.operatorPeers = operatorPeers
	em.logger.Sugar().Infow("Fetched executor peers",
		zap.Int("numPeers", len(peers)),
		zap.Any("peers", peers),
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
			if err := em.HandleTask(ctx, task); err != nil {
				em.logger.Sugar().Errorw("Failed to handle task",
					"taskId", task.TaskId,
					"error", err,
				)
			}
		case result := <-em.resultsQueue:
			em.logger.Sugar().Infow("Received task result", zap.Any("taskSession", result))

			if chainCaller, ok := em.chainContractCallers[result.Task.ChainId]; ok {
				em.logger.Sugar().Infow("Calling chain contract", zap.Uint("chainId", uint(result.Task.ChainId)))

				// TODO: (brandon c) remove this and handle case of submission to same block task was created.
				time.Sleep(em.config.WriteDelaySeconds)

				if result.AggregateCertificate == nil {
					em.logger.Sugar().Errorw("Received nil aggregate certificate", zap.String("taskId", result.Task.TaskId))
					return fmt.Errorf("received nil aggregate certificate")
				}

				receipt, err := chainCaller.SubmitTaskResult(ctx, result.AggregateCertificate)
				if err != nil {
					// TODO: emit metric
					em.logger.Sugar().Errorw("Failed to submit task result", "error", err)
				} else {
					em.logger.Sugar().Infow("Successfully submitted task result",
						zap.String("taskId", result.Task.TaskId),
						zap.String("transactionHash", receipt.TxHash.String()),
					)
				}

				continue
			}
			// TODO: emit metric
			em.logger.Sugar().Errorw("Failed to find contract caller for task", "taskId", result.Task.TaskId)
			return fmt.Errorf("failed to find contract caller for task")
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
	if !slices.Contains(em.getListOfContractAddresses(), strings.ToLower(lg.Address)) {
		return nil
	}

	switch lg.EventName {
	case "TaskCreated":
		return em.processTask(lwb)
	case "OperatorAddedToOperatorSet":
		return em.processOperatorAdded(lwb)
	case "OperatorRemovedFromOperatorSet":
		return em.processOperatorRemoved(lwb)
	}

	em.logger.Sugar().Infow("Ignoring log",
		zap.String("eventName", lg.EventName),
		zap.String("contractAddress", lg.Address),
		zap.Strings("addresses", em.getListOfContractAddresses()),
	)
	return nil
}

func (em *AvsExecutionManager) HandleTask(ctx context.Context, task *types.Task) error {
	em.logger.Sugar().Infow("Handling task",
		zap.String("taskId", task.TaskId),
	)
	if _, ok := em.inflightTasks.Load(task.TaskId); ok {
		return fmt.Errorf("task %s is already being processed", task.TaskId)
	}
	ctx, cancel := context.WithDeadline(ctx, *task.DeadlineUnixSeconds)

	sig, err := em.signer.SignMessage(task.Payload)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to sign task payload: %w", err)
	}

	ts, err := taskSession.NewTaskSession(
		ctx,
		cancel,
		task,
		em.config.AggregatorAddress,
		em.config.AggregatorUrl,
		sig,
		em.resultsQueue,
		em.logger,
	)
	if err != nil {
		cancel()
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

	go func() {
		if err := ts.Process(); err != nil {
			em.logger.Sugar().Errorw("Failed to process task",
				zap.String("taskId", task.TaskId),
				zap.Error(err),
			)
		}
		<-ctx.Done()
		// check if deadline was reached
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			em.logger.Sugar().Errorw("Task session context deadline exceeded",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
			return
		}
		em.logger.Sugar().Errorw("Task session context done",
			zap.String("taskId", task.TaskId),
			zap.Error(ctx.Err()),
		)
	}()
	return nil
}

func (em *AvsExecutionManager) HandleTaskResultFromExecutor(taskResult *types.TaskResult) error {
	task, ok := em.inflightTasks.Load(taskResult.TaskId)
	if !ok {
		em.logger.Sugar().Warnw("Received result for unknown task")
		return nil
	}

	ts := task.(*taskSession.TaskSession)
	ts.RecordResult(taskResult)
	return nil
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
	var peers []*peering.OperatorPeerInfo
	for _, peer := range em.operatorPeers {
		if slices.Contains(peer.OperatorSetIds, task.OperatorSetId) {
			clonedPeer, err := peer.Copy()
			if err != nil {
				em.logger.Sugar().Errorw("Failed to clone peer",
					zap.String("peer", peer.OperatorAddress),
					zap.Error(err),
				)
				return fmt.Errorf("failed to clone peer: %w", err)
			}
			peers = append(peers, clonedPeer)
		}
	}
	task.RecipientOperators = peers
	em.taskQueue <- task
	em.logger.Sugar().Infow("Added task to queue")
	return nil
}

func (em *AvsExecutionManager) parseOperatorSetData(
	lwb *chainPoller.LogWithBlock,
) (operatorSetRegistrationData, error) {
	lg := lwb.Log
	em.logger.Sugar().Infow("Received operator registration event",
		zap.String("eventName", lg.EventName),
		zap.String("contractAddress", lg.Address),
	)

	operatorAddr, ok := lg.Arguments[0].Value.(string)
	if !ok {
		return operatorSetRegistrationData{}, fmt.Errorf("failed to parse operator address from event")
	}

	outputBytes, err := json.Marshal(lg.OutputData)
	if err != nil {
		return operatorSetRegistrationData{}, fmt.Errorf("failed to marshal output data: %w", err)
	}

	type operatorSetData struct {
		Avs string `json:"avs"`
		Id  uint32 `json:"id"`
	}

	var operatorSet operatorSetData
	if err := json.Unmarshal(outputBytes, &operatorSet); err != nil {
		return operatorSetRegistrationData{}, fmt.Errorf("failed to unmarshal operatorSet data: %w", err)
	}

	em.logger.Sugar().Infow("Parsed operator registration",
		zap.String("operator", operatorAddr),
		zap.String("avs", strings.ToLower(operatorSet.Avs)),
		zap.Uint32("operatorSetId", operatorSet.Id),
	)

	return operatorSetRegistrationData{
		AvsId:           operatorSet.Avs,
		OperatorAddress: operatorAddr,
		OperatorSetId:   operatorSet.Id,
	}, nil
}

func (em *AvsExecutionManager) processOperatorAdded(lwb *chainPoller.LogWithBlock) error {
	registration, err := em.parseOperatorSetData(lwb)
	if err != nil {
		return err
	}
	if registration.AvsId != em.config.AvsAddress {
		return nil
	}
	if operatorPeering, ok := em.operatorPeers[registration.OperatorAddress]; ok {
		operatorPeering.OperatorSetIds = append(operatorPeering.OperatorSetIds, registration.OperatorSetId)
		return nil
	}
	observedPeers, err := em.chainContractCallers[lwb.Block.ChainId].GetOperatorSetMembersWithPeering(
		registration.AvsId,
		registration.OperatorSetId,
	)
	if err != nil {
		// TODO: emit metric
		return err
	}
	for _, observedPeer := range observedPeers {
		if observedPeer.OperatorAddress == registration.OperatorAddress {
			em.operatorPeers[registration.OperatorAddress] = observedPeer
			break
		}
	}
	return nil
}

func (em *AvsExecutionManager) processOperatorRemoved(lwb *chainPoller.LogWithBlock) error {
	deregistration, err := em.parseOperatorSetData(lwb)
	if err != nil {
		return err
	}
	if deregistration.AvsId != em.config.AvsAddress {
		return nil
	}
	peerInfo, ok := em.operatorPeers[deregistration.OperatorAddress]
	if !ok {
		// TODO: emit metric
		return fmt.Errorf("peer not found for deregistration: %s", deregistration.OperatorAddress)
	}
	for i, operatorSetId := range peerInfo.OperatorSetIds {
		if deregistration.OperatorSetId == operatorSetId {
			peerInfo.OperatorSetIds = append(peerInfo.OperatorSetIds[:i], peerInfo.OperatorSetIds[i+1:]...)
			break
		}
	}
	return nil
}
