package transactionSigner

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// Web3Signer implements ITransactionSigner using Web3Signer service
type Web3Signer struct {
	ethClient        *ethclient.Client
	logger           *zap.Logger
	chainID          *big.Int
	web3SignerClient *web3signer.Client
	fromAddress      common.Address
}

// NewWeb3Signer creates a new Web3Signer
func NewWeb3Signer(web3SignerClient *web3signer.Client, fromAddress common.Address, ethClient *ethclient.Client, logger *zap.Logger) (*Web3Signer, error) {
	// Get chain ID during initialization
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return &Web3Signer{
		ethClient:        ethClient,
		logger:           logger,
		chainID:          chainID,
		web3SignerClient: web3SignerClient,
		fromAddress:      fromAddress,
	}, nil
}

// GetTransactOpts returns transaction options for creating unsigned transactions
func (w3s *Web3Signer) GetTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts := &bind.TransactOpts{
		From:    w3s.fromAddress,
		Context: ctx,
		NoSend:  true,
		Signer:  w3s.signTransaction,
	}
	return opts, nil
}

// SignAndSendTransaction signs a transaction and sends it to the network
func (w3s *Web3Signer) SignAndSendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	// Convert transaction to Web3Signer format
	txData := map[string]interface{}{
		"to":       tx.To().Hex(),
		"value":    hexutil.EncodeBig(tx.Value()),
		"gas":      hexutil.EncodeUint64(tx.Gas()),
		"gasPrice": hexutil.EncodeBig(tx.GasPrice()),
		"nonce":    hexutil.EncodeUint64(tx.Nonce()),
		"data":     hexutil.Encode(tx.Data()),
	}

	// Sign with Web3Signer
	signedTxHex, err := w3s.web3SignerClient.EthSignTransaction(ctx, w3s.fromAddress.Hex(), txData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction with Web3Signer: %w", err)
	}

	// Parse signed transaction
	signedTxBytes, err := hexutil.Decode(signedTxHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signed transaction: %w", err)
	}

	var signedTx types.Transaction
	err = signedTx.UnmarshalBinary(signedTxBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal signed transaction: %w", err)
	}

	// Send the transaction
	err = w3s.ethClient.SendTransaction(ctx, &signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for receipt
	receipt, err := bind.WaitMined(ctx, w3s.ethClient, &signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction receipt: %w", err)
	}

	return receipt, nil
}

// GetFromAddress returns the address that will be used for signing
func (w3s *Web3Signer) GetFromAddress() common.Address {
	return w3s.fromAddress
}

// EstimateGasPriceAndLimit estimates gas price and limit for a transaction
func (w3s *Web3Signer) EstimateGasPriceAndLimit(ctx context.Context, tx *types.Transaction) (*big.Int, uint64, error) {
	// For now, return nil values - this method isn't fully implemented yet
	return nil, 0, nil
}

// signTransaction is a signing function for bind.TransactOpts
func (w3s *Web3Signer) signTransaction(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
	// This would be called by go-ethereum binding code
	// Implementation depends on specific requirements
	return nil, fmt.Errorf("direct signing not supported for Web3Signer")
}
