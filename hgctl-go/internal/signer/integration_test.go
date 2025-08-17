package signer

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Web3SignerClient L1 port (matches goTest.sh configuration)
	web3signerL1Port = "9100"
)

// TestWeb3SignerIntegration tests the Web3SignerClient client against a real Web3SignerClient service
// This test expects that Web3SignerClient containers are already running via goTest.sh
func TestWeb3SignerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create client for the L1 Web3SignerClient service (started by goTest.sh)
	client := createTestClient(t)

	// Wait for service to be ready
	err := waitForService(client, 30*time.Second)
	require.NoError(t, err, "Web3SignerClient service failed to start")

	// Get the available accounts for use in tests
	ctx := context.Background()
	accounts, err := client.EthAccounts(ctx)
	require.NoError(t, err, "Failed to get accounts for testing")
	require.NotEmpty(t, accounts, "No accounts available for testing")

	// Run integration tests
	t.Run("EthAccounts", func(t *testing.T) {
		testEthAccountsIntegration(t, client)
	})

	t.Run("EthSign", func(t *testing.T) {
		testEthSignIntegration(t, client, accounts[0])
	})

	t.Run("EthSignTransaction", func(t *testing.T) {
		testEthSignTransactionIntegration(t, client, accounts[0])
	})

}

// testEthAccountsIntegration tests listing accounts from the Web3SignerClient service
func testEthAccountsIntegration(t *testing.T, client *client.Web3SignerClient) {
	ctx := context.Background()

	accounts, err := client.EthAccounts(ctx)
	require.NoError(t, err, "Failed to list accounts")

	// Should have accounts available
	assert.NotEmpty(t, accounts, "No accounts returned")

	// All accounts should be valid Ethereum addresses
	for _, account := range accounts {
		assert.True(t, len(account) == 42, "Account should be 42 characters long")
		assert.True(t, account[:2] == "0x", "Account should start with 0x")

		// Should be valid hex
		_, err := hex.DecodeString(account[2:])
		assert.NoError(t, err, "Account should be valid hex: %s", account)
	}

	t.Logf("Found %d accounts: %v", len(accounts), accounts)
}

// testEthSignIntegration tests signing a message with the Web3SignerClient service
func testEthSignIntegration(t *testing.T, client *client.Web3SignerClient, account string) {
	ctx := context.Background()

	// Test message to sign
	message := "Hello, Web3SignerClient!"
	messageHash := crypto.Keccak256Hash([]byte(message))
	dataToSign := hex.EncodeToString(messageHash[:])

	// Add 0x prefix as expected by Web3SignerClient
	dataToSign = "0x" + dataToSign

	// Sign with the account
	signature, err := client.EthSign(ctx, account, dataToSign)
	require.NoError(t, err, "Failed to sign message")

	// Verify signature format
	assert.NotEmpty(t, signature, "Signature is empty")
	assert.True(t, len(signature) > 0, "Signature should not be empty")

	// Signature should start with 0x
	assert.True(t, signature[:2] == "0x", "Signature should start with 0x")

	// Signature should be valid hex
	_, err = hex.DecodeString(signature[2:])
	assert.NoError(t, err, "Signature should be valid hex")

	t.Logf("Successfully signed message with account %s", account)
	t.Logf("Message: %s", message)
	t.Logf("Message hash: %s", dataToSign)
	t.Logf("Signature: %s", signature)
}

// testEthSignTransactionIntegration tests signing a transaction with the Web3SignerClient service
func testEthSignTransactionIntegration(t *testing.T, client *client.Web3SignerClient, account string) {
	ctx := context.Background()

	// Create a simple transaction with proper formatting
	transaction := map[string]interface{}{
		"to":       "0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23",
		"value":    "0x1",
		"gas":      "0x5208",
		"gasPrice": "0x9184e72a000", // Required field
		"nonce":    "0x0",
		"data":     "0x",
	}

	// Sign the transaction
	signature, err := client.EthSignTransaction(ctx, account, transaction)
	require.NoError(t, err, "Failed to sign transaction")

	// Verify signature format
	assert.NotEmpty(t, signature, "Signature is empty")
	assert.True(t, len(signature) > 0, "Signature should not be empty")

	// Signature should start with 0x
	assert.True(t, signature[:2] == "0x", "Signature should start with 0x")

	// Signature should be valid hex
	_, err = hex.DecodeString(signature[2:])
	assert.NoError(t, err, "Signature should be valid hex")

	t.Logf("Successfully signed transaction with account %s", account)
	t.Logf("Transaction: %+v", transaction)
	t.Logf("Signature: %s", signature)
}

// Helper functions

// createTestClient creates a Web3SignerClient client configured for testing
func createTestClient(t *testing.T) *client.Web3SignerClient {
	l := logger.NewLogger(false)

	config := client.DefaultWeb3SignerConfig()
	config.BaseURL = fmt.Sprintf("http://localhost:%s", web3signerL1Port)
	config.Timeout = 10 * time.Second

	c, err := client.NewWeb3Signer(config, l)
	require.NoError(t, err)
	return c
}

// waitForService waits for the Web3SignerClient service to be ready
func waitForService(c *client.Web3SignerClient, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Web3SignerClient service")
		case <-ticker.C:
			// Try to list accounts as a health check
			_, err := c.EthAccounts(ctx)
			if err == nil {
				return nil
			}

			// Also check if it's just a connection error
			var httpErr *client.Web3SignerError
			if errors.As(err, &httpErr) {
				if httpErr.Code == http.StatusServiceUnavailable {
					continue // Service is starting up
				}
			}
		}
	}
}
