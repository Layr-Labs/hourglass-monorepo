package signerUtils

import (
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	web3SignerClient "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/web3Signer"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func ParseSignersFromOperatorConfig(opConfig *config.OperatorConfig, l *zap.Logger) (signer.Signers, error) {
	signers := signer.Signers{}
	if opConfig.SigningKeys.BLS != nil {
		// Load up the keystore
		var err error
		var storedKeys *keystore.EIP2335Keystore
		if opConfig.SigningKeys.BLS.Keystore != "" {
			storedKeys, err = keystore.ParseKeystoreJSON(opConfig.SigningKeys.BLS.Keystore)
			if err != nil {
				return signers, fmt.Errorf("failed to parse keystore JSON: %w", err)
			}
		} else {
			storedKeys, err = keystore.LoadKeystoreFile(opConfig.SigningKeys.BLS.KeystoreFile)
			if err != nil {
				return signers, fmt.Errorf("failed to load keystore file: '%s' %w", opConfig.SigningKeys.BLS.KeystoreFile, err)
			}
		}

		privateSigningKey, err := storedKeys.GetBN254PrivateKey(opConfig.SigningKeys.BLS.Password)
		if err != nil {
			return signers, fmt.Errorf("failed to get private key: %w", err)
		}

		signers.BLSSigner = inMemorySigner.NewInMemorySigner(privateSigningKey, config.CurveTypeBN254)
	}

	if opConfig.SigningKeys.ECDSA != nil {
		if opConfig.SigningKeys.ECDSA.UseRemoteSigner && opConfig.SigningKeys.ECDSA.RemoteSignerConfig != nil {
			client, err := web3SignerClient.NewWeb3SignerClientFromRemoteSignerConfig(opConfig.SigningKeys.ECDSA.RemoteSignerConfig, l)
			if err != nil {
				return signers, fmt.Errorf("failed to create web3signer client: %w", err)
			}
			sig, err := web3Signer.NewWeb3Signer(
				client,
				common.HexToAddress(opConfig.SigningKeys.ECDSA.RemoteSignerConfig.FromAddress),
				opConfig.SigningKeys.ECDSA.RemoteSignerConfig.PublicKey,
				config.CurveTypeECDSA,
				l,
			)
			if err != nil {
				return signers, fmt.Errorf("failed to create web3 signer: %w", err)
			}
			signers.ECDSASigner = sig
		} else if opConfig.SigningKeys.ECDSA.PrivateKey != "" {
			ecdsaPk, err := ecdsa.NewPrivateKeyFromHexString(opConfig.SigningKeys.ECDSA.PrivateKey)
			if err != nil {
				return signers, fmt.Errorf("failed to create ECDSA private key: %w", err)
			}
			signers.ECDSASigner = inMemorySigner.NewInMemorySigner(ecdsaPk, config.CurveTypeECDSA)
		} else {
			l.Sugar().Warn("No ECDSA signing key provided")
		}
	}
	return signers, nil
}
