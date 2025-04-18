package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"

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

		sugar.Infow("Starting aggregator...")

		return runWithShutdown(func(ctx context.Context) error {
			return startAggregator(ctx, Config, log)
		}, log)
	},
}

func initRunCmd(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := viper.BindPFlag(config.KebabToSnakeCase(f.Name), f); err != nil {
			fmt.Printf("Failed to bind flag '%s': %+v\n", f.Name, err)
		}
		if err := viper.BindEnv(f.Name); err != nil {
			fmt.Printf("Failed to bind env '%s': %+v\n", f.Name, err)
		}
	})
}

func runWithShutdown(startFunc func(ctx context.Context) error, logger *zap.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := startFunc(ctx); err != nil {
		return err
	}

	gracefulShutdownNotifier := shutdown.CreateGracefulShutdownChannel()
	done := make(chan bool)

	shutdown.ListenForShutdown(gracefulShutdownNotifier, done, func() {
		logger.Sugar().Info("Shutting down aggregator...")
		cancel()
	}, 5*time.Second, logger)

	return nil
}

func startAggregator(ctx context.Context, cfg *aggregatorConfig.AggregatorConfig, logger *zap.Logger) error {

	agg := aggregator.NewAggregator(cfg, logger)

	go func() {
		if err := agg.Start(ctx); err != nil {
			logger.Sugar().Fatalw("Aggregator start failed", zap.Error(err))
		}
	}()

	return nil
}
