package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"slices"
	"time"

	aggregatorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle/runnable"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainWriter/simulatedChainWriter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/fetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulations/executor/service"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the aggregator",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)
		log, _ := logger.NewLogger(&logger.LoggerConfig{Debug: Config.Debug})
		sugar := log.Sugar()

		if err := Config.Validate(); err != nil {
			sugar.Errorw("Invalid configuration", "error", err)
			return err
		}

		// Load up the keystore
		storedKeys, err := keystore.ParseKeystoreJSON(Config.Operator.SigningKeys.BLS.Keystore)
		if err != nil {
			return fmt.Errorf("failed to parse keystore JSON: %w", err)
		}

		keyScheme, err := keystore.GetSigningSchemeForCurveType(storedKeys.CurveType)
		if err != nil {
			return fmt.Errorf("failed to get signing scheme: %w", err)
		}

		privateSigningKey, err := storedKeys.GetPrivateKey(Config.Operator.SigningKeys.BLS.Password, keyScheme)
		if err != nil {
			return fmt.Errorf("failed to get private key: %w", err)
		}

		sig := inMemorySigner.NewInMemorySigner(privateSigningKey)

		// load the contracts and create the store
		// TODO: configure which L1 chain to use
		coreContracts, err := eigenlayer.LoadCoreContractsForL1Chain(config.ChainId_EthereumHolesky)
		if err != nil {
			return fmt.Errorf("failed to load core contracts: %w", err)
		}

		imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, log)

		tlp := transactionLogParser.NewTransactionLogParser(imContractStore, log)

		sugar.Infof("Aggregator config: %+v\n", Config)
		sugar.Infow("Building aggregator components...")

		var pdf *fetcher.LocalPeeringDataFetcher
		if Config.SimulationConfig.SimulatePeering.Enabled {
			pdf = fetcher.NewLocalPeeringDataFetcher(&fetcher.LocalPeeringDataFetcherConfig{
				OperatorPeers: util.Map(Config.SimulationConfig.SimulatePeering.OperatorPeers, func(p config.SimulatedPeer, i uint64) *peering.OperatorPeerInfo {
					return &peering.OperatorPeerInfo{
						OperatorAddress: p.OperatorAddress,
						Port:            p.Port,
						PublicKey:       p.PublicKey,
						OperatorSetId:   p.OperatorSetId,
						NetworkAddress:  p.NetworkAddress,
					}
				}),
			}, log)
		} else {
			return fmt.Errorf("peering data fetcher not implemented")
		}

		agg, err := aggregator.NewAggregatorWithRpcServer(
			Config.ServerConfig.Port,
			&aggregator.AggregatorConfig{
				AVSs:   Config.Avss,
				Chains: Config.Chains,
			},
			imContractStore,
			tlp,
			nil,
			pdf,
			sig,
			log,
		)
		if err != nil {
			return fmt.Errorf("failed to create aggregator: %w", err)
		}

		if err := agg.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize aggregator: %w", err)
		}

		ctx, cancel := context.WithCancel(cmd.Context())

		go func() {
			if err := agg.Start(ctx); err != nil {
				cancel()
			}
		}()

		gracefulShutdownNotifier := shutdown.CreateGracefulShutdownChannel()
		done := make(chan bool)
		shutdown.ListenForShutdown(gracefulShutdownNotifier, done, func() {
			log.Sugar().Info("Shutting down...")
			cancel()
		}, time.Second*5, log)
		return nil
	},
}

func initRunCmd(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := viper.BindPFlag(f.Name, f); err != nil {
			fmt.Printf("Failed to bind flag '%s': %+v\n", f.Name, err)
		}
		if err := viper.BindEnv(f.Name); err != nil {
			fmt.Printf("Failed to bind env '%s': %+v\n", f.Name, err)
		}
	})
}

/*
func buildChainPollers(
	cfg *aggregatorConfig.AggregatorConfig,
	taskQueue chan *types.Task,
	logger *zap.Logger,
) ([]chainPoller.IChainPoller, error) {
	var listeners []chainPoller.IChainPoller

	for i, chain := range cfg.Chains {
		if cfg.SimulationConfig.Enabled {
			port := cfg.SimulationConfig.PollerPort + i + 1

			var listener chainPoller.IChainPoller
			if cfg.SimulationConfig.AutomaticPoller {
				listenerConfig := &simulatedChainPoller.SimulatedChainPollerConfig{
					ChainId:      &chain.ChainId,
					Port:         port,
					TaskInterval: 250 * time.Millisecond,
				}

				listener = simulatedChainPoller.NewSimulatedChainPoller(
					taskQueue,
					listenerConfig,
					logger,
				)
			} else {
				listenerConfig := &manualPushChainPoller.ManualPushChainPollerConfig{
					ChainId: &chain.ChainId,
					Port:    port,
				}

				listener = manualPushChainPoller.NewManualPushChainPoller(
					taskQueue,
					listenerConfig,
					logger,
				)
			}

			listeners = append(listeners, listener)
			logger.Sugar().Infow("Created simulated chain listener", "chainId", chain.ChainId, "port", port)
		} else {
			abiEntries, err := contracts.GetChainAbis(chain.ChainId, inboxContractName)
			if err != nil {
				logger.Sugar().Fatalw("Failed to load contract address", "error", err)
				return nil, err
			}
			for _, entry := range abiEntries {
				logParser := transactionLogParser.NewTransactionLogParser(logger, nil)
				ethClient := ethereum.NewClient(&ethereum.EthereumClientConfig{
					BaseUrl:   chain.RpcURL,
					BlockType: ethereum.BlockType_Latest,
				}, logger)

				listenerConfig := EVMChainPoller.NewEVMChainPollerDefaultConfig(
					chain.ChainId,
					entry.Address.String(),
				)

				listener := EVMChainPoller.NewEVMChainPoller(
					ethClient,
					taskQueue,
					logParser,
					listenerConfig,
					logger,
				)
				listeners = append(listeners, listener)
				logger.Sugar().Infow("Created Ethereum chain listener", "chainId", chain.ChainId, "url", chain.RpcURL)
			}
		}
	}

	return listeners, nil
}*/

//nolint:unused
func buildWriters(resultQueue chan *types.TaskResult, logger *zap.Logger) []runnable.IRunnable {
	// TODO: implement ponosChainWriter and use when appropriate
	writerConfig := &simulatedChainWriter.SimulatedChainWriterConfig{
		Interval: 1 * time.Millisecond,
	}

	writer := simulatedChainWriter.NewSimulatedChainWriter(
		writerConfig,
		resultQueue,
		logger,
	)

	logger.Sugar().Infow("Created simulated chain writer")
	return []runnable.IRunnable{writer}
}

//nolint:unused
func buildSimulatedExecutors(ctx context.Context, cfg *aggregatorConfig.AggregatorConfig, logger *zap.Logger) ([]runnable.IRunnable, error) {
	var executors []runnable.IRunnable
	aggregatorUrl := fmt.Sprintf("localhost:%d", cfg.ServerConfig.Port)
	var allocatedPorts []int

	for _, peer := range cfg.SimulationConfig.SimulatePeering.OperatorPeers {
		if slices.Contains(allocatedPorts, peer.Port) {
			return nil, fmt.Errorf("port %d is already allocated", peer.Port)
		}

		rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{GrpcPort: peer.Port}, logger)
		if err != nil {
			logger.Sugar().Fatalw("Failed to create rpcServer for executor", "error", err)
			return nil, err
		}

		clientConn, err := clients.NewGrpcClient(aggregatorUrl, false)
		if err != nil {
			logger.Sugar().Fatalw("Failed to create aggregator client", "error", err)
			return nil, err
		}

		aggregatorClient := aggregatorpb.NewAggregatorServiceClient(clientConn)
		exe := service.NewSimulatedExecutorServer(rpc, aggregatorClient, peer.PublicKey)
		executorpb.RegisterExecutorServiceServer(rpc.GetGrpcServer(), exe)
		err = rpc.Start(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to start executor: %w", err)
		}

		executors = append(executors, exe)
		allocatedPorts = append(allocatedPorts, peer.Port)

		logger.Sugar().Infow("Created simulated executor",
			zap.String("publicKey", peer.PublicKey),
			zap.Int("port", peer.Port),
		)
	}

	return executors, nil
}
