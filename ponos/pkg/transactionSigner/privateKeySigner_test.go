package transactionSigner

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewPrivateKeySigner(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Create a mock signing context
	signingContext := &SigningContext{
		ethClient: nil, // We'll use nil for unit tests
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	// Use the actual hex representation of the private key
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := "0x" + common.Bytes2Hex(privateKeyBytes)

	signer, err := NewPrivateKeySigner(privateKeyHex, signingContext)
	require.NoError(t, err)
	assert.NotNil(t, signer)
	assert.Equal(t, crypto.PubkeyToAddress(privateKey.PublicKey), signer.GetFromAddress())

	// Test invalid private key
	_, err = NewPrivateKeySigner("invalid-key", signingContext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestPrivateKeySigner_GetTransactOpts(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := "0x" + common.Bytes2Hex(privateKeyBytes)

	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	signer, err := NewPrivateKeySigner(privateKeyHex, signingContext)
	require.NoError(t, err)

	ctx := context.Background()
	opts, err := signer.GetTransactOpts(ctx)
	require.NoError(t, err)

	assert.Equal(t, crypto.PubkeyToAddress(privateKey.PublicKey), opts.From)
	assert.True(t, opts.NoSend)
	assert.Equal(t, ctx, opts.Context)
}

func TestPrivateKeySigner_GetFromAddress(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := "0x" + common.Bytes2Hex(privateKeyBytes)

	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	signer, err := NewPrivateKeySigner(privateKeyHex, signingContext)
	require.NoError(t, err)

	expectedAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	assert.Equal(t, expectedAddress, signer.GetFromAddress())
}

func TestPrivateKeySigner_EstimateGasPriceAndLimit(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := "0x" + common.Bytes2Hex(privateKeyBytes)

	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	signer, err := NewPrivateKeySigner(privateKeyHex, signingContext)
	require.NoError(t, err)

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
