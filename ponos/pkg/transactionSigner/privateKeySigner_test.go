package transactionSigner

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrivateKeySignerWithAnvil tests the privateKeySigner with a real anvil instance
func TestPrivateKeySignerWithAnvil(t *testing.T) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Setup logger
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	// Kill any existing anvil instances
	_ = killAllAnvils()

	// Start anvil - simple command without state file
	anvilCmd, err := startSimpleAnvil(ctx)
	require.NoError(t, err)
	defer func() {
		if anvilCmd != nil && anvilCmd.Process != nil {
			_ = anvilCmd.Process.Kill()
			_ = anvilCmd.Wait()
		}
	}()

	// Give anvil time to start
	time.Sleep(3 * time.Second)

	// Connect to anvil
	client, err := ethclient.Dial("http://localhost:8545")
	require.NoError(t, err)
	defer client.Close()

	// Use one of anvil's default funded accounts
	// Anvil provides 10 accounts with 10000 ETH each
	// First account private key: 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
	senderPrivateKey := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	// Create the signer
	signer, err := NewPrivateKeySigner(senderPrivateKey, client, l)
	require.NoError(t, err)
	assert.NotNil(t, signer)

	t.Run("SignAndSendTransaction", func(t *testing.T) {
		// Create a recipient address (second default anvil account)
		recipient := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")

		// Get the nonce for the sender
		nonce, err := client.PendingNonceAt(ctx, signer.GetFromAddress())
		require.NoError(t, err)

		// Create a simple ETH transfer transaction
		value := big.NewInt(1000000000000000) // 0.001 ETH in wei
		gasLimit := uint64(21000)             // Standard gas limit for ETH transfer
		gasPrice, err := client.SuggestGasPrice(ctx)
		require.NoError(t, err)

		// Create the transaction
		tx := types.NewTransaction(
			nonce,
			recipient,
			value,
			gasLimit,
			gasPrice,
			nil, // No data for simple ETH transfer
		)

		// Get initial sender balance
		senderBalanceBefore, err := client.BalanceAt(ctx, signer.GetFromAddress(), nil)
		require.NoError(t, err)

		// Sign and send the transaction
		receipt, err := signer.SignAndSendTransaction(ctx, tx)
		require.NoError(t, err)
		assert.NotNil(t, receipt)
		assert.Equal(t, uint64(1), receipt.Status) // Success status

		// Verify transaction was mined
		assert.NotNil(t, receipt.BlockNumber)
		assert.Greater(t, receipt.BlockNumber.Uint64(), uint64(0))

		// Verify gas was used
		assert.Greater(t, receipt.GasUsed, uint64(0))
		assert.LessOrEqual(t, receipt.GasUsed, gasLimit)

		// Verify sender balance decreased (due to gas fees at minimum)
		senderBalanceAfter, err := client.BalanceAt(ctx, signer.GetFromAddress(), nil)
		require.NoError(t, err)
		assert.True(t, senderBalanceAfter.Cmp(senderBalanceBefore) < 0, "Sender balance should decrease after transaction")

		t.Logf("Transaction successful! Hash: %s", receipt.TxHash.Hex())
		t.Logf("Gas used: %d", receipt.GasUsed)
	})

	t.Run("InvalidPrivateKey", func(t *testing.T) {
		// Test with invalid private key
		_, err := NewPrivateKeySigner("invalid-key", client, l)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})

	t.Run("MultipleTransactions", func(t *testing.T) {
		// Test sending multiple transactions in sequence
		recipient := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
		value := big.NewInt(100000000000000) // 0.0001 ETH

		for i := 0; i < 3; i++ {
			nonce, err := client.PendingNonceAt(ctx, signer.GetFromAddress())
			require.NoError(t, err)

			gasPrice, err := client.SuggestGasPrice(ctx)
			require.NoError(t, err)

			tx := types.NewTransaction(
				nonce,
				recipient,
				value,
				uint64(21000),
				gasPrice,
				nil,
			)

			receipt, err := signer.SignAndSendTransaction(ctx, tx)
			require.NoError(t, err)
			assert.Equal(t, uint64(1), receipt.Status)

			t.Logf("Transaction %d successful: %s", i+1, receipt.TxHash.Hex())
		}
	})
}

// startSimpleAnvil starts anvil for basic testing
// Uses anvil's default configuration with pre-funded accounts
func startSimpleAnvil(ctx context.Context) (*exec.Cmd, error) {
	// Start anvil with default settings (no fork, just local testing chain)
	// This gives us 10 accounts with 10000 ETH each for testing
	cmd := exec.CommandContext(ctx, "anvil", "--port", "8545")
	cmd.Stderr = os.Stderr

	// Optionally show anvil output for debugging
	if os.Getenv("ANVIL_DEBUG") == "true" {
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start anvil: %w", err)
	}

	// Wait for anvil to be ready
	rpcUrl := "http://localhost:8545"
	for i := 0; i < 10; i++ {
		res, err := http.Post(rpcUrl, "application/json", nil)
		if err == nil && res.StatusCode == 200 {
			res.Body.Close()
			return cmd, nil
		}
		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("anvil failed to start after 10 seconds")
}

// killAllAnvils kills all running anvil processes
func killAllAnvils() error {
	cmd := exec.Command("pkill", "-f", "anvil")
	// Ignore error as pkill returns non-zero if no processes found
	_ = cmd.Run()
	return nil
}
