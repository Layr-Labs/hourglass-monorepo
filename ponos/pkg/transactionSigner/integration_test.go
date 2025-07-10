package transactionSigner

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestWeb3SignerIntegration tests the Web3Signer integration with a real Web3Signer service
// This test expects that Web3Signer containers are already running via goTest.sh
func TestWeb3SignerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)

	// Create Web3Signer client
	config := &web3signer.Config{
		BaseURL: "http://localhost:9100", // L1 Web3Signer port
		Timeout: 10 * time.Second,
	}

	web3SignerClient, err := web3signer.NewClient(config, logger)
	require.NoError(t, err)

	// Wait for service to be ready and get available accounts
	ctx := context.Background()
	accounts, err := web3SignerClient.EthAccounts(ctx)
	require.NoError(t, err, "Failed to get accounts from Web3Signer")
	require.NotEmpty(t, accounts, "No accounts available in Web3Signer")

	fromAddress := common.HexToAddress(accounts[0])

	// Create signing context (without real ethClient for this test)
	signingContext := &SigningContext{
		ethClient: nil, // We're only testing the signing part, not transaction sending
		logger:    logger,
		chainID:   big.NewInt(31337), // Chain ID matching goTest.sh
	}

	// Create Web3Signer
	signer := NewWeb3Signer(web3SignerClient, fromAddress, signingContext)

	// Test GetFromAddress
	assert.Equal(t, fromAddress, signer.GetFromAddress())

	// Test GetTransactOpts
	opts, err := signer.GetTransactOpts(ctx)
	require.NoError(t, err)
	assert.Equal(t, fromAddress, opts.From)
	assert.True(t, opts.NoSend)
	assert.Equal(t, ctx, opts.Context)

	// Test transaction signing format (without actually sending)
	// Create a test transaction
	tx := types.NewTransaction(
		0, // nonce
		common.HexToAddress("0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23"), // to
		big.NewInt(1),          // value
		21000,                  // gas limit
		big.NewInt(1000000000), // gas price
		nil,                    // data
	)

	// Test that we can format the transaction for Web3Signer
	// This tests the internal transaction formatting logic
	txData := map[string]interface{}{
		"to":       tx.To().Hex(),
		"value":    "0x1",
		"gas":      "0x5208",
		"gasPrice": "0x3b9aca00",
		"nonce":    "0x0",
		"data":     "0x",
	}

	// Test signing with Web3Signer
	signedTxHex, err := web3SignerClient.EthSignTransaction(ctx, fromAddress.Hex(), txData)
	require.NoError(t, err, "Failed to sign transaction with Web3Signer")
	assert.NotEmpty(t, signedTxHex, "Signed transaction should not be empty")
	assert.True(t, len(signedTxHex) > 2, "Signed transaction should be longer than '0x'")

	t.Logf("Successfully signed transaction with Web3Signer")
	t.Logf("From address: %s", fromAddress.Hex())
	t.Logf("Signed transaction: %s", signedTxHex)
}

// TestPrivateKeySignerIntegration tests that the private key signer works correctly
func TestPrivateKeySignerIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Use a test private key (same as used in web3signer tests)
	privateKeyHex := "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"

	// Create signing context (without real ethClient for this test)
	signingContext := &SigningContext{
		ethClient: nil, // We're only testing the signing part, not transaction sending
		logger:    logger,
		chainID:   big.NewInt(31337),
	}

	// Create private key signer
	signer, err := NewPrivateKeySigner(privateKeyHex, signingContext)
	require.NoError(t, err)

	// Test GetFromAddress
	expectedAddress := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	assert.Equal(t, expectedAddress, signer.GetFromAddress())

	// Test GetTransactOpts
	ctx := context.Background()
	opts, err := signer.GetTransactOpts(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedAddress, opts.From)
	assert.True(t, opts.NoSend)
	assert.Equal(t, ctx, opts.Context)

	t.Logf("Successfully created private key signer")
	t.Logf("From address: %s", expectedAddress.Hex())
}
