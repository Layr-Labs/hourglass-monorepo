package main

import (
	"github.com/Layr-Labs/go-ponos/pkg/config"
	"github.com/Layr-Labs/go-ponos/pkg/executor/executorConfig"
)

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "worker",
	Short: "Execute avs code",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	initConfig(rootCmd)

	rootCmd.PersistentFlags().Bool("debug", false, `"true" or "false"`)

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

func main() {
	Execute()
}
