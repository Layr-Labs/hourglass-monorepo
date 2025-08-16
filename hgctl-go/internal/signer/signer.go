package signer

import (
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer/web3Signer"
	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type Signers struct {
	ECDSASigner ISigner
	BLSSigner   ISigner
}

type ISigner interface {
	SignMessage(data []byte) ([]byte, error)
	SignMessageForSolidity(data []byte) ([]byte, error)
}

// LoadSignerFromContext loads the appropriate signer based on the context configuration
func FromContext(ctx *config.Context, logger logger.Logger) (ISigner, error) {
	if ctx.Signer == nil {
		return nil, fmt.Errorf("no signer configuration found in context")
	}

	switch ctx.Signer.Type {
	case "keystore":
		return loadKeystoreSigner(ctx)
	case "web3signer":
		return loadWeb3Signer(ctx, logger)
	case "privatekey":
		return loadPrivateKeySigner(ctx)
	default:
		return nil, fmt.Errorf("unsupported signer type: %s", ctx.Signer.Type)
	}
}

func loadKeystoreSigner(ctx *config.Context) (ISigner, error) {
	if ctx.Signer.Keystore == "" {
		return nil, fmt.Errorf("keystore name not specified")
	}

	// Find the keystore reference
	var keystoreRef *config.KeystoreReference
	for _, ks := range ctx.Keystores {
		if ks.Name == ctx.Signer.Keystore {
			keystoreRef = &ks
			break
		}
	}

	if keystoreRef == nil {
		return nil, fmt.Errorf("keystore '%s' not found in context", ctx.Signer.Keystore)
	}

	keystorePath := keystoreRef.Path
	if !filepath.IsAbs(keystorePath) {
		keystorePath = filepath.Join(config.GetConfigDir(), keystorePath)
	}
	password := os.Getenv("KEYSTORE_PASSWORD")

	switch keystoreRef.Type {
	case "bn254":

		storedKeys, err := keystore.LoadKeystoreFile(keystorePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load keystore file '%s': %w", keystorePath, err)
		}
		privateKey, err := storedKeys.GetBN254PrivateKey(password)
		if err != nil {
			return nil, fmt.Errorf("failed to get BN254 private key: %w", err)
		}
		return inMemorySigner.NewInMemorySigner(privateKey, config.CurveTypeBN254), nil

	case "ecdsa":

		keyStoreContents, err := os.ReadFile(filepath.Clean(keystorePath))
		if err != nil {
			return nil, err
		}

		key, err := gethkeystore.DecryptKey(keyStoreContents, password)
		if err != nil {
			return nil, err
		}
		return inMemorySigner.NewInMemorySigner(key.PrivateKey, config.CurveTypeECDSA), nil

	default:
		return nil, fmt.Errorf("unsupported keystore type: %s", keystoreRef.Type)
	}
}

func loadWeb3Signer(ctx *config.Context, logger logger.Logger) (ISigner, error) {
	if ctx.Signer.Web3Signer == "" {
		return nil, fmt.Errorf("web3signer name not specified")
	}

	var web3SignerRef *config.Web3SignerReference
	for _, ws := range ctx.Web3Signers {
		if ws.Name == ctx.Signer.Web3Signer {
			web3SignerRef = &ws
			break
		}
	}

	if web3SignerRef == nil {
		return nil, fmt.Errorf("web3signer '%s' not found in context", ctx.Signer.Web3Signer)
	}

	var signerConfig *config.RemoteSignerConfig
	if web3SignerRef.ConfigPath != "" {
		configPath := web3SignerRef.ConfigPath
		if !filepath.IsAbs(configPath) {
			configPath = filepath.Join(config.GetConfigDir(), configPath)
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read web3signer config: %w", err)
		}

		signerConfig = &config.RemoteSignerConfig{}
		if err := yaml.Unmarshal(data, signerConfig); err != nil {
			return nil, fmt.Errorf("failed to parse web3signer config: %w", err)
		}
	} else {
		signerConfig = &config.RemoteSignerConfig{
			Url: os.Getenv(fmt.Sprintf("WEB3SIGNER_URL_%s", strings.ToUpper(web3SignerRef.Name))),
		}
		if signerConfig.Url == "" {
			return nil, fmt.Errorf("web3signer URL not found for '%s'", web3SignerRef.Name)
		}
	}

	if web3SignerRef.CACertPath != "" {
		caCert, err := loadCertFile(web3SignerRef.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA cert: %w", err)
		}
		signerConfig.CACert = caCert
	}

	if web3SignerRef.ClientCertPath != "" {
		clientCert, err := loadCertFile(web3SignerRef.ClientCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %w", err)
		}
		signerConfig.Cert = clientCert
	}

	if web3SignerRef.ClientKeyPath != "" {
		clientKey, err := loadCertFile(web3SignerRef.ClientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client key: %w", err)
		}
		signerConfig.Key = clientKey
	}

	c, err := client.NewWeb3SignerClientFromRemoteSignerConfig(signerConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create web3signer client: %w", err)
	}

	return web3Signer.NewWeb3Signer(
		c,
		common.HexToAddress(signerConfig.FromAddress),
		signerConfig.PublicKey,
		config.CurveTypeECDSA,
		logger,
	)
}

func loadPrivateKeySigner(ctx *config.Context) (ISigner, error) {
	// Check for private key in order of precedence:
	// 1. Context PrivateKey field (set via flag)
	// 2. Environment variable
	privateKey := ctx.PrivateKey
	if privateKey == "" {
		privateKey = os.Getenv("PRIVATE_KEY")
	}
	if privateKey == "" {
		return nil, fmt.Errorf("private key not found")
	}

	ecdsaPk, err := ecdsa.NewPrivateKeyFromHexString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA private key: %w", err)
	}

	return inMemorySigner.NewInMemorySigner(ecdsaPk, config.CurveTypeECDSA), nil
}

func loadCertFile(path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(config.GetConfigDir(), path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
