package caller

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	ethereum2 "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
	"math/big"
)

var (
	FallbackGasTipCap = big.NewInt(15000000000)
)

func (cc *ContractCaller) EstimateGasPriceAndLimitAndSendTx(
	ctx context.Context,
	fromAddress common.Address,
	tx *ethereumTypes.Transaction,
	privateKey *ecdsa.PrivateKey,
	tag string,
) (*ethereumTypes.Receipt, error) {

	gasTipCap, err := cc.ethclient.SuggestGasTipCap(ctx)
	if err != nil {
		// If the transaction failed because the backend does not support
		// eth_maxPriorityFeePerGas, fallback to using the default constant.
		// Currently Alchemy is the only backend provider that exposes this
		// method, so in the event their API is unreachable we can fallback to a
		// degraded mode of operation. This also applies to our test
		// environments, as hardhat doesn't support the query either.
		cc.logger.Sugar().Debugw("EstimateGasPriceAndLimitAndSendTx: cannot get gasTipCap",
			zap.String("error", err.Error()),
		)

		gasTipCap = FallbackGasTipCap
	}

	header, err := cc.ethclient.HeaderByNumber(ctx, nil)
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
	gasLimit, err := cc.ethclient.EstimateGas(ctx, ethereum2.CallMsg{
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

	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, tx.ChainId())
	if err != nil {
		return nil, fmt.Errorf("EstimateGasPriceAndLimitAndSendTx: cannot create transactOpts: %w", err)
	}
	opts.Context = ctx
	opts.Nonce = new(big.Int).SetUint64(tx.Nonce())
	opts.GasTipCap = gasTipCap
	opts.GasFeeCap = gasFeeCap
	opts.GasLimit = addGasBuffer(gasLimit)

	contract := bind.NewBoundContract(*tx.To(), abi.ABI{}, cc.ethclient, cc.ethclient, cc.ethclient)

	cc.logger.Sugar().Infof("EstimateGasPriceAndLimitAndSendTx: sending txn (%s) with gasTipCap=%v gasFeeCap=%v gasLimit=%v", tag, gasTipCap, gasFeeCap, opts.GasLimit)

	tx, err = contract.RawTransact(opts, tx.Data())
	if err != nil {
		return nil, fmt.Errorf("EstimateGasPriceAndLimitAndSendTx: failed to send txn (%s): %w", tag, err)
	}

	cc.logger.Sugar().Infof("EstimateGasPriceAndLimitAndSendTx: sent txn (%s) with hash=%s", tag, tx.Hash().Hex())

	receipt, err := cc.EnsureTransactionEvaled(ctx, tx, tag)
	if err != nil {
		return nil, err
	}

	return receipt, err
}

func (cc *ContractCaller) EnsureTransactionEvaled(ctx context.Context, tx *ethereumTypes.Transaction, tag string) (*ethereumTypes.Receipt, error) {
	cc.logger.Sugar().Infow("EnsureTransactionEvaled entered")

	receipt, err := bind.WaitMined(ctx, cc.ethclient, tx)
	if err != nil {
		return nil, fmt.Errorf("EnsureTransactionEvaled: failed to wait for transaction (%s) to mine: %w", tag, err)
	}
	if receipt.Status != 1 {
		cc.logger.Sugar().Errorf("EnsureTransactionEvaled: transaction (%s) failed: %v", tag, receipt)
		return nil, errors.New("ErrTransactionFailed")
	}
	cc.logger.Sugar().Infof("EnsureTransactionEvaled: transaction (%s) succeeded: %v", tag, receipt.TxHash.Hex())
	return receipt, nil
}

func addGasBuffer(gasLimit uint64) uint64 {
	return 6 * gasLimit / 5 // add 20% buffer to gas limit
}
