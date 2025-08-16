package transactionSigner

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
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
	// We need to provide a Signer function that returns the transaction unsigned
	// The actual signing happens in SignAndSendTransaction via Web3Signer
	opts := &bind.TransactOpts{
		From:    w3s.fromAddress,
		Context: ctx,
		NoSend:  true,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			// Just return the transaction as-is without signing
			// The actual signing will happen in SignAndSendTransaction
			return tx, nil
		},
	}
	return opts, nil
}

// SignAndSendTransaction signs a transaction and sends it to the network
func (w3s *Web3Signer) SignAndSendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	var FallbackGasTipCap = big.NewInt(15000000000)

	// Estimate gas tip cap
	gasTipCap, err := w3s.ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		// If the transaction failed because the backend does not support
		// eth_maxPriorityFeePerGas, fallback to using the default constant.
		w3s.logger.Debug("SignAndSendTransaction: cannot get gasTipCap, using fallback",
			zap.Error(err),
		)
		gasTipCap = FallbackGasTipCap
	}

	// Get the latest block header for base fee calculation
	header, err := w3s.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block header: %w", err)
	}

	// Calculate gas fee cap: basefee * 3/2 + tip
	overestimatedBasefee := new(big.Int).Div(new(big.Int).Mul(header.BaseFee, big.NewInt(3)), big.NewInt(2))
	gasFeeCap := new(big.Int).Add(overestimatedBasefee, gasTipCap)

	// Estimate gas limit with proper parameters
	gasLimit, err := w3s.ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From:      w3s.fromAddress,
		To:        tx.To(),
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Value:     tx.Value(),
		Data:      tx.Data(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Add 20% buffer to gas limit
	gasLimitWithBuffer := addGasBuffer(gasLimit)

	// Get nonce if not provided
	nonce := tx.Nonce()
	if nonce == 0 {
		pendingNonce, err := w3s.ethClient.PendingNonceAt(ctx, w3s.fromAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}
		nonce = pendingNonce
	}

	// Convert transaction to Web3Signer format with legacy transaction parameters
	// Using legacy transaction type to avoid EIP-1559
	txData := map[string]interface{}{
		"to":       tx.To().Hex(),
		"value":    hexutil.EncodeBig(tx.Value()),
		"gas":      hexutil.EncodeUint64(gasLimitWithBuffer),
		"gasPrice": hexutil.EncodeBig(gasFeeCap), // Use gasFeeCap as gasPrice for legacy tx
		"nonce":    hexutil.EncodeUint64(nonce),
		"data":     hexutil.Encode(tx.Data()),
		"type":     "0x0", // Legacy transaction type
	}

	w3s.logger.Info("SignAndSendTransaction: sending transaction",
		zap.String("to", tx.To().Hex()),
		zap.String("gasTipCap", gasTipCap.String()),
		zap.String("gasFeeCap", gasFeeCap.String()),
		zap.Uint64("gasLimit", gasLimitWithBuffer),
		zap.Uint64("nonce", nonce),
	)

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

	w3s.logger.Info("SignAndSendTransaction: transaction sent",
		zap.String("txHash", signedTx.Hash().Hex()),
	)

	// Wait for receipt and check status
	receipt, err := bind.WaitMined(ctx, w3s.ethClient, &signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction receipt: %w", err)
	}

	// Check transaction status
	if receipt.Status != 1 {
		w3s.logger.Error("SignAndSendTransaction: transaction failed",
			zap.String("txHash", receipt.TxHash.Hex()),
			zap.Uint64("status", receipt.Status),
		)
		return nil, fmt.Errorf("transaction failed with status %d", receipt.Status)
	}

	w3s.logger.Info("SignAndSendTransaction: transaction succeeded",
		zap.String("txHash", receipt.TxHash.Hex()),
	)

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
