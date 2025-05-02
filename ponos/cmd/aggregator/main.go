package main

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/spf13/pflag"
	"os"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "aggregator",
	Short: "Coordinate task distribution and completion",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var configFile string
var Config aggregatorConfig.AggregatorConfig

func init() {
	cobra.OnInitialize(initConfigIfPresent)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")

	initConfig()

	rootCmd.PersistentFlags().Bool(aggregatorConfig.Debug, false, `"true" or "false"`)
	rootCmd.PersistentFlags().Lookup(aggregatorConfig.Debug)

	rootCmd.AddCommand(runCmd)
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		key := config.KebabToSnakeCase(f.Name)
		if err := viper.BindPFlag(key, f); err != nil {
			fmt.Printf("Failed to bind flag %s: %v\n", key, err)
			panic(err)
		}
		if err := viper.BindEnv(key); err != nil {
			fmt.Printf("Failed to bind env %s: %v\n", key, err)
			panic(err)
		}
	})
}

func initConfig() {
	viper.SetEnvPrefix(executorConfig.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
}

func initConfigIfPresent() {
	hasConfig := false
	if configFile != "" {
		viper.SetConfigFile(configFile)
		hasConfig = true
	}
	if hasConfig {
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Failed to read config: %v\n", err)
			panic(err)
		}
		if err := viper.Unmarshal(&Config); err != nil {
			fmt.Printf("Failed to unmarshal config: %v\n", err)
			panic(err)
		}
	} else {
		Config = *aggregatorConfig.NewAggregatorConfig()
	}
}

func main() {
	Execute()
}
