package executor

import (
	"context"
	"fmt"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerPoolManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Executor struct {
	logger        *zap.Logger
	config        *executorConfig.ExecutorConfig
	rpcServer     *rpcServer.RpcServer
	signer        signer.ISigner
	inflightTasks *sync.Map

	performerPoolManager *performerPoolManager.PerformerPoolManager
	capacityPlanner      *performerPoolManager.PerformerCapacityPlanner
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
) (*Executor, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: port,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %v", err)
	}

	return NewExecutor(config, rpc, logger, signer, peeringFetcher), nil
}

func NewExecutor(
	config *executorConfig.ExecutorConfig,
	rpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
	signer signer.ISigner,
	peeringFetcher peering.IPeeringDataFetcher,
) *Executor {
	return &Executor{
		logger:          logger,
		config:          config,
		rpcServer:       rpcServer,
		signer:          signer,
		inflightTasks:   &sync.Map{},
		peeringFetcher:  peeringFetcher,
		chainEventsChan: make(chan *chainPoller.LogWithBlock, 10000),
	}
}

func (e *Executor) Initialize() error {
	e.logger.Sugar().Infow("Initializing executor...")

	// Load contracts
	err := e.loadContracts()
	if err != nil {
		return err
	}

	// Initialize transaction log parser
	e.transactionLogParser = transactionLogParser.NewTransactionLogParser(e.contractStore, e.logger)

	// Initialize Ethereum clients
	e.ethereumClient = ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   e.config.Chain.RpcURL,
		BlockType: ethereum.BlockType_Latest,
	}, e.logger)

	// Initialize contract callers
	e.contractCaller, err = e.initializeContractCaller()
	if err != nil {
		return fmt.Errorf("failed to initialize contract callers: %w", err)
	}

	// Initialize capacity planner with contract callers but without chain poller responsibilities
	operatorAddress := e.config.Operator.Address
	e.capacityPlanner = performerPoolManager.NewPerformerCapacityPlanner(
		e.logger,
		operatorAddress,
		e.contractStore,
		e.contractCaller,
	)

	// Register AVS performers with the capacity planner
	for _, performer := range e.config.AvsPerformers {
		e.capacityPlanner.RegisterAVS(performer.AvsAddress, performer)
	}

	// Instead of using the planner to initialize chain pollers, do it directly here
	chainId := e.config.Chain.ChainId
	pollerConfig := &EVMChainPoller.EVMChainPollerConfig{
		ChainId:                 chainId,
		PollingInterval:         time.Duration(e.config.Chain.PollIntervalSeconds) * time.Second,
		EigenLayerCoreContracts: e.contractStore.ListContractAddressesForChain(chainId),
		InterestingContracts:    []string{e.config.AvsArtifactRegistry, e.config.AvsTaskRegistrar},
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
		"chainId", e.config.Chain.ChainId,
		"pollingInterval", pollerConfig.PollingInterval,
	)

	// Create the performer pool manager
	e.performerPoolManager = performerPoolManager.NewPerformerPoolManager(
		e.config,
		e.logger,
		e.peeringFetcher,
		e.capacityPlanner,
	)

	// Initialize the performer pool manager
	if err := e.performerPoolManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize performer pool manager: %w", err)
	}

	// Register GRPC handlers
	if err := e.registerHandlers(e.rpcServer.GetGrpcServer()); err != nil {
		return fmt.Errorf("failed to register handlers: %w", err)
	}

	e.logger.Sugar().Infow("Executor initialized successfully")
	return nil
}

func (e *Executor) loadContracts() error {
	// Load the core contracts
	var coreContracts []*contracts.Contract
	var err error

	if len(e.config.Contracts) > 0 {
		e.logger.Sugar().Infow("Loading core contracts from runtime config")
		coreContracts, err = eigenlayer.LoadContractsFromRuntime(string(e.config.Contracts))
		if err != nil {
			return fmt.Errorf("failed to load core contracts from runtime: %w", err)
		}
	} else {
		e.logger.Sugar().Infow("Loading core contracts from embedded config")
		coreContracts, err = eigenlayer.LoadContracts()
		if err != nil {
			return fmt.Errorf("failed to load core contracts: %w", err)
		}
	}

	// Create the contract store
	e.contractStore = inMemoryContractStore.NewInMemoryContractStore(coreContracts, e.logger)

	// Create the transaction log parser
	e.transactionLogParser = transactionLogParser.NewTransactionLogParser(e.contractStore, e.logger)

	return nil
}

func (e *Executor) initializeContractCaller() (*caller.ContractCaller, error) {
	e.logger.Sugar().Infow("Initializing contract caller...")

	chain := e.config.Chain
	// Get contract addresses
	var (
		avsRegistrarAddress     string
		artifactRegistryAddress string
	)

	if chain.Simulation != nil && chain.Simulation.Enabled {
		// Use simulation addresses
		avsRegistrarAddress = config.AVSRegistrarSimulationAddress
		artifactRegistryAddress = e.config.AvsArtifactRegistry // Use configured address
	} else {
		// Find contracts by name in contract store
		for _, contract := range e.contractStore.ListContracts() {
			if contract.ChainId == chain.ChainId {
				switch contract.Name {
				case "AVSDirectRegistrar":
					avsRegistrarAddress = contract.Address
				case "AVSArtifactRegistry":
					artifactRegistryAddress = contract.Address
				}
			}
		}

		if avsRegistrarAddress == "" {
			e.logger.Sugar().Errorw("AVS registrar contract not found",
				zap.Uint64("chainId", uint64(chain.ChainId)),
			)
			return nil, fmt.Errorf("AVS registrar contract not found for chain %s", chain.Name)
		}

		if artifactRegistryAddress == "" {
			e.logger.Sugar().Warnw("AVS artifact registry contract not found in store, using configured address",
				zap.Uint64("chainId", uint64(chain.ChainId)),
			)
			artifactRegistryAddress = e.config.AvsArtifactRegistry
		}
	}

	ethereumContractCaller, err := e.ethereumClient.GetEthereumContractCaller()
	if err != nil {
		e.logger.Sugar().Errorw("Failed to get ethereum contract caller", "error", err)
		return nil, err
	}

	// Create contract caller
	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:                 e.config.Operator.OperatorPrivateKey,
		AVSRegistrarAddress:        avsRegistrarAddress,
		AVSArtifactRegistryAddress: artifactRegistryAddress,
	}, ethereumContractCaller, e.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract caller: %w", err)
	}

	e.logger.Sugar().Infow("Initialized contract caller for chain",
		"chainId", chain.ChainId,
		"chainName", chain.Name,
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

func (e *Executor) BootPerformers(ctx context.Context) error {
	// Boot the performers
	return e.performerPoolManager.BootPerformers(ctx)
}

func (e *Executor) Start(ctx context.Context) error {
	e.logger.Info("Executor is starting",
		zap.String("version", "1.0.0"),
		zap.String("operatorAddress", e.config.Operator.Address),
	)
	// Start the capacity planner's event processor
	e.capacityPlanner.Start(ctx)

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
