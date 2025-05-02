package avsExecutionManager

import (
	"context"
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
	"github.com/ethereum/go-ethereum/common"
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
func (em *AvsExecutionManager) HandleLog(lwb *chainPoller.LogWithBlock, log *log.DecodedLog) error {
	if log.EventName == "TaskCreated" && slices.Contains(em.getListOfContractAddresses(), strings.ToLower(log.Address)) {
		task, err := convertTask(log, lwb.Block, log.Address)
		if err != nil {
			return fmt.Errorf("failed to convert task: %w", err)
		}
		em.taskQueue <- task
	}

	return nil
}

func (em *AvsExecutionManager) HandleTask(ctx context.Context, task *types.Task) error {
	if _, ok := em.inflightTasks.Load(task.TaskId); ok {
		return fmt.Errorf("task %s is already being processed", task.TaskId)
	}
	ctx, cancel := context.WithDeadline(ctx, *task.DeadlineUnixSeconds)

	peers := []*peering.OperatorPeerInfo{}
	for _, peer := range em.peers {
		if uint32(peer.OperatorSetId) == task.OperatorSetId {
			peers = append(peers, peer)
		}
	}

	ts := taskSession.NewTaskSession(ctx, cancel, task, nil, peers, func() error {
		return nil
	}, em.logger)

	em.inflightTasks.Store(task.TaskId, ts)

	go func() {
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

		if err := ts.Process(); err != nil {
			em.logger.Sugar().Errorw("Failed to process task",
				zap.String("taskId", task.TaskId),
				zap.Error(err),
			)
		}
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
