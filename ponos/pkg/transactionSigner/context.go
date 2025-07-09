package transactionSigner

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// SigningContext provides common functionality for transaction signing
type SigningContext struct {
	ethClient *ethclient.Client
	logger    *zap.Logger
	chainID   *big.Int
}

// NewSigningContext creates a new signing context
func NewSigningContext(ethClient *ethclient.Client, logger *zap.Logger) (*SigningContext, error) {
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return &SigningContext{
		ethClient: ethClient,
		logger:    logger,
		chainID:   chainID,
	}, nil
}

// EstimateGasPriceAndLimit provides common gas estimation logic
func (sc *SigningContext) EstimateGasPriceAndLimit(ctx context.Context, tx *types.Transaction) (*big.Int, uint64, error) {
	// Implementation for gas estimation logic
	// This would contain the current EstimateGasPriceAndLimitAndSendTx logic
	// extracted and made generic
	return nil, 0, nil
}
