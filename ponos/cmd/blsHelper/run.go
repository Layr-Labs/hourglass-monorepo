package main

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/blsHelper/blsHelperConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var generateOperatorData = &cobra.Command{
	Use:   "generate-operator-data",
	Short: "Generate operator registration data",
	RunE: func(cmd *cobra.Command, args []string) error {
		initRunCmd(cmd)
		ctx := cmd.Context()
		cfg := blsHelperConfig.NewBlsHelperConfig()

		l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: cfg.Debug})

		ks, err := keystore.LoadKeystoreFile(cfg.KeyfilePath)
		if err != nil {
			l.Sugar().Errorw("Failed to load keystore file", "keyfile", cfg.KeyfilePath, "error", err)
			return err
		}

		if ks == nil {
			l.Sugar().Errorw("EIP2335Keystore is nil", "keyfile", cfg.KeyfilePath)
			return fmt.Errorf("EIP2335Keystore is nil")
		}

		if ks.CurveType != "bn254" {
			l.Sugar().Errorw("Unsupported curve type", "curveType", ks.CurveType)
			return fmt.Errorf("Unsupported curve type: %s", ks.CurveType)
		}

		ksPrivateKey, err := ks.GetPrivateKey(cfg.KeyPassword, bn254.NewScheme())
		if err != nil {
			l.Sugar().Errorw("Failed to get private key", "keyfile", cfg.KeyfilePath, "error", err)
			return err
		}

		// convert to the typed bn254 private key rather than the generic signing.PrivateKey
		privateKey, err := bn254.NewPrivateKeyFromBytes(ksPrivateKey.Bytes())
		if err != nil {
			l.Sugar().Errorw("Failed to create private key from bytes", "error", err)
			return err
		}

		ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
			BaseUrl: cfg.RpcUrl,
		}, l)

		ethClient, err := ethereumClient.GetEthereumContractCaller()
		if err != nil {
			return err
		}

		cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
			AVSRegistrarAddress: cfg.AvsAddress,
		}, ethClient, l)
		if err != nil {
			return err
		}

		keyData, err := cc.EncodeBN254KeyData(privateKey.Public())
		if err != nil {
			l.Sugar().Fatalf("failed to encode BN254 key data: %v", err)
		}

		messageHash, err := cc.GetOperatorRegistrationMessageHash(
			ctx,
			common.HexToAddress(cfg.OperatorAddress),
			common.HexToAddress(cfg.AvsAddress),
			cfg.OperatorSetId,
			keyData,
		)
		if err != nil {
			l.Sugar().Fatalf("failed to get operator registration message hash: %v", err)
		}

		sig, err := privateKey.SignSolidityCompatible(messageHash)
		if err != nil {
			l.Sugar().Fatalf("failed to sign message hash: %v", err)
		}

		g1Point := &bn254.G1Point{
			G1Affine: sig.GetG1Point(),
		}
		g1Bytes, err := g1Point.ToPrecompileFormat()
		if err != nil {
			l.Sugar().Fatalf("failed to convert G1 point to precompile format: %v", err)
		}
		fmt.Printf("%x", g1Bytes)
		return nil
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
