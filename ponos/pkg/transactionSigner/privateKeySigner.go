package transactionSigner

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// PrivateKeySigner implements ITransactionSigner using a private key
type PrivateKeySigner struct {
	*SigningContext
	privateKey  *ecdsa.PrivateKey
	fromAddress common.Address
}

// NewPrivateKeySigner creates a new private key signer
func NewPrivateKeySigner(privateKeyHex string, signingContext *SigningContext) (*PrivateKeySigner, error) {
	privateKey, err := cryptoUtils.StringToECDSAPrivateKey(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &PrivateKeySigner{
		SigningContext: signingContext,
		privateKey:     privateKey,
		fromAddress:    fromAddress,
	}, nil
}

// GetTransactOpts returns transaction options for creating unsigned transactions
func (pks *PrivateKeySigner) GetTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(pks.privateKey, pks.SigningContext.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	opts.NoSend = true
	opts.Context = ctx
	return opts, nil
}

// SignAndSendTransaction signs a transaction and sends it to the network
func (pks *PrivateKeySigner) SignAndSendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	// Use the same gas estimation and transaction sending logic as the original implementation
	return pks.estimateGasPriceAndLimitAndSendTx(ctx, pks.fromAddress, tx, "SignAndSendTransaction")
}

// GetFromAddress returns the address that will be used for signing
func (pks *PrivateKeySigner) GetFromAddress() common.Address {
	return pks.fromAddress
}

// EstimateGasPriceAndLimit estimates gas price and limit for a transaction
func (pks *PrivateKeySigner) EstimateGasPriceAndLimit(ctx context.Context, tx *types.Transaction) (*big.Int, uint64, error) {
	return pks.SigningContext.EstimateGasPriceAndLimit(ctx, tx)
}

// estimateGasPriceAndLimitAndSendTx replicates the original EstimateGasPriceAndLimitAndSendTx method
func (pks *PrivateKeySigner) estimateGasPriceAndLimitAndSendTx(ctx context.Context, fromAddress common.Address, tx *types.Transaction, tag string) (*types.Receipt, error) {
	var FallbackGasTipCap = big.NewInt(15000000000)

	gasTipCap, err := pks.SigningContext.ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		// If the transaction failed because the backend does not support
		// eth_maxPriorityFeePerGas, fallback to using the default constant.
		pks.SigningContext.logger.Sugar().Debugw("estimateGasPriceAndLimitAndSendTx: cannot get gasTipCap",
			"error", err.Error(),
		)
		gasTipCap = FallbackGasTipCap
	}

	header, err := pks.SigningContext.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	// get header basefee * 3/2
	overestimatedBasefee := new(big.Int).Div(new(big.Int).Mul(header.BaseFee, big.NewInt(3)), big.NewInt(2))
	gasFeeCap := new(big.Int).Add(overestimatedBasefee, gasTipCap)

	// The estimated gas limits performed by RawTransact fail semi-regularly
	// with out of gas exceptions. To remedy this we extract the internal calls
	// to perform gas price/gas limit estimation here and add a buffer to
	// account for any network variability.
	gasLimit, err := pks.SigningContext.ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From:      fromAddress,
		To:        tx.To(),
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Value:     nil,
		Data:      tx.Data(),
	})

	if err != nil {
		return nil, err
	}

	opts, err := bind.NewKeyedTransactorWithChainID(pks.privateKey, pks.SigningContext.chainID)
	if err != nil {
		return nil, fmt.Errorf("estimateGasPriceAndLimitAndSendTx: cannot create transactOpts: %w", err)
	}
	opts.Context = ctx
	opts.Nonce = new(big.Int).SetUint64(tx.Nonce())
	opts.GasTipCap = gasTipCap
	opts.GasFeeCap = gasFeeCap
	opts.GasLimit = addGasBuffer(gasLimit)

	contract := bind.NewBoundContract(*tx.To(), abi.ABI{}, pks.SigningContext.ethClient, pks.SigningContext.ethClient, pks.SigningContext.ethClient)

	pks.SigningContext.logger.Sugar().Infof("estimateGasPriceAndLimitAndSendTx: sending txn (%s) with gasTipCap=%v gasFeeCap=%v gasLimit=%v", tag, gasTipCap, gasFeeCap, opts.GasLimit)

	tx, err = contract.RawTransact(opts, tx.Data())
	if err != nil {
		return nil, fmt.Errorf("estimateGasPriceAndLimitAndSendTx: failed to send txn (%s): %w", tag, err)
	}

	pks.SigningContext.logger.Sugar().Infof("estimateGasPriceAndLimitAndSendTx: sent txn (%s) with hash=%s", tag, tx.Hash().Hex())

	receipt, err := pks.ensureTransactionEvaled(ctx, tx, tag)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// ensureTransactionEvaled waits for transaction to be mined and checks status
func (pks *PrivateKeySigner) ensureTransactionEvaled(ctx context.Context, tx *types.Transaction, tag string) (*types.Receipt, error) {
	pks.SigningContext.logger.Sugar().Infow("ensureTransactionEvaled entered")

	receipt, err := bind.WaitMined(ctx, pks.SigningContext.ethClient, tx)
	if err != nil {
		return nil, fmt.Errorf("ensureTransactionEvaled: failed to wait for transaction (%s) to mine: %w", tag, err)
	}
	if receipt.Status != 1 {
		pks.SigningContext.logger.Sugar().Errorf("ensureTransactionEvaled: transaction (%s) failed: %v", tag, receipt)
		return nil, fmt.Errorf("transaction failed")
	}
	pks.SigningContext.logger.Sugar().Infof("ensureTransactionEvaled: transaction (%s) succeeded: %v", tag, receipt.TxHash.Hex())
	return receipt, nil
}

// addGasBuffer adds a buffer to the gas limit
func addGasBuffer(gasLimit uint64) uint64 {
	return 6 * gasLimit / 5 // add 20% buffer to gas limit
}
