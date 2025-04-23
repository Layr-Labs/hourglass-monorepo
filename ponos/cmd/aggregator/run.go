package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/lifecycle"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the aggregator",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)
		configFile := viper.GetString("config")

		if configFile != "" {
			data, err := os.ReadFile(configFile)
			fmt.Printf("config bytes: %s", data)
			if err != nil {
				return err
			}
			Config, err = aggregatorConfig.NewAggregatorConfigFromYamlBytes(data)
			fmt.Printf("config simulation enabled: %t", Config.SimulationConfig.Enabled)
			fmt.Printf("Config object: %+v\n", Config)
			if err != nil {
				return err
			}
		} else {
			Config = aggregatorConfig.NewAggregatorConfig()
		}

		log, _ := logger.NewLogger(&logger.LoggerConfig{Debug: Config.Debug})
		sugar := log.Sugar()

		if err := Config.Validate(); err != nil {
			sugar.Errorw("Invalid configuration", "error", err)
			return err
		}

		sugar.Infow("Starting aggregator...")

		fmt.Printf("simulation enabled: %t", Config.SimulationConfig.Enabled)
		return lifecycle.RunWithShutdown(func(ctx context.Context) error {
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

func startAggregator(ctx context.Context, cfg *aggregatorConfig.AggregatorConfig, logger *zap.Logger) error {
	agg := aggregator.NewAggregator(cfg, logger)
	if err := agg.Start(ctx); err != nil {
		return err
	}
	return nil
}
