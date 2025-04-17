package main

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
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

	// setup sub commands
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		key := config.KebabToSnakeCase(f.Name)
		viper.BindPFlag(key, f) //nolint:errcheck
		viper.BindEnv(key)      //nolint:errcheck
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
		fmt.Printf("Using config file: %s\n", configFile)
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
