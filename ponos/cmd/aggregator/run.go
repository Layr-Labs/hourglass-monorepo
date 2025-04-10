package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/shutdown"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the aggregator",
	Run: func(cmd *cobra.Command, args []string) {
		initRunCmd(cmd)

		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})

		l.Sugar().Infow("aggregator run")

		_, cancel := context.WithCancel(context.Background())
		gracefulShutdownNotifier := shutdown.CreateGracefulShutdownChannel()
		done := make(chan bool)
		shutdown.ListenForShutdown(gracefulShutdownNotifier, done, func() {
			l.Sugar().Info("Shutting down...")
			cancel()
		}, time.Second*5, l)
	},
}

func initRunCmd(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := viper.BindPFlag(config.KebabToSnakeCase(f.Name), f); err != nil {
			fmt.Printf("Failed to bind flag '%s' - %+v\n", f.Name, err)
		}
		if err := viper.BindEnv(f.Name); err != nil {
			fmt.Printf("Failed to bind env '%s' - %+v\n", f.Name, err)
		}

	})
}
