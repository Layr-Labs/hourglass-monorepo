package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/keygen/keygenConfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate and manage BLS signing keys",
	Long:  `A tool for generating and managing BLS signing keys for both BLS12-381 and BN254 curves.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var configFile string
var Config *keygenConfig.KeygenConfig

func init() {
	cobra.OnInitialize(initConfigIfPresent)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")

	initConfig(rootCmd)

	rootCmd.PersistentFlags().Bool(keygenConfig.Debug, false, `"true" or "false"`)
	rootCmd.PersistentFlags().String(keygenConfig.CurveType, "bls381", "Curve type: bls381 or bn254")
	rootCmd.PersistentFlags().String(keygenConfig.OutputDir, "./keys", "Directory to save generated keys")
	rootCmd.PersistentFlags().String(keygenConfig.FilePrefix, "key", "Prefix for generated key files")

	// setup sub commands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(infoCmd)

	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		key := config.KebabToSnakeCase(f.Name)
		viper.BindPFlag(key, f) //nolint:errcheck
		viper.BindEnv(key)      //nolint:errcheck
	})
}

func initConfig(cmd *cobra.Command) {
	viper.SetEnvPrefix(keygenConfig.EnvPrefix)
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
		Config = keygenConfig.NewKeygenConfig()
	}
}

func main() {
	Execute()
}
