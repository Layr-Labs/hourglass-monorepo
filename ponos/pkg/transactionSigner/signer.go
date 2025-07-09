package transactionSigner

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TransactionSigner provides methods for signing Ethereum transactions
type TransactionSigner interface {
	// GetTransactOpts returns transaction options for creating unsigned transactions
	GetTransactOpts(ctx context.Context) (*bind.TransactOpts, error)

	// SignAndSendTransaction signs a transaction and sends it to the network
	SignAndSendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error)

	// GetFromAddress returns the address that will be used for signing
	GetFromAddress() common.Address

	// EstimateGasPriceAndLimit estimates gas price and limit for a transaction
	EstimateGasPriceAndLimit(ctx context.Context, tx *types.Transaction) (*big.Int, uint64, error)
}
