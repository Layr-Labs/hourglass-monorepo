package main

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/blsHelper/blsHelperConfig"
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

func init() {
	initConfig(rootCmd)

	rootCmd.AddCommand(generateOperatorData)

	rootCmd.PersistentFlags().Bool(blsHelperConfig.Debug, false, `"true" or "false"`)
	rootCmd.PersistentFlags().String(blsHelperConfig.KeyfilePath, "", "path to the keyfile")
	rootCmd.PersistentFlags().String(blsHelperConfig.KeyPassword, "", "password for the keyfile")

	generateOperatorData.PersistentFlags().String(blsHelperConfig.RpcUrl, "", "RPC URL to connect to the Ethereum node")
	generateOperatorData.PersistentFlags().String(blsHelperConfig.OperatorAddress, "", "address of the operator")
	generateOperatorData.PersistentFlags().String(blsHelperConfig.AvsRegistrarAddress, "", "address of the AVS registrar contract")
	generateOperatorData.PersistentFlags().String(blsHelperConfig.Socket, "", "socket path for IPC connection")

	for _, cmd := range []*cobra.Command{rootCmd} {
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			key := config.KebabToSnakeCase(f.Name)
			if err := viper.BindPFlag(key, f); err != nil {
				panic(fmt.Errorf("failed to bind flag %s: %w", key, err))
			}
			if err := viper.BindEnv(key); err != nil {
				panic(fmt.Errorf("failed to bind env %s: %w", key, err))
			}
		})
	}
}

func initConfig(cmd *cobra.Command) {
	viper.SetEnvPrefix(keygenConfig.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
}

func main() {
	Execute()
}
