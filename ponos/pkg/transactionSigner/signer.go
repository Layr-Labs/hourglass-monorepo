package transactionSigner

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ITransactionSigner provides methods for signing Ethereum transactions
type ITransactionSigner interface {
	// GetTransactOpts returns transaction options for creating unsigned transactions
	GetTransactOpts(ctx context.Context) (*bind.TransactOpts, error)

	// SignAndSendTransaction signs a transaction and sends it to the network
	SignAndSendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error)

	// GetFromAddress returns the address that will be used for signing
	GetFromAddress() common.Address

	// EstimateGasPriceAndLimit estimates gas price and limit for a transaction
	EstimateGasPriceAndLimit(ctx context.Context, tx *types.Transaction) (*big.Int, uint64, error)
}

func NewTransactionSigner(cfg *config.ECDSAKeyConfig, ethClient *ethclient.Client, logger *zap.Logger) (ITransactionSigner, error) {
	if cfg.UseRemoteSigner {
		// Create web3signer config with TLS support
		var web3SignerConfig *web3signer.Config
		if cfg.RemoteSignerConfig != nil {
			web3SignerConfig = web3signer.NewConfigWithTLS(
				cfg.RemoteSignerConfig.Url,
				cfg.RemoteSignerConfig.CACert,
				cfg.RemoteSignerConfig.Cert,
				cfg.RemoteSignerConfig.Key,
			)
		} else {
			web3SignerConfig = web3signer.DefaultConfig()
		}

		web3SignerClient, err := web3signer.NewClient(web3SignerConfig, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create web3signer client: %w", err)
		}

		fromAddress := common.Address{}
		if cfg.RemoteSignerConfig != nil && cfg.RemoteSignerConfig.FromAddress != "" {
			fromAddress = common.HexToAddress(cfg.RemoteSignerConfig.FromAddress)
		}

		return NewWeb3Signer(web3SignerClient, fromAddress, ethClient, logger)
	}

	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("private key cannot be empty")
	}

	return NewPrivateKeySigner(cfg.PrivateKey, ethClient, logger)
}
