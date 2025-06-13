package aggregator

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/avsExecutionManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"time"
)

type AggregatorConfig struct {
	Address       string
	AggregatorUrl string
	PrivateKey    string
	AVSs          []*aggregatorConfig.AggregatorAvs
	Chains        []*aggregatorConfig.Chain
	L1ChainId     config.ChainId
}

type Aggregator struct {
	logger    *zap.Logger
	rpcServer *rpcServer.RpcServer
	config    *AggregatorConfig

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

	// signer is used to sign transactions and communicate securely with executors
	// Since we're only using bn254 at the moment, we only need one signer.
	// In the future, this should be passed to the executionManager based on which
	// curve it requires.
	signer signer.ISigner

	// chainEventsChan is a channel for receiving events from the chain pollers and
	// sequentially processing them
	chainEventsChan chan *chainPoller.LogWithBlock
}

func NewAggregatorWithRpcServer(
	rpcPort int,
	cfg *AggregatorConfig,
	contractStore contractStore.IContractStore,
	tlp *transactionLogParser.TransactionLogParser,
	peeringDataFetcher peering.IPeeringDataFetcher,
	signer signer.ISigner,
	logger *zap.Logger,
) (*Aggregator, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: rpcPort,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %w", err)
	}

	return NewAggregator(rpc, cfg, contractStore, tlp, peeringDataFetcher, signer, logger)
}

func NewAggregator(
	rpcServer *rpcServer.RpcServer,
	cfg *AggregatorConfig,
	contractStore contractStore.IContractStore,
	tlp *transactionLogParser.TransactionLogParser,
	peeringDataFetcher peering.IPeeringDataFetcher,
	signer signer.ISigner,
	logger *zap.Logger,
) (*Aggregator, error) {
	if cfg.L1ChainId == 0 {
		return nil, fmt.Errorf("L1ChainId must be set in AggregatorConfig")
	}
	agg := &Aggregator{
		rpcServer:            rpcServer,
		contractStore:        contractStore,
		transactionLogParser: tlp,
		config:               cfg,
		logger:               logger,
		signer:               signer,
		peeringDataFetcher:   peeringDataFetcher,
		chainContractCallers: make(map[config.ChainId]contractCaller.IContractCaller),
		chainPollers:         make(map[config.ChainId]chainPoller.IChainPoller),
		chainEventsChan:      make(chan *chainPoller.LogWithBlock, 10000),
		avsExecutionManagers: make(map[string]*avsExecutionManager.AvsExecutionManager),
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
		aem, err := avsExecutionManager.NewAvsExecutionManager(&avsExecutionManager.AvsExecutionManagerConfig{
			AvsAddress: avs.Address,
			SupportedChainIds: util.Map(avs.ChainIds, func(id uint, i uint64) config.ChainId {
				return config.ChainId(id)
			}),
			MailboxContractAddresses: util.Reduce(avs.ChainIds, func(acc map[config.ChainId]string, chainId uint) map[config.ChainId]string {
				cId := config.ChainId(chainId)
				chainTaskMailbox := util.Find(loadedContracts, func(c *contracts.Contract) bool {
					return c.Name == config.ContractName_TaskMailbox && c.ChainId == cId
				})
				if chainTaskMailbox == nil {
					a.logger.Sugar().Warnw("TaskMailbox contract not found for chain",
						zap.Uint64("chainId", uint64(cId)),
					)
					return acc
				}

				acc[cId] = chainTaskMailbox.Address
				return acc
			}, make(map[config.ChainId]string)),
			AggregatorAddress: a.config.Address,
			L1ChainId:         a.config.L1ChainId,
		},
			a.chainContractCallers,
			a.signer,
			a.peeringDataFetcher,
			a.logger,
		)
		if err != nil {
			return fmt.Errorf("failed to create AVS Execution Manager for %s: %w", avs.Address, err)
		}

		a.avsExecutionManagers[avs.Address] = aem
	}
	return nil
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

		pCfg := &EVMChainPoller.EVMChainPollerConfig{
			ChainId:              chain.ChainId,
			PollingInterval:      time.Duration(chain.PollIntervalSeconds) * time.Second,
			InterestingContracts: a.contractStore.ListContractAddressesForChain(chain.ChainId),
		}

		a.chainPollers[chain.ChainId] = EVMChainPoller.NewEVMChainPoller(ec, a.chainEventsChan, a.transactionLogParser, pCfg, a.logger)
	}
	return nil
}

func InitializeContractCaller(
	chain *aggregatorConfig.Chain,
	privateKey string,
	contractStore contractStore.IContractStore,
	avsRegistrarAddress string,
	isL1Chain bool,
	logger *zap.Logger,
) (contractCaller.IContractCaller, error) {
	ec := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   chain.RpcURL,
		BlockType: ethereum.BlockType_Latest,
	}, logger)

	mailboxContract := util.Find(contractStore.ListContracts(), func(c *contracts.Contract) bool {
		return c.ChainId == chain.ChainId && c.Name == config.ContractName_TaskMailbox
	})
	if mailboxContract == nil {
		logger.Sugar().Errorw("Mailbox contract not found",
			zap.Uint64("chainId", uint64(chain.ChainId)),
		)
		return nil, fmt.Errorf("mailbox contract not found for chain %s", chain.Name)
	}
	mailboxContractAddress := mailboxContract.Address

	ethereumContractCaller, err := ec.GetEthereumContractCaller()
	if err != nil {
		logger.Sugar().Errorw("failed to get ethereum contract caller", "error", err)
		return nil, err
	}

	callerConfig := &caller.ContractCallerConfig{
		PrivateKey:          privateKey,
		AVSRegistrarAddress: avsRegistrarAddress,
		TaskMailboxAddress:  mailboxContractAddress,
	}
	if isL1Chain {
		protocolContractAddresses, err := config.GetCoreContractsForChainId(chain.ChainId)
		if err != nil {
			logger.Sugar().Errorw("failed to get protocol contract addresses", "error", err)
			return nil, fmt.Errorf("failed to get protocol contract addresses for chain %s: %w", chain.Name, err)
		}
		callerConfig.KeyRegistrarAddress = protocolContractAddresses.KeyRegistrar
		callerConfig.CrossChainRegistryAddress = protocolContractAddresses.CrossChainRegistry
	}

	return caller.NewContractCaller(callerConfig, ethereumContractCaller, logger)
}

func (a *Aggregator) initializeContractCallers() (map[config.ChainId]contractCaller.IContractCaller, error) {
	a.logger.Sugar().Infow("Initializing contract callers...")
	contractCallers := make(map[config.ChainId]contractCaller.IContractCaller)
	for _, chain := range a.config.Chains {

		cc, err := InitializeContractCaller(
			chain,
			a.config.PrivateKey,
			a.contractStore,
			a.config.AVSs[0].AVSRegistrarAddress,
			chain.ChainId == a.config.L1ChainId,
			a.logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize contract caller for chain %s: %w", chain.Name, err)
		}
		contractCallers[chain.ChainId] = cc
	}
	return contractCallers, nil
}

// Start starts the aggregator and its components
func (a *Aggregator) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	// start the RPC server
	go func() {
		if err := a.rpcServer.Start(ctx); err != nil {
			a.logger.Sugar().Errorw("RPC server failed to start", "error", err)
			cancel()
		}
	}()

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
