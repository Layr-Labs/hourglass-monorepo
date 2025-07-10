package transactionSigner

import (
	"context"
	"math/big"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewWeb3Signer(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock web3signer client
	config := &web3signer.Config{
		BaseURL: "http://localhost:9000",
	}
	client, err := web3signer.NewClient(config, logger)
	require.NoError(t, err)

	// Create a test address
	fromAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23")

	// Create a mock signing context
	signingContext := &SigningContext{
		ethClient: nil, // We'll use nil for unit tests
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	signer := NewWeb3Signer(client, fromAddress, signingContext)

	assert.NotNil(t, signer)
	assert.Equal(t, fromAddress, signer.GetFromAddress())
	assert.Equal(t, client, signer.web3SignerClient)
	assert.Equal(t, signingContext, signer.SigningContext)
}

func TestWeb3Signer_GetTransactOpts(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock web3signer client
	config := &web3signer.Config{
		BaseURL: "http://localhost:9000",
	}
	client, err := web3signer.NewClient(config, logger)
	require.NoError(t, err)

	// Create a test address
	fromAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23")

	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	signer := NewWeb3Signer(client, fromAddress, signingContext)

	ctx := context.Background()
	opts, err := signer.GetTransactOpts(ctx)
	require.NoError(t, err)

	assert.Equal(t, fromAddress, opts.From)
	assert.True(t, opts.NoSend)
	assert.Equal(t, ctx, opts.Context)
	assert.NotNil(t, opts.Signer)
}

func TestWeb3Signer_GetFromAddress(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock web3signer client
	config := &web3signer.Config{
		BaseURL: "http://localhost:9000",
	}
	client, err := web3signer.NewClient(config, logger)
	require.NoError(t, err)

	// Create a test address
	fromAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23")

	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	signer := NewWeb3Signer(client, fromAddress, signingContext)

	assert.Equal(t, fromAddress, signer.GetFromAddress())
}

func TestWeb3Signer_EstimateGasPriceAndLimit(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock web3signer client
	config := &web3signer.Config{
		BaseURL: "http://localhost:9000",
	}
	client, err := web3signer.NewClient(config, logger)
	require.NoError(t, err)

	// Create a test address
	fromAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23")

	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	signer := NewWeb3Signer(client, fromAddress, signingContext)

	// Create a dummy transaction
	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23"),
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		nil,
	)

	ctx := context.Background()
	gasPrice, gasLimit, err := signer.EstimateGasPriceAndLimit(ctx, tx)

	// Since we haven't implemented the actual estimation logic yet,
	// we expect nil values and no error
	assert.NoError(t, err)
	assert.Nil(t, gasPrice)
	assert.Equal(t, uint64(0), gasLimit)
}

func TestWeb3Signer_SignTransaction(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock web3signer client
	config := &web3signer.Config{
		BaseURL: "http://localhost:9000",
	}
	client, err := web3signer.NewClient(config, logger)
	require.NoError(t, err)

	// Create a test address
	fromAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23")

	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	signer := NewWeb3Signer(client, fromAddress, signingContext)

	// Create a dummy transaction
	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23"),
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		nil,
	)

	// Test the signTransaction method (should return error for direct signing)
	signedTx, err := signer.signTransaction(fromAddress, tx)
	assert.Error(t, err)
	assert.Nil(t, signedTx)
	assert.Contains(t, err.Error(), "direct signing not supported")
}
