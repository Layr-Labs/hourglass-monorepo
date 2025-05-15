package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulations/peers"
	"slices"
	"strconv"
	"strings"
	"time"

	aggregatorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle/runnable"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulations/executor/service"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
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

		privateSigningKey, err := storedKeys.GetBN254PrivateKey(Config.Operator.SigningKeys.BLS.Password)
		if err != nil {
			return fmt.Errorf("failed to get private key: %w", err)
		}

		sig := inMemorySigner.NewInMemorySigner(privateSigningKey)

		// load the contracts and create the store
		var coreContracts []*contracts.Contract
		if len(Config.Contracts) > 0 {
			log.Sugar().Infow("Loading core contracts from runtime config")
			coreContracts, err = eigenlayer.LoadContractsFromRuntime(string(Config.Contracts))
			if err != nil {
				return fmt.Errorf("failed to load core contracts from runtime: %w", err)
			}
		} else {
			log.Sugar().Infow("Loading core contracts from embedded config")
			coreContracts, err = eigenlayer.LoadContracts()
			if err != nil {
				return fmt.Errorf("failed to load core contracts: %w", err)
			}
		}

		imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, log)

		tlp := transactionLogParser.NewTransactionLogParser(imContractStore, log)

		sugar.Infof("Aggregator config: %+v\n", Config)
		sugar.Infow("Building aggregator components...")

		var pdf *localPeeringDataFetcher.LocalPeeringDataFetcher
		if Config.SimulationConfig.SimulatePeering.Enabled {
			simulatedPeers, err := peers.NewSimulatedPeersFromConfig(Config.SimulationConfig.SimulatePeering.OperatorPeers)
			if err != nil {
				log.Sugar().Fatalw("Failed to create simulated peers", zap.Error(err))
			}

			pdf = localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
				OperatorPeers: simulatedPeers,
			}, log)
		} else {
			return fmt.Errorf("peering data fetcher not implemented")
		}

		if Config.SimulationConfig.SimulateExecutors {
			log.Sugar().Infow("Loading simulated executors from runtime config")
			c := &aggregatorConfig.AggregatorConfig{
				Avss:             Config.Avss,
				Chains:           Config.Chains,
				Operator:         Config.Operator,
				ServerConfig:     Config.ServerConfig,
				SimulationConfig: Config.SimulationConfig,
				L1ChainId:        Config.L1ChainId,
			}
			executors, err := buildSimulatedExecutors(context.Background(), c, log)
			if err != nil {
				return fmt.Errorf("failed to build executors: %w", err)
			}
			for _, executor := range executors {
				err := executor.Start(context.Background())
				if err != nil {
					return err
				}
			}
		}
		agg, err := aggregator.NewAggregatorWithRpcServer(
			Config.ServerConfig.Port,
			&aggregator.AggregatorConfig{
				AVSs:              Config.Avss,
				Chains:            Config.Chains,
				Address:           Config.Operator.Address,
				PrivateKey:        Config.Operator.OperatorPrivateKey,
				AggregatorUrl:     Config.ServerConfig.AggregatorUrl,
				WriteDelaySeconds: time.Duration(Config.SimulationConfig.WriteDelaySeconds) * time.Second,
			},
			imContractStore,
			tlp,
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

func buildSimulatedExecutors(ctx context.Context, cfg *aggregatorConfig.AggregatorConfig, logger *zap.Logger) ([]runnable.IRunnable, error) {
	var executors []runnable.IRunnable
	aggregatorUrl := fmt.Sprintf("localhost:%d", cfg.ServerConfig.Port)
	var allocatedPorts []int

	for _, peer := range cfg.SimulationConfig.SimulatePeering.OperatorPeers {
		addrParts := strings.Split(peer.NetworkAddress, ":")
		if len(addrParts) < 2 {
			return nil, fmt.Errorf("invalid network address format: %s", peer.NetworkAddress)
		}
		port, err := strconv.Atoi(addrParts[len(addrParts)-1])
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", addrParts[len(addrParts)-1])
		}
		if slices.Contains(allocatedPorts, port) {
			return nil, fmt.Errorf("port %d is already allocated", port)
		}

		clientConn, err := clients.NewGrpcClient(aggregatorUrl, false)
		if err != nil {
			logger.Sugar().Fatalw("Failed to create aggregator client", "error", err)
			return nil, err
		}

		aggregatorClient := aggregatorpb.NewAggregatorServiceClient(clientConn)
		exe, err := service.NewSimulatedExecutorWithRpcServer(port, logger, aggregatorClient, peer.OperatorAddress)
		if err != nil {
			logger.Sugar().Fatalw("Failed to create simulated executor", "error", err)
			return nil, err
		}

		executors = append(executors, exe)
		allocatedPorts = append(allocatedPorts, port)

		logger.Sugar().Infow("Created simulated executor",
			zap.String("publicKey", peer.PublicKey),
			zap.Int("port", port),
		)
	}

	return executors, nil
}
