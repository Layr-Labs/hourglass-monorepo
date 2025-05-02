package avsExecutionManager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/taskSession"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
	"math/big"
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
}

type AvsExecutionManager struct {
	logger *zap.Logger
	config *AvsExecutionManagerConfig

	// will be a proper type when another PR is merged
	chainContractCallers map[config.ChainId]interface{}

	signer signer.ISigner

	peeringDataFetcher peering.IPeeringDataFetcher

	peers []*peering.OperatorPeerInfo

	taskQueue chan *types.Task

	resultsQueue chan *taskSession.TaskSession

	inflightTasks sync.Map
}

func NewAvsExecutionManager(
	config *AvsExecutionManagerConfig,
	chainContractCallers map[config.ChainId]interface{},
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
	peers, err := em.peeringDataFetcher.ListExecutorOperators()
	if err != nil {
		return fmt.Errorf("failed to fetch executor peers: %w", err)
	}

	em.peers = peers
	em.logger.Sugar().Infow("Fetched executor peers",
		zap.Int("numPeers", len(peers)),
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
			// TODO: post result to the chain
			em.logger.Sugar().Infow("Received task result",
				zap.Any("taskSession", result),
			)
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
	if lg.EventName == "TaskCreated" && slices.Contains(em.getListOfContractAddresses(), strings.ToLower(lg.Address)) {
		em.logger.Sugar().Infow("Received TaskCreated event",
			zap.String("eventName", lg.EventName),
			zap.String("contractAddress", lg.Address),
		)
		task, err := convertTask(lg, lwb.Block, lg.Address)
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
	} else {

		em.logger.Sugar().Infow("Ignoring log",
			zap.String("eventName", lg.EventName),
			zap.String("contractAddress", lg.Address),
			zap.Strings("addresses", em.getListOfContractAddresses()),
		)
	}

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

	peers := []*peering.OperatorPeerInfo{}
	for _, peer := range em.peers {
		if slices.Contains(peer.OperatorSetIds, task.OperatorSetId) {
			peers = append(peers, peer)
		}
	}

	sig, err := em.signer.SignMessage(task.Payload)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to sign task payload: %w", err)
	}

	ts := taskSession.NewTaskSession(
		ctx,
		cancel,
		task,
		em.config.AggregatorAddress,
		em.config.AggregatorUrl,
		sig,
		peers,
		em.resultsQueue,
		em.logger,
	)

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

func convertTask(log *log.DecodedLog, block *ethereum.EthereumBlock, inboxAddress string) (*types.Task, error) {
	var avsAddress string
	var taskId string

	taskId, ok := log.Arguments[1].Value.(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse task id")
	}
	avsAddress, ok = log.Arguments[2].Value.(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse task event address")
	}

	// it aint stupid if it works...
	// take the output data, turn it into a json string, then Unmarshal it into a typed struct
	// rather than trying to coerce data types
	outputBytes, err := json.Marshal(log.OutputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output data: %w", err)
	}

	type outputDataType struct {
		ExecutorOperatorSetId uint32
		TaskDeadline          uint64
		Payload               []byte
	}
	var od *outputDataType
	if err := json.Unmarshal(outputBytes, &od); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output data: %w", err)
	}
	parsedTaskDeadline := new(big.Int).SetUint64(od.TaskDeadline)
	taskDeadlineTime := time.Now().Add(time.Duration(parsedTaskDeadline.Int64()) * time.Second)

	return &types.Task{
		TaskId:              taskId,
		AVSAddress:          strings.ToLower(avsAddress),
		OperatorSetId:       od.ExecutorOperatorSetId,
		CallbackAddr:        inboxAddress,
		DeadlineUnixSeconds: &taskDeadlineTime,
		Payload:             []byte(od.Payload),
		ChainId:             block.ChainId,
		BlockNumber:         block.Number.Value(),
		BlockHash:           block.Hash.Value(),
	}, nil
}
