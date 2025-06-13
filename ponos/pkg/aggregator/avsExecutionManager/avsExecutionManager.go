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
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"slices"
	"strings"
	"sync"
)

type AvsExecutionManagerConfig struct {
	AvsAddress               string
	SupportedChainIds        []config.ChainId
	MailboxContractAddresses map[config.ChainId]string
	AggregatorAddress        string
	AggregatorUrl            string
	L1ChainId                config.ChainId
}

type operatorSetRegistrationData struct {
	Avs             string
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

	inflightTasks sync.Map
}

func NewAvsExecutionManager(
	config *AvsExecutionManagerConfig,
	chainContractCallers map[config.ChainId]contractCaller.IContractCaller,
	signer signer.ISigner,
	peeringDataFetcher peering.IPeeringDataFetcher,
	logger *zap.Logger,
) (*AvsExecutionManager, error) {
	if config.L1ChainId == 0 {
		return nil, fmt.Errorf("L1ChainId must be set in AvsExecutionManagerConfig")
	}
	if _, ok := chainContractCallers[config.L1ChainId]; !ok {
		return nil, fmt.Errorf("chainContractCallers must contain L1ChainId: %d", config.L1ChainId)
	}

	manager := &AvsExecutionManager{
		config:               config,
		logger:               logger,
		chainContractCallers: chainContractCallers,
		signer:               signer,
		peeringDataFetcher:   peeringDataFetcher,
		inflightTasks:        sync.Map{},
		taskQueue:            make(chan *types.Task, 10000),
	}
	return manager, nil
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
		em.logger.Sugar().Infow("Ignoring log from different contract",
			zap.String("contractAddress", lg.Address),
			zap.Strings("addresses", em.getListOfContractAddresses()),
		)
		return nil
	}

	switch lg.EventName {
	case "TaskCreated":
		return em.processTask(lwb)
	case "OperatorAddedToOperatorSet":
		return em.processOperatorAddedToOperatorSet(lwb)
	case "OperatorRemovedFromOperatorSet":
		return em.processOperatorRemovedFromOperatorSet(lwb)
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

	chainCC, err := em.getContractCallerForChain(task.ChainId)
	if err != nil {
		cancel()
		em.logger.Sugar().Errorw("Failed to get contract caller for chain",
			zap.Uint("chainId", uint(task.ChainId)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get contract caller for chain: %w", err)
	}

	tableData, err := chainCC.GetOperatorTableDataForOperatorSet(
		ctx,
		common.HexToAddress(task.AVSAddress),
		task.OperatorSetId,
		task.ChainId,
		task.BlockNumber,
	)
	if err != nil {
		cancel()
		em.logger.Sugar().Errorw("Failed to get operator table data",
			zap.String("avsAddress", task.AVSAddress),
			zap.Uint32("operatorSetId", task.OperatorSetId),
			zap.Uint("chainId", uint(task.ChainId)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get operator table data: %w", err)
	}

	ts, err := taskSession.NewTaskSession(
		ctx,
		cancel,
		task,
		em.config.AggregatorAddress,
		sig,
		tableData,
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

	doneChan := make(chan bool, 1)
	errorsChan := make(chan error, 1)

	go func() {
		em.logger.Sugar().Infow("Processing task session",
			zap.String("taskId", task.TaskId),
		)
		cert, err := ts.Process()
		if err != nil {
			cancel()
			em.logger.Sugar().Errorw("Failed to process task",
				zap.String("taskId", task.TaskId),
				zap.Error(err),
			)
			errorsChan <- fmt.Errorf("failed to process task: %w", err)
			return
		}
		em.logger.Sugar().Infow("Received task response and certificate",
			zap.String("taskId", task.TaskId),
			zap.String("taskResponseDigest", string(cert.TaskResponseDigest)),
		)

		chainCaller, ok := em.chainContractCallers[ts.Task.ChainId]
		if !ok {
			errorsChan <- fmt.Errorf("failed to find chain caller for task: %s", task.TaskId)
			return
		}

		em.logger.Sugar().Infow("Calling chain contract", zap.Uint("chainId", uint(ts.Task.ChainId)))

		if cert == nil {
			em.logger.Sugar().Errorw("Received nil aggregate certificate", zap.String("taskId", ts.Task.TaskId))
			errorsChan <- fmt.Errorf("received nil aggregate certificate")
			return
		}

		receipt, err := chainCaller.SubmitTaskResultRetryable(ctx, cert, tableData.LatestReferenceTimestamp)
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
	}()

	select {
	case <-doneChan:
		em.logger.Sugar().Infow("Task session completed",
			zap.String("taskId", task.TaskId),
		)
	case <-errorsChan:
		em.logger.Sugar().Errorw("Task session failed", zap.Error(err))
		return err
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			em.logger.Sugar().Errorw("Task session context deadline exceeded",
				zap.String("taskId", task.TaskId),
				zap.Error(ctx.Err()),
			)
			return fmt.Errorf("task session context deadline exceeded: %w", ctx.Err())
		}
		em.logger.Sugar().Errorw("Task session context done",
			zap.String("taskId", task.TaskId),
			zap.Error(ctx.Err()),
		)
		return nil
	}
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
		if peer.IncludesOperatorSetId(task.OperatorSetId) {
			clonedPeer, err := peer.Clone()
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

func (em *AvsExecutionManager) parseOperatorAddedRemovedFromSet(lwb *chainPoller.LogWithBlock) (operatorSetRegistrationData, error) {
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
		Avs:             operatorSet.Avs,
		OperatorAddress: operatorAddr,
		OperatorSetId:   operatorSet.Id,
	}, nil
}

func (em *AvsExecutionManager) getContractCallerForChain(chainId config.ChainId) (contractCaller.IContractCaller, error) {
	caller, ok := em.chainContractCallers[chainId]
	if !ok {
		return nil, fmt.Errorf("no contract caller found for chain ID: %d", chainId)
	}
	return caller, nil
}

func (em *AvsExecutionManager) getL1ContractCaller() (contractCaller.IContractCaller, error) {
	return em.getContractCallerForChain(em.config.L1ChainId)
}

func (em *AvsExecutionManager) processOperatorAddedToOperatorSet(lwb *chainPoller.LogWithBlock) error {
	registration, err := em.parseOperatorAddedRemovedFromSet(lwb)
	if err != nil {
		return err
	}
	if !strings.EqualFold(registration.Avs, em.config.AvsAddress) {
		em.logger.Sugar().Infow("Ignoring operator registration for different AVS address",
			zap.String("avsId", registration.Avs),
			zap.String("currentAvsAddress", em.config.AvsAddress),
			zap.String("operatorAddress", registration.OperatorAddress),
			zap.Uint32("operatorSetId", registration.OperatorSetId),
		)
		return nil
	}
	l1cc, err := em.getL1ContractCaller()
	if err != nil {
		return fmt.Errorf("failed to get L1 contract caller: %w", err)
	}

	// if the operator is already present in the map, just append the new operator set id
	if operatorPeering, ok := em.operatorPeers[registration.OperatorAddress]; ok {

		opsetDetails, err := l1cc.GetOperatorSetDetailsForOperator(
			common.HexToAddress(registration.OperatorAddress),
			registration.Avs,
			registration.OperatorSetId,
		)
		if err != nil {
			return fmt.Errorf("failed to get operator set details: %w", err)
		}

		operatorPeering.OperatorSets = append(operatorPeering.OperatorSets, opsetDetails)
		return nil
	}

	// first time seeing the operator, so need to fetch all of their data
	observedPeers, err := l1cc.GetOperatorSetMembersWithPeering(
		registration.Avs,
		registration.OperatorSetId,
	)
	if err != nil {
		// TODO: emit metric
		return err
	}
	for _, observedPeer := range observedPeers {
		if strings.EqualFold(observedPeer.OperatorAddress, registration.OperatorAddress) {
			em.operatorPeers[registration.OperatorAddress] = observedPeer
			break
		}
	}
	return nil
}

func (em *AvsExecutionManager) processOperatorRemovedFromOperatorSet(lwb *chainPoller.LogWithBlock) error {
	deregistration, err := em.parseOperatorAddedRemovedFromSet(lwb)
	if err != nil {
		return err
	}
	if !strings.EqualFold(deregistration.Avs, em.config.AvsAddress) {
		em.logger.Sugar().Infow("Ignoring operator deregistration for different AVS address",
			zap.String("avsId", deregistration.Avs),
			zap.String("currentAvsAddress", em.config.AvsAddress),
			zap.String("operatorAddress", deregistration.OperatorAddress),
			zap.Uint32("operatorSetId", deregistration.OperatorSetId),
		)
		return nil
	}
	peerInfo, ok := em.operatorPeers[deregistration.OperatorAddress]
	if !ok {
		em.logger.Sugar().Infow("Operator not found in peers, ignoring deregistration",
			zap.String("operatorAddress", deregistration.OperatorAddress),
			zap.Uint32("operatorSetId", deregistration.OperatorSetId),
			zap.String("avsAddress", deregistration.Avs),
		)
		return nil
	}
	peerInfo.OperatorSets = slices.DeleteFunc(peerInfo.OperatorSets, func(os *peering.OperatorSet) bool {
		return os.OperatorSetID == deregistration.OperatorSetId
	})
	em.operatorPeers[deregistration.OperatorAddress] = peerInfo
	return nil
}
