package aggregator

import (
	"context"
	"fmt"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/avsExecutionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"
	"time"
)

type AggregatorConfig struct {
	Address          string
	PrivateKeyConfig *config.ECDSAKeyConfig
	AVSs             []*aggregatorConfig.AggregatorAvs
	Chains           []*aggregatorConfig.Chain
	L1ChainId        config.ChainId
}

type Aggregator struct {
	logger *zap.Logger
	config *AggregatorConfig

	// chainPollers is a map of chainId to its chain poller
	chainPollers map[config.ChainId]chainPoller.IChainPoller

	// transactionLogParser is used to decode logs from the chain
	transactionLogParser *transactionLogParser.TransactionLogParser

	// contractStore is used to fetch contract addresses and ABIs
	contractStore contractStore.IContractStore

	// chainContractCallers is a future-proof placeholder for the ContractCaller in another PR
	chainContractCallers map[config.ChainId]contractCaller.IContractCaller

	// avsExecutionManagers map of avsAddress to its AvsExecutionManager
	avsExecutionManagers map[string]*avsExecutionManager.AvsExecutionManager

	// peeringDataFetcher is used to fetch peering data (typically from the L1)
	peeringDataFetcher peering.IPeeringDataFetcher

	signers signer.Signers

	// chainEventsChan is a channel for receiving events from the chain pollers and
	// sequentially processing them
	chainEventsChan chan *chainPoller.LogWithBlock

	managementRpcServer *rpcServer.RpcServer

	// store is the persistence layer for the aggregator
	store storage.AggregatorStore

	// authVerifier handles authentication for management APIs
	authVerifier *auth.Verifier
}

func NewAggregatorWithManagementRpcServer(
	managementServerGrpcPort int,
	cfg *AggregatorConfig,
	contractStore contractStore.IContractStore,
	tlp *transactionLogParser.TransactionLogParser,
	peeringDataFetcher peering.IPeeringDataFetcher,
	signers signer.Signers,
	store storage.AggregatorStore,
	logger *zap.Logger,
) (*Aggregator, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: managementServerGrpcPort,
	}, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create aggregator management rpc server with port %d", managementServerGrpcPort)
	}

	return NewAggregator(cfg, contractStore, tlp, peeringDataFetcher, signers, store, rpc, logger)
}

func NewAggregator(
	cfg *AggregatorConfig,
	contractStore contractStore.IContractStore,
	tlp *transactionLogParser.TransactionLogParser,
	peeringDataFetcher peering.IPeeringDataFetcher,
	signers signer.Signers,
	store storage.AggregatorStore,
	managementRpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
) (*Aggregator, error) {
	if cfg.L1ChainId == 0 {
		return nil, fmt.Errorf("L1ChainId must be set in AggregatorConfig")
	}
	if store == nil {
		return nil, fmt.Errorf("store is required")
	}
	// Initialize auth verifier for management APIs
	var authVerifier *auth.Verifier
	if signers.ECDSASigner != nil {
		tokenManager := auth.NewChallengeTokenManager(cfg.Address, 5*time.Minute)
		authVerifier = auth.NewVerifier(tokenManager, signers.ECDSASigner)
	}

	agg := &Aggregator{
		contractStore:        contractStore,
		transactionLogParser: tlp,
		config:               cfg,
		logger:               logger,
		signers:              signers,
		peeringDataFetcher:   peeringDataFetcher,
		store:                store,
		chainContractCallers: make(map[config.ChainId]contractCaller.IContractCaller),
		chainPollers:         make(map[config.ChainId]chainPoller.IChainPoller),
		chainEventsChan:      make(chan *chainPoller.LogWithBlock, 10000),
		avsExecutionManagers: make(map[string]*avsExecutionManager.AvsExecutionManager),
		managementRpcServer:  managementRpcServer,
		authVerifier:         authVerifier,
	}
	return agg, nil
}

// Initialize sets up chain pollers and AVSExecutionManagers
func (a *Aggregator) Initialize() error {
	if err := a.initializePollers(); err != nil {
		return fmt.Errorf("failed to initialize pollers: %w", err)
	}

	callers, err := a.initializeContractCallers()
	if err != nil {
		return fmt.Errorf("failed to initialize contract callers: %w", err)
	}
	a.chainContractCallers = callers

	loadedContracts := a.contractStore.ListContracts()
	for _, c := range loadedContracts {
		a.logger.Sugar().Infow("Loaded contract",
			zap.String("name", c.Name),
			zap.String("address", c.Address),
			zap.Uint64("chainId", uint64(c.ChainId)),
		)
	}

	for _, avs := range a.config.AVSs {
		if err := a.registerAvs(avs); err != nil {
			return fmt.Errorf("failed to register AVS %s: %w", avs.Address, err)
		}
	}

	a.registerHandlers()

	return nil
}

func (a *Aggregator) registerAvs(avs *aggregatorConfig.AggregatorAvs) error {
	if _, ok := a.avsExecutionManagers[avs.Address]; ok {
		return fmt.Errorf("AVS Execution Manager for %s already exists", avs.Address)
	}
	supportedChains, err := a.getValidChainsForAvs(avs.Address)
	if err != nil {
		return fmt.Errorf("failed to get valid chains for AVS %s: %w", avs.Address, err)
	}

	om := operatorManager.NewOperatorManager(&operatorManager.OperatorManagerConfig{
		AvsAddress: avs.Address,
		ChainIds:   supportedChains,
		L1ChainId:  a.config.L1ChainId,
	}, a.chainContractCallers, a.peeringDataFetcher, a.logger)

	aem, err := avsExecutionManager.NewAvsExecutionManager(&avsExecutionManager.AvsExecutionManagerConfig{
		AvsAddress:               avs.Address,
		SupportedChainIds:        supportedChains,
		MailboxContractAddresses: getMailboxAddressesForChains(a.contractStore.ListContracts()),
		L1ChainId:                a.config.L1ChainId,
		AggregatorAddress:        a.config.Address,
	},
		a.chainContractCallers,
		a.signers,
		a.contractStore,
		om,
		a.store,
		a.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create AVS Execution Manager for %s: %w", avs.Address, err)
	}

	a.avsExecutionManagers[avs.Address] = aem
	return nil
}

func getMailboxAddressesForChains(allContracts []*contracts.Contract) map[config.ChainId]string {
	return util.Reduce(allContracts, func(acc map[config.ChainId]string, c *contracts.Contract) map[config.ChainId]string {
		if c.Name != config.ContractName_TaskMailbox {
			return acc
		}
		acc[c.ChainId] = c.Address
		return acc
	}, map[config.ChainId]string{})
}

// getValidChainsForAvs returns the valid chain IDs for a given AVS address
// for now, this is a stub for eventually calling on-chain to get the valid chains
func (a *Aggregator) getValidChainsForAvs(avsAddress string) ([]config.ChainId, error) {
	avs := util.Find(a.config.AVSs, func(avs *aggregatorConfig.AggregatorAvs) bool {
		return strings.EqualFold(avs.Address, avsAddress)
	})
	if avs == nil {
		return nil, fmt.Errorf("AVS with address %s not found in config", avsAddress)
	}
	return util.Map(avs.ChainIds, func(id uint, i uint64) config.ChainId {
		return config.ChainId(id)
	}), nil
}

func (a *Aggregator) initializePollers() error {
	a.logger.Sugar().Infow("Initializing chain pollers...",
		zap.Any("chains", a.config.Chains),
	)

	for _, chain := range a.config.Chains {
		if _, ok := a.chainPollers[chain.ChainId]; ok {
			a.logger.Sugar().Warnw("L1Chain poller already exists for chain", "chainId", chain.ChainId)
			continue
		}
		ec := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
			BaseUrl:   chain.RpcURL,
			BlockType: ethereum.BlockType_Latest,
		}, a.logger)

		pollInterval := chain.PollIntervalSeconds
		if pollInterval <= 0 {
			a.logger.Sugar().Warnw("Invalid poll interval for chain", "chainId", chain.ChainId, "pollInterval", pollInterval)
			pollInterval = 10 // default to 10 seconds if not set or invalid
		}

		pCfg := &EVMChainPoller.EVMChainPollerConfig{
			ChainId:              chain.ChainId,
			PollingInterval:      time.Duration(pollInterval) * time.Second,
			InterestingContracts: a.contractStore.ListContractAddressesForChain(chain.ChainId),
		}

		a.chainPollers[chain.ChainId] = EVMChainPoller.NewEVMChainPoller(ec, a.chainEventsChan, a.transactionLogParser, pCfg, a.store, a.logger)
	}
	return nil
}

func InitializeContractCaller(
	chain *aggregatorConfig.Chain,
	privateKeyConfig *config.ECDSAKeyConfig,
	logger *zap.Logger,
) (contractCaller.IContractCaller, error) {
	ec := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   chain.RpcURL,
		BlockType: ethereum.BlockType_Latest,
	}, logger)

	ethereumContractCaller, err := ec.GetEthereumContractCaller()
	if err != nil {
		logger.Sugar().Errorw("failed to get ethereum contract caller", "error", err)
		return nil, err
	}

	txSigner, err := transactionSigner.NewTransactionSigner(privateKeyConfig, ethereumContractCaller, logger)
	if err != nil {
		logger.Sugar().Errorw("failed to create transactionSigner", "error", err)
		return nil, fmt.Errorf("failed to create transactionSigner: %w", err)
	}

	return caller.NewContractCaller(ethereumContractCaller, txSigner, logger)
}

func (a *Aggregator) initializeContractCallers() (map[config.ChainId]contractCaller.IContractCaller, error) {
	a.logger.Sugar().Infow("Initializing contract callers...")
	contractCallers := make(map[config.ChainId]contractCaller.IContractCaller)
	for _, chain := range a.config.Chains {

		cc, err := InitializeContractCaller(chain, a.config.PrivateKeyConfig, a.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize contract caller for chain %s: %w", chain.Name, err)
		}
		contractCallers[chain.ChainId] = cc
	}
	if _, ok := contractCallers[a.config.L1ChainId]; !ok {
		return nil, fmt.Errorf("no contract caller initialized for L1 chain %d", a.config.L1ChainId)
	}
	return contractCallers, nil
}

// Start starts the aggregator and its components
func (a *Aggregator) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	// consume the events channel
	go func() {
		if err := a.processEventsChan(ctx); err != nil {
			a.logger.Sugar().Errorw("Aggregator failed to process events channel", "error", err)
			cancel()
		}
	}()

	// run execution managers
	for _, avsExec := range a.avsExecutionManagers {
		go func(avsExec *avsExecutionManager.AvsExecutionManager) {
			if err := avsExec.Init(ctx); err != nil {
				a.logger.Sugar().Errorw("AVS Execution Manager failed to initialize", "error", err)
				cancel()
			}
			if err := avsExec.Start(ctx); err != nil {
				a.logger.Sugar().Errorw("AVS Execution Manager failed to start", "error", err)
				cancel()
			}
		}(avsExec)
	}
	a.logger.Sugar().Infow("Execution managers started")

	// start polling for blocks
	for _, poller := range a.chainPollers {
		a.logger.Sugar().Infow("Starting chain poller", "poller", poller)
		if err := poller.Start(ctx); err != nil {
			a.logger.Sugar().Errorw("L1Chain poller failed to start", "error", err)
			cancel()
		}
	}

	if err := a.managementRpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start management RPC server: %v", err)
	}

	<-ctx.Done()
	a.logger.Sugar().Infow("Aggregator context done, stopping")
	return nil
}

func (a *Aggregator) processEventsChan(ctx context.Context) error {
	a.logger.Sugar().Infow("Starting to process events channel...")
	for {
		select {
		case <-ctx.Done():
			a.logger.Sugar().Info("Aggregator context done, stopping event processing")
			return nil
		case logWithBlock := <-a.chainEventsChan:
			if err := a.processLog(logWithBlock); err != nil {
				a.logger.Sugar().Errorw("Error processing log", "error", err)
				return err
			}
		}
	}
}

func (a *Aggregator) processLog(lwb *chainPoller.LogWithBlock) error {
	a.logger.Sugar().Debugw("Processing log",
		zap.String("eventName", lwb.Log.EventName),
		zap.Any("lwb", lwb),
	)
	for _, avs := range a.avsExecutionManagers {
		if err := avs.HandleLog(lwb); err != nil {
			a.logger.Error("Error processing log in AVS Execution Manager", zap.Error(err))
			return err
		}
	}
	return nil
}

func (a *Aggregator) registerHandlers() {
	aggregatorV1.RegisterAggregatorManagementServiceServer(a.managementRpcServer.GetGrpcServer(), a)
}
