package main

import (
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
var Config *aggregatorConfig.AggregatorConfig

func init() {
	cobra.OnInitialize(initConfigIfPresent)

	rootCmd.PersistentFlags().String("config", "", "config file path")
	rootCmd.PersistentFlags().Lookup("config")

	rootCmd.PersistentFlags().Bool(aggregatorConfig.Debug, false, `"true" or "false"`)
	rootCmd.PersistentFlags().Lookup(aggregatorConfig.Debug)

	viper.SetEnvPrefix(executorConfig.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	rootCmd.AddCommand(runCmd)
}

func initConfigIfPresent() {
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			panic(err)
		}
		config, err := aggregatorConfig.NewAggregatorConfigFromYamlBytes(data)
		if err != nil {
			panic(err)
		}
		Config = config
	} else {
		Config = aggregatorConfig.NewAggregatorConfig()
	}
}

func main() {
	Execute()
}
