package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/peeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/signerUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the aggregator",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)
		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: Config.Debug})
		sugar := l.Sugar()

		if err := Config.Validate(); err != nil {
			sugar.Errorw("Invalid configuration", "error", err)
			return err
		}

		signers, err := signerUtils.ParseSignersFromOperatorConfig(Config.Operator, l)
		if err != nil {
			return fmt.Errorf("failed to parse signers from operator config: %w", err)
		}

		// load the contracts and create the store
		var coreContracts []*contracts.Contract
		if len(Config.Contracts) > 0 {
			l.Sugar().Infow("Loading core contracts from runtime config")
			coreContracts, err = eigenlayer.LoadContractsFromRuntime(string(Config.Contracts))
			if err != nil {
				return fmt.Errorf("failed to load core contracts from runtime: %w", err)
			}
		} else {
			l.Sugar().Infow("Loading core contracts from embedded config")
			coreContracts, err = eigenlayer.LoadContracts()
			if err != nil {
				return fmt.Errorf("failed to load core contracts: %w", err)
			}
		}
		imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, l)

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

		tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)

		sugar.Infof("Aggregator config: %+v\n", Config)
		sugar.Infow("Building aggregator components...")

		l1Chain := util.Find(Config.Chains, func(c *aggregatorConfig.Chain) bool {
			return c.ChainId == Config.L1ChainId
		})
		if l1Chain == nil {
			return fmt.Errorf("l1 chain not found in config")
		}

		cc, err := aggregator.InitializeContractCaller(
			&aggregatorConfig.Chain{
				ChainId: l1Chain.ChainId,
				RpcURL:  l1Chain.RpcURL,
			},
			Config.Operator.OperatorPrivateKey,
			l,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize contract caller: %w", err)
		}

		pdf := peeringDataFetcher.NewPeeringDataFetcher(cc, l)

		agg, err := aggregator.NewAggregator(
			&aggregator.AggregatorConfig{
				AVSs:             Config.Avss,
				Chains:           Config.Chains,
				Address:          Config.Operator.Address,
				PrivateKeyConfig: Config.Operator.OperatorPrivateKey,
				L1ChainId:        Config.L1ChainId,
			},
			imContractStore,
			tlp,
			pdf,
			signers,
			l,
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
			l.Sugar().Info("Shutting down...")
			cancel()
		}, time.Second*5, l)

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
