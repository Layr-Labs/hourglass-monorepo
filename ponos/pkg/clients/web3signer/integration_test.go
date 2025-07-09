package web3signer

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	// Web3Signer L1 port (matches goTest.sh configuration)
	web3signerL1Port = "9100"
)

// TestWeb3SignerIntegration tests the Web3Signer client against a real Web3Signer service
// This test expects that Web3Signer containers are already running via goTest.sh
func TestWeb3SignerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create client for the L1 Web3Signer service (started by goTest.sh)
	client := createTestClient(t)

	// Wait for service to be ready
	err := waitForService(client, 30*time.Second)
	require.NoError(t, err, "Web3Signer service failed to start")

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

	t.Run("HealthCheck", func(t *testing.T) {
		testHealthCheckIntegration(t, client)
	})

	t.Run("ReloadKeys", func(t *testing.T) {
		testReloadKeysIntegration(t, client)
	})
}

// testEthAccountsIntegration tests listing accounts from the Web3Signer service
func testEthAccountsIntegration(t *testing.T, client *Client) {
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

// testEthSignIntegration tests signing a message with the Web3Signer service
func testEthSignIntegration(t *testing.T, client *Client, account string) {
	ctx := context.Background()

	// Test message to sign
	message := "Hello, Web3Signer!"
	messageHash := crypto.Keccak256Hash([]byte(message))
	dataToSign := hex.EncodeToString(messageHash[:])

	// Add 0x prefix as expected by Web3Signer
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

// testEthSignTransactionIntegration tests signing a transaction with the Web3Signer service
func testEthSignTransactionIntegration(t *testing.T, client *Client, account string) {
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

// testHealthCheckIntegration tests the health check endpoint
func testHealthCheckIntegration(t *testing.T, client *Client) {
	ctx := context.Background()

	health, err := client.HealthCheck(ctx)
	require.NoError(t, err, "Failed to perform health check")

	assert.NotNil(t, health, "Health check result should not be nil")
	assert.Equal(t, "UP", health.Status, "Service should be UP")
	assert.Equal(t, "UP", health.Outcome, "Health outcome should be UP")
	assert.NotEmpty(t, health.Checks, "Health checks should not be empty")

	t.Logf("Health check successful: Status=%s, Outcome=%s, Checks=%d",
		health.Status, health.Outcome, len(health.Checks))
}

// testReloadKeysIntegration tests the key reload functionality
func testReloadKeysIntegration(t *testing.T, client *Client) {
	ctx := context.Background()

	err := client.Reload(ctx)
	require.NoError(t, err, "Failed to reload keys")

	t.Logf("Key reload successful")
}

// Helper functions

// createTestClient creates a Web3Signer client configured for testing
func createTestClient(t *testing.T) *Client {
	logger := zaptest.NewLogger(t)

	config := &Config{
		BaseURL: fmt.Sprintf("http://localhost:%s", web3signerL1Port),
		Timeout: 10 * time.Second,
	}

	client := NewClient(config, logger)
	return client
}

// waitForService waits for the Web3Signer service to be ready
func waitForService(client *Client, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Web3Signer service")
		case <-ticker.C:
			// Try to list accounts as a health check
			_, err := client.EthAccounts(ctx)
			if err == nil {
				return nil
			}

			// Also check if it's just a connection error
			if httpErr, ok := err.(*Web3SignerError); ok {
				if httpErr.Code == http.StatusServiceUnavailable {
					continue // Service is starting up
				}
			}
		}
	}
}
