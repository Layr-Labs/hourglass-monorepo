package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/simulations/peers"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
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
		var err error
		var storedKeys *keystore.EIP2335Keystore
		if Config.Operator.SigningKeys.BLS.Keystore != "" {
			storedKeys, err = keystore.ParseKeystoreJSON(Config.Operator.SigningKeys.BLS.Keystore)
			if err != nil {
				return fmt.Errorf("failed to parse keystore JSON: %w", err)
			}
		} else {
			storedKeys, err = keystore.LoadKeystoreFile(Config.Operator.SigningKeys.BLS.KeystoreFile)
			if err != nil {
				return fmt.Errorf("failed to load keystore file: '%s' %w", Config.Operator.SigningKeys.BLS.KeystoreFile, err)
			}
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

		// Allow overriding contracts from the runtime config
		if Config.OverrideContracts != nil {
			if Config.OverrideContracts.TaskMailbox != nil && len(Config.OverrideContracts.TaskMailbox.Contract) > 0 {
				overrideContract, err := eigenlayer.LoadOverrideContract(Config.OverrideContracts.TaskMailbox.Contract)
				if err != nil {
					return fmt.Errorf("failed to load override contract: %w", err)
				}
				if err := imContractStore.OverrideContract(overrideContract.Name, Config.OverrideContracts.TaskMailbox.ChainIds, overrideContract); err != nil {
					return fmt.Errorf("failed to override contract: %w", err)
				}
			}
		}

		tlp := transactionLogParser.NewTransactionLogParser(imContractStore, log)

		sugar.Infof("Aggregator config: %+v\n", Config)
		sugar.Infow("Building aggregator components...")

		var pdf peering.IPeeringDataFetcher
		if Config.SimulationConfig != nil && Config.SimulationConfig.SimulatePeering != nil && Config.SimulationConfig.SimulatePeering.Enabled {
			simulatedPeers, err := peers.NewSimulatedPeersFromConfig(Config.SimulationConfig.SimulatePeering.OperatorPeers)
			if err != nil {
				log.Sugar().Fatalw("Failed to create simulated peers", zap.Error(err))
			}

			pdf = localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
				OperatorPeers: simulatedPeers,
			}, log)
		} else {
			l1Chain := util.Find(Config.Chains, func(c *aggregatorConfig.Chain) bool {
				return c.ChainId == Config.L1ChainId
			})
			if l1Chain == nil {
				return fmt.Errorf("l1 chain not found in config")
			}

			cc, err := aggregator.InitializeContractCaller(&aggregatorConfig.Chain{
				ChainId: l1Chain.ChainId,
				RpcURL:  l1Chain.RpcURL,
			}, "", imContractStore, Config.Avss[0].AVSRegistrarAddress, true, log)
			if err != nil {
				return fmt.Errorf("failed to initialize contract caller: %w", err)
			}

			pdf = peeringDataFetcher.NewPeeringDataFetcher(cc, log)
		}

		agg, err := aggregator.NewAggregatorWithRpcServer(
			Config.ServerConfig.Port,
			&aggregator.AggregatorConfig{
				AVSs:          Config.Avss,
				Chains:        Config.Chains,
				Address:       Config.Operator.Address,
				PrivateKey:    Config.Operator.OperatorPrivateKey,
				AggregatorUrl: Config.ServerConfig.AggregatorUrl,
				L1ChainId:     Config.L1ChainId,
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
