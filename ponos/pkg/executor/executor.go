package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerCapacityPlanner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerPoolManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const artifactRegistry = "ArtifactRegistry"

type Executor struct {
	logger        *zap.Logger
	config        *executorConfig.ExecutorConfig
	rpcServer     *rpcServer.RpcServer
	signer        signer.ISigner
	inflightTasks *sync.Map

	performerPoolManager *performerPoolManager.PerformerPoolManager
	capacityPlanner      *performerCapacityPlanner.PerformerCapacityPlanner
	peeringFetcher       peering.IPeeringDataFetcher

	// Chain events channel to be shared with pollers and planners
	chainEventsChan chan *chainPoller.LogWithBlock

	// Chain poller for monitoring on-chain events
	chainPoller chainPoller.IChainPoller

	// Transaction log parser for processing events
	transactionLogParser *transactionLogParser.TransactionLogParser

	// Contract store for accessing contract addresses and ABIs
	contractStore contractStore.IContractStore

	// Ethereum client
	ethereumClient *ethereum.Client

	// Contract caller for querying on-chain state
	contractCaller contractCaller.IContractCaller
}

func NewExecutorWithRpcServer(
	port int,
	config *executorConfig.ExecutorConfig,
	logger *zap.Logger,
	signer signer.ISigner,
	peeringFetcher peering.IPeeringDataFetcher,
	contractStore contractStore.IContractStore,
) (*Executor, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: port,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %v", err)
	}

	return NewExecutor(config, rpc, logger, signer, peeringFetcher, contractStore), nil
}

func NewExecutor(
	config *executorConfig.ExecutorConfig,
	rpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
	signer signer.ISigner,
	peeringFetcher peering.IPeeringDataFetcher,
	contractStore contractStore.IContractStore,
) *Executor {
	return &Executor{
		logger:          logger,
		config:          config,
		rpcServer:       rpcServer,
		signer:          signer,
		inflightTasks:   &sync.Map{},
		peeringFetcher:  peeringFetcher,
		contractStore:   contractStore,
		chainEventsChan: make(chan *chainPoller.LogWithBlock, 10000),
	}
}

func (e *Executor) Initialize() error {
	e.logger.Sugar().Infow("Initializing executor...")

	// Initialize transaction log parser
	e.transactionLogParser = transactionLogParser.NewTransactionLogParser(e.contractStore, e.logger)
	var artifactRegistryAddress string
	artifactRegistryContract := util.Find(e.contractStore.ListContracts(), func(c *contracts.Contract) bool {
		return strings.ToLower(c.Name) == strings.ToLower(config.ContractName_ArtifactRegistry)
	})
	if artifactRegistryContract == nil {
		return fmt.Errorf("could not find avs artifact registry contract")
	}
	artifactRegistryAddress = artifactRegistryContract.Address

	// Initialize Ethereum clients
	e.ethereumClient = ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   e.config.L1Chain.RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, e.logger)

	// Initialize contract callers
	var err error
	e.contractCaller, err = e.initializeContractCaller(artifactRegistryAddress)
	if err != nil {
		return fmt.Errorf("failed to initialize contract callers: %w", err)
	}

	// Initialize capacity planner with contract callers
	operatorAddress := e.config.Operator.Address
	e.capacityPlanner = performerCapacityPlanner.NewPerformerCapacityPlanner(
		e.logger,
		operatorAddress,
		e.contractStore,
		e.contractCaller,
		e.chainEventsChan,
		e.config.AvsPerformers,
	)

	chainId := e.config.L1Chain.ChainId
	pollerConfig := &EVMChainPoller.EVMChainPollerConfig{
		ChainId:                 chainId,
		PollingInterval:         time.Duration(e.config.L1Chain.PollIntervalSeconds) * time.Second,
		EigenLayerCoreContracts: e.contractStore.ListContractAddressesForChain(chainId),
		InterestingContracts:    []string{artifactRegistryAddress},
	}

	// Create chain poller
	e.chainPoller = EVMChainPoller.NewEVMChainPoller(
		e.ethereumClient,
		e.chainEventsChan,
		e.transactionLogParser,
		pollerConfig,
		e.logger,
	)

	e.logger.Sugar().Infow("Created chain poller",
		"chainId", e.config.L1Chain.ChainId,
		"pollingInterval", pollerConfig.PollingInterval,
	)

	// Initialize container manager
	containerMgr, err := containerManager.NewDockerContainerManager(e.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize container manager: %w", err)
	}

	// Create the performer pool manager
	e.performerPoolManager = performerPoolManager.NewPerformerPoolManager(
		e.config,
		e.logger,
		e.peeringFetcher,
		e.capacityPlanner,
		containerMgr,
	)

	// Initialize the performer pool manager
	if err = e.performerPoolManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize performer pool manager: %w", err)
	}

	// Register GRPC handlers
	if err = e.registerHandlers(e.rpcServer.GetGrpcServer()); err != nil {
		return fmt.Errorf("failed to register handlers: %w", err)
	}

	e.logger.Sugar().Infow("Executor initialized successfully")
	return nil
}

func (e *Executor) initializeContractCaller(artifactRegistryAddress string) (*caller.ContractCaller, error) {
	e.logger.Sugar().Infow("Initializing contract caller...")

	chain := e.config.L1Chain

	// Get TaskMailbox address from contract store
	taskMailboxContract := util.Find(e.contractStore.ListContracts(), func(c *contracts.Contract) bool {
		return c.ChainId == chain.ChainId && c.Name == config.ContractName_TaskMailbox
	})
	if taskMailboxContract == nil {
		return nil, fmt.Errorf("TaskMailbox contract not found for chain %d", chain.ChainId)
	}

	// Get contract addresses
	ethereumContractCaller, err := e.ethereumClient.GetEthereumContractCaller()
	if err != nil {
		e.logger.Sugar().Errorw("Failed to get ethereum contract caller", "error", err)
		return nil, err
	}

	// Create contract caller
	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:              e.config.Operator.OperatorPrivateKey,
		AVSRegistrarAddress:     e.config.AvsPerformers[0].AVSRegistrarAddress, // Using first AVS performer's registrar
		ArtifactRegistryAddress: artifactRegistryAddress,
		TaskMailboxAddress:      taskMailboxContract.Address,
	}, ethereumContractCaller, e.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract caller: %w", err)
	}

	e.logger.Sugar().Infow("Initialized contract caller for chain",
		"chainId", chain.ChainId,
		"chainName", chain.Name,
		"taskMailboxAddress", taskMailboxContract.Address,
		"artifactRegistryAddress", artifactRegistryAddress,
	)

	return cc, nil
}

func (e *Executor) startChainPoller(ctx context.Context) {
	e.logger.Sugar().Infow("Starting chain poller")
	go func() {
		err := e.chainPoller.Start(ctx)
		if err != nil {
			e.logger.Sugar().Errorw("Chain poller encountered an error", "error", err)
		}
	}()
}

func (e *Executor) Start(ctx context.Context) error {
	e.logger.Info("Executor is starting",
		zap.String("version", "1.0.0"),
		zap.String("operatorAddress", e.config.Operator.Address),
	)
	// Start the capacity planner's event processor
	e.capacityPlanner.Start(ctx)

	// Start the pool manager to run performers
	e.performerPoolManager.Start(ctx)

	// Start chain poller
	e.startChainPoller(ctx)

	if err := e.rpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start RPC server: %v", err)
	}
	return nil
}

func (e *Executor) registerHandlers(grpcServer *grpc.Server) error {
	executorV1.RegisterExecutorServiceServer(grpcServer, e)

	return nil
}

// GetChainEventsChan returns the chain events channel for external components to send events
func (e *Executor) GetChainEventsChan() chan<- *chainPoller.LogWithBlock {
	return e.chainEventsChan
}
