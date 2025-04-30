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

	// setup sub commands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(testCmd)

	rootCmd.PersistentFlags().Bool(keygenConfig.Debug, false, `"true" or "false"`)
	rootCmd.PersistentFlags().String(keygenConfig.CurveType, "", "Curve type: bls381 or bn254")

	generateCmd.PersistentFlags().String(keygenConfig.OutputDir, "./keys", "Directory to save generated keys")
	generateCmd.PersistentFlags().String(keygenConfig.FilePrefix, "key", "Prefix for generated key files")

	// Generate command flags
	generateCmd.PersistentFlags().String("seed", "", "Hex-encoded seed for deterministic key generation")
	generateCmd.PersistentFlags().String("path", "", "Derivation path for EIP-2333 (BLS12-381 only), e.g., m/12381/3600/0/0")
	generateCmd.PersistentFlags().String("password", "", "Password for encrypting the private key (used with keystore format)")
	generateCmd.PersistentFlags().Bool("use-keystore", false, "Save the private key in Web3 Secret Storage format")
	generateCmd.PersistentFlags().Bool("use-random-password", false, "Generate a random password for the keystore")
	generateCmd.PersistentFlags().Bool("light-encryption", false, "Use light encryption for the keystore (faster but less secure)")

	// Info command flags
	infoCmd.PersistentFlags().String("key-file", "", "Path to the key file to display information about")
	infoCmd.PersistentFlags().String("password", "", "Password to decrypt the key file")

	// Test command flags
	// testCmd.PersistentFlags().String("key-file", "", "Path to the keystore file to test")
	// testCmd.PersistentFlags().String("password", "", "Password to decrypt the keystore")

	for _, cmd := range []*cobra.Command{rootCmd, generateCmd, infoCmd, testCmd} {
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
		Config = keygenConfig.NewKeygenConfig()
	}
}

func main() {
	Execute()
}
