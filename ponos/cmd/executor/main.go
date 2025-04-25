package main

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"log"
)

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "executor",
	Short: "Execute tasks",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var configFile string
var Config *executorConfig.ExecutorConfig

func init() {
	cobra.OnInitialize(initConfigIfPresent)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")

	initConfig(rootCmd)

	rootCmd.PersistentFlags().Bool(executorConfig.Debug, false, `"true" or "false"`)
	rootCmd.PersistentFlags().Int(executorConfig.GrpcPort, 9090, "gRPC port")
	rootCmd.PersistentFlags().String(executorConfig.PerformerNetworkName, "", "Docker network name for executor (leave blank if using localhost)")

	// setup sub commands
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		key := config.KebabToSnakeCase(f.Name)
		if err := viper.BindPFlag(key, f); err != nil {
			log.Fatalf("Failed to bind flag '%s' - %+v\n", f.Name, err)
		}
		if err := viper.BindEnv(key); err != nil {
			log.Fatalf("Failed to bind env '%s' - %+v\n", key, err)
		}
	})

}

func initConfig(cmd *cobra.Command) {
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
			panic(err)
		}
		if err := viper.Unmarshal(&Config); err != nil {
			panic(err)
		}
	} else {
		Config = executorConfig.NewExecutorConfig()
	}
}

func main() {
	Execute()
}
