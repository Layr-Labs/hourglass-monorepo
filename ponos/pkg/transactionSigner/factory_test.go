package transactionSigner

import (
	"math/big"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestCreateSigner_PrivateKey(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := "0x" + common.Bytes2Hex(privateKeyBytes)

	// Create a mock signing context - we'll pass nil for ethClient since we're not testing actual network calls
	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	// Test that we can create a private key signer directly
	signer, err := NewPrivateKeySigner(privateKeyHex, signingContext)
	require.NoError(t, err)

	assert.NotNil(t, signer)
	assert.Equal(t, crypto.PubkeyToAddress(privateKey.PublicKey), signer.GetFromAddress())

	// Test that the signer implements the TransactionSigner interface
	var _ TransactionSigner = signer
}

func TestCreateSigner_Web3Signer(t *testing.T) {
	logger := zaptest.NewLogger(t)

	fromAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23")

	// Create a mock signing context
	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	// Test that we can create a web3signer directly with real web3signer client
	web3SignerConfig := &web3signer.Config{
		BaseURL: "http://localhost:9000",
	}
	web3SignerClient, err := web3signer.NewClient(web3SignerConfig, logger)
	require.NoError(t, err)
	signer := NewWeb3Signer(web3SignerClient, fromAddress, signingContext)

	assert.NotNil(t, signer)
	assert.Equal(t, fromAddress, signer.GetFromAddress())

	// Test that the signer implements the TransactionSigner interface
	var _ TransactionSigner = signer
}

func TestCreateSigner_InvalidPrivateKey(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a mock signing context
	signingContext := &SigningContext{
		ethClient: nil,
		logger:    logger,
		chainID:   big.NewInt(1337),
	}

	// Test invalid private key
	_, err := NewPrivateKeySigner("invalid-key", signingContext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse private key")
}
