package aggregator

import (
	"context"
	"fmt"
	"sync"
	"time"

	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/avsExecutionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultPollIntervalSeconds = 10
)

type AggregatorConfig struct {
	Address          string
	PrivateKeyConfig *config.ECDSAKeyConfig
	AVSs             []*aggregatorConfig.AggregatorAvs
	Chains           []*aggregatorConfig.Chain
	L1ChainId        config.ChainId
	Authentication   *auth.Config
	TLSEnabled       bool
}

// AvsExecutionManagerInfo encapsulates all information related to a running AVS
type AvsExecutionManagerInfo struct {
	Address          string
	ExecutionManager *avsExecutionManager.AvsExecutionManager
	CancelFunc       context.CancelFunc
}

type Aggregator struct {
	logger *zap.Logger
	config *AggregatorConfig

	// rootCtx is the root context for the aggregator lifetime, used for creating child contexts for AVSs
	// This ensures AVSs registered via API calls have proper lifecycle management tied to aggregator shutdown
	rootCtx context.Context

	// transactionLogParser is used to decode logs from the chain
	transactionLogParser *transactionLogParser.TransactionLogParser

	// contractStore is used to fetch contract addresses and ABIs
	contractStore contractStore.IContractStore

	// chainContractCallers is a future-proof placeholder for the ContractCaller in another PR
	chainContractCallers map[config.ChainId]contractCaller.IContractCaller

	// avsMutex protects concurrent access to avsManagers
	avsMutex sync.RWMutex

	// avsManagers map of avsAddress to AvsExecutionManagerInfo containing execution manager and cancel function
	avsManagers map[string]*AvsExecutionManagerInfo

	// peeringDataFetcher is used to fetch peering data (typically from the L1)
	peeringDataFetcher peering.IPeeringDataFetcher

	signers signer.Signers

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

	authVerifier := getAuthVerifier(cfg, signers, logger)

	return &Aggregator{
		contractStore:        contractStore,
		transactionLogParser: tlp,
		config:               cfg,
		logger:               logger,
		signers:              signers,
		peeringDataFetcher:   peeringDataFetcher,
		store:                store,
		chainContractCallers: make(map[config.ChainId]contractCaller.IContractCaller),
		avsManagers:          make(map[string]*AvsExecutionManagerInfo),
		managementRpcServer:  managementRpcServer,
		authVerifier:         authVerifier,
	}, nil
}

func (a *Aggregator) Initialize() error {
	a.logger.Sugar().Infow("Starting aggregator initialization",
		zap.String("address", a.config.Address),
		zap.Int("numAVSs", len(a.config.AVSs)),
		zap.Int("numChains", len(a.config.Chains)),
		zap.Bool("authEnabled", a.authVerifier != nil),
	)

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

	a.registerHandlers()

	a.logger.Sugar().Infow("Aggregator initialization completed successfully",
		zap.Bool("managementServerReady", a.managementRpcServer != nil),
	)

	return nil
}

func (a *Aggregator) registerAvs(avs *aggregatorConfig.AggregatorAvs) error {

	a.avsMutex.RLock()
	if _, ok := a.avsManagers[avs.Address]; ok {
		a.avsMutex.RUnlock()
		return fmt.Errorf("AVS Execution Manager for %s already exists", avs.Address)
	}
	a.avsMutex.RUnlock()

	supportedChains, err := a.getValidChainsForAvs(avs.ChainIds)
	if err != nil {
		return fmt.Errorf("failed to get valid chains for AVS %s: %w", avs.Address, err)
	}

	om := operatorManager.NewOperatorManager(&operatorManager.OperatorManagerConfig{
		AvsAddress: avs.Address,
		ChainIds:   supportedChains,
		L1ChainId:  a.config.L1ChainId,
	}, a.chainContractCallers, a.peeringDataFetcher, a.logger)

	taskQueue := make(chan *types.Task)
	chainPollers := a.getChainPollers(supportedChains, avs.Address, taskQueue)

	avsExeConfig := &avsExecutionManager.AvsExecutionManagerConfig{
		AvsAddress:               avs.Address,
		SupportedChainIds:        supportedChains,
		MailboxContractAddresses: getMailboxAddressesForChains(a.contractStore.ListContracts()),
		L1ChainId:                a.config.L1ChainId,
		AggregatorAddress:        a.config.Address,
<<<<<<< HEAD
		TlsEnabled:               a.config.TLSEnabled,
	},
=======
	}

	aem, err := avsExecutionManager.NewAvsExecutionManager(
		avsExeConfig,
>>>>>>> 0eccd65 (Refactored task handling to EVMChainPoller and making it aware of AvsAddress)
		a.chainContractCallers,
		a.signers,
		a.contractStore,
		om,
		taskQueue,
		chainPollers,
		a.store,
		a.logger,
	)

	if err != nil {
		return fmt.Errorf("failed to create AVS Execution Manager for %s: %w", avs.Address, err)
	}

	// Use root context instead of passed context to ensure proper lifecycle management
	if a.rootCtx == nil {
		return fmt.Errorf("aggregator not started, cannot register AVS %s", avs.Address)
	}

	avsCtx, avsCancel := context.WithCancel(a.rootCtx)

	err = a.storeAvsManager(avs.Address, aem, avsCancel)
	if err != nil {
		avsCancel()
		return err
	}

	err = aem.Start(avsCtx)
	if err != nil {
		a.removeAvsManager(avs.Address)
		avsCancel()
		return fmt.Errorf("failed to start AVS Execution Manager for %s: %w", avs.Address, err)
	}

	a.logger.Sugar().Infow("AVS execution manager started successfully",
		zap.String("avsAddress", avs.Address))

	return nil
}

func (a *Aggregator) deregisterAvs(avsAddress string) error {
	a.logger.Sugar().Infow("Deregistering AVS",
		zap.String("avsAddress", avsAddress),
	)

	a.avsMutex.Lock()

	avsInfo, exists := a.avsManagers[avsAddress]
	if !exists {
		a.avsMutex.Unlock()
		return fmt.Errorf("AVS %s is not registered", avsAddress)
	}

	delete(a.avsManagers, avsAddress)
	cancelFunc := avsInfo.CancelFunc

	a.avsMutex.Unlock()

	if cancelFunc != nil {
		a.logger.Sugar().Infow("Cancelling AVS context to stop execution manager gracefully",
			zap.String("avsAddress", avsAddress),
		)
		cancelFunc()
	} else {
		a.logger.Sugar().Warnw("No cancel function found for AVS, proceeding with removal",
			zap.String("avsAddress", avsAddress),
		)
	}

	a.logger.Sugar().Infow("AVS deregistered successfully",
		zap.String("avsAddress", avsAddress),
	)
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

// getValidChainsForAvs returns the valid chain IDs of the aggregator
// for now, this is a stub for eventually calling on-chain to get the valid chains
func (a *Aggregator) getValidChainsForAvs(chainIds []uint) ([]config.ChainId, error) {
	validChainMap := make(map[uint]bool)
	for _, chain := range a.config.Chains {
		validChainMap[uint(chain.ChainId)] = true
	}

	validChainIds := util.Filter(chainIds, func(id uint) bool {
		return validChainMap[id]
	})

	if len(validChainIds) != len(chainIds) {
		return nil, fmt.Errorf("not all chainIds provided are valid")
	}

	return util.Map(validChainIds, func(id uint, i uint64) config.ChainId {
		return config.ChainId(id)
	}), nil
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

	a.logger.Sugar().Infow("Initializing contract callers")
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

	a.rootCtx = ctx

	for _, avs := range a.config.AVSs {

		a.logger.Sugar().Infow("Registering and starting AVS from config",
			zap.String("address", avs.Address),
			zap.Any("chainIds", avs.ChainIds),
		)

		if err := a.registerAvs(avs); err != nil {
			return fmt.Errorf("failed to register AVS %s: %w", avs.Address, err)
		}

		a.logger.Sugar().Infow("AVS registered and started successfully",
			zap.String("address", avs.Address),
		)
	}

	if err := a.managementRpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start management RPC server: %v", err)
	}
	a.logger.Sugar().Infow("Management RPC server started successfully")

	a.logger.Sugar().Infow("Aggregator fully started and running",
		zap.Bool("authEnabled", a.authVerifier != nil),
	)

	<-ctx.Done()
	a.logger.Sugar().Infow("Aggregator context done, stopping")

	return nil
}

func (a *Aggregator) verifyAuth(auth *commonV1.AuthSignature) error {

	if a.authVerifier == nil {

		if auth != nil {
			a.logger.Sugar().Warnw("Authentication provided but not enabled")
			return status.Error(codes.Unimplemented, "authentication is not enabled")
		}

		a.logger.Sugar().Debugw("Authentication verifier not configured, skipping verification")
		return nil
	}

	if err := a.authVerifier.VerifyAuthentication(auth); err != nil {
		a.logger.Sugar().Warnw("Authentication verification failed",
			zap.Error(err),
		)
		return err
	}
	a.logger.Sugar().Debugw("Authentication verification successful")
	return nil
}

// getAuthVerifier creates and returns an authentication verifier based on configuration
func getAuthVerifier(cfg *AggregatorConfig, signers signer.Signers, logger *zap.Logger) *auth.Verifier {
	if cfg.Authentication == nil {
		logger.Sugar().Infow("No authentication configuration provided, authentication disabled")
		return nil
	}

	logger.Sugar().Infow("Authentication configuration loaded",
		zap.Bool("enabled", cfg.Authentication.IsEnabled),
	)

	if !cfg.Authentication.IsEnabled {
		logger.Sugar().Infow("Authentication is disabled via configuration")
		return nil
	}

	logger.Sugar().Infow("Authentication is enabled, initializing verifier")

	var authSigner signer.ISigner
	if signers.ECDSASigner != nil {
		authSigner = signers.ECDSASigner
		logger.Sugar().Infow("Using ECDSA signer for authentication")
	} else if signers.BLSSigner != nil {
		authSigner = signers.BLSSigner
		logger.Sugar().Infow("Using BLS signer for authentication")
	} else {
		logger.Sugar().Warnw("Authentication enabled but no signer available")
		return nil
	}

	logger.Sugar().Infow("Creating authentication verifier",
		zap.String("address", cfg.Address),
		zap.Duration("tokenExpiry", 5*time.Minute),
	)
	tokenManager := auth.NewChallengeTokenManager(cfg.Address, 5*time.Minute)
	authVerifier := auth.NewVerifier(tokenManager, authSigner)
	logger.Sugar().Infow("Authentication verifier created successfully")

	return authVerifier
}

// getChainPollers creates and returns a map of chain pollers for the given AVS
func (a *Aggregator) getChainPollers(supportedChains []config.ChainId, avsAddress string, taskQueue chan *types.Task) map[config.ChainId]*EVMChainPoller.EVMChainPoller {
	chainPollers := make(map[config.ChainId]*EVMChainPoller.EVMChainPoller)

	for _, chainId := range supportedChains {
		chain := a.getChainConfig(chainId)
		if chain == nil {
			a.logger.Sugar().Warnw("Chain config not found for chainId", "chainId", chainId)
			continue
		}

		ec := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
			BaseUrl:   chain.RpcURL,
			BlockType: ethereum.BlockType_Latest,
		}, a.logger)

		pollInterval := chain.PollIntervalSeconds
		if pollInterval <= 0 {
			a.logger.Sugar().Warnw("Invalid poll interval for chain", "chainId", chainId, "pollInterval", pollInterval)
			pollInterval = defaultPollIntervalSeconds
		}

		pollerConfig := &EVMChainPoller.EVMChainPollerConfig{
			ChainId:              chainId,
			AvsAddress:           avsAddress,
			PollingInterval:      time.Duration(pollInterval) * time.Second,
			InterestingContracts: a.contractStore.ListContractAddressesForChain(chainId),
		}

		poller := EVMChainPoller.NewEVMChainPoller(
			ec,
			taskQueue,
			a.transactionLogParser,
			pollerConfig,
			a.contractStore,
			a.store,
			a.logger,
		)
		chainPollers[chainId] = poller

		a.logger.Sugar().Infow("Created poller for AVS on chain",
			"avsAddress", avsAddress,
			"chainId", chainId)
	}

	return chainPollers
}

// storeAvsManager stores the AVS manager in the map with proper mutex handling
func (a *Aggregator) storeAvsManager(avsAddress string, aem *avsExecutionManager.AvsExecutionManager, cancelFunc context.CancelFunc) error {
	a.avsMutex.Lock()
	defer a.avsMutex.Unlock()

	if _, ok := a.avsManagers[avsAddress]; ok {
		return fmt.Errorf("AVS Execution Manager for %s already exists", avsAddress)
	}

	avsInfo := &AvsExecutionManagerInfo{
		Address:          avsAddress,
		ExecutionManager: aem,
		CancelFunc:       cancelFunc,
	}
	a.avsManagers[avsAddress] = avsInfo

	return nil
}

// removeAvsManager removes the AVS manager from the map with proper mutex handling
func (a *Aggregator) removeAvsManager(avsAddress string) {
	a.avsMutex.Lock()
	defer a.avsMutex.Unlock()

	delete(a.avsManagers, avsAddress)
}

func (a *Aggregator) getChainConfig(chainId config.ChainId) *aggregatorConfig.Chain {
	for _, chain := range a.config.Chains {
		if chain.ChainId == chainId {
			return chain
		}
	}
	return nil
}

func (a *Aggregator) registerHandlers() {
	aggregatorV1.RegisterAggregatorManagementServiceServer(a.managementRpcServer.GetGrpcServer(), a)
}
