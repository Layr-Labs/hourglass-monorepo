package transactionSigner

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// SignerConfig represents configuration for creating signers
type SignerConfig struct {
	Type string `yaml:"type"` // "private_key" or "web3signer"

	// Private key configuration
	PrivateKey string `yaml:"private_key,omitempty"`

	// Web3Signer configuration
	Web3SignerURL string `yaml:"web3signer_url,omitempty"`
	FromAddress   string `yaml:"from_address,omitempty"`
}

// CreateSigner creates a signer based on configuration
func CreateSigner(config *SignerConfig, ethClient *ethclient.Client, logger *zap.Logger) (TransactionSigner, error) {
	signingContext, err := NewSigningContext(ethClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing context: %w", err)
	}

	switch config.Type {
	case "private_key":
		return NewPrivateKeySigner(config.PrivateKey, signingContext)

	case "web3signer":
		web3SignerConfig := &web3signer.Config{
			BaseURL: config.Web3SignerURL,
		}
		web3SignerClient := web3signer.NewClient(web3SignerConfig, logger)

		fromAddress := common.HexToAddress(config.FromAddress)
		return NewWeb3Signer(web3SignerClient, fromAddress, signingContext), nil

	default:
		return nil, fmt.Errorf("unsupported signer type: %s", config.Type)
	}
}
