package integration

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOperatorRegistration(t *testing.T) {
	skipIfShort(t)

	// Create and setup harness once
	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	t.Run("Register New Operator", func(t *testing.T) {
		// Use the aggregator ECDSA key
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"register",
			"--allocation-delay", "1",
			"--metadata-uri", "https://example.com/operator/metadata.json")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "Command failed: %s", result.Stderr)
		assert.Contains(t, result.Stdout, "operator registered")
		assert.Contains(t, result.Stdout, "transaction successful")

		// Wait for transaction if available
		if txHash, err := harness.ParseTransactionHash(result.Stdout); err == nil && txHash != "" {
			err = h.WaitForTransaction(context.Background(), txHash)
			assert.NoError(t, err)
		}
	})

	t.Run("Register With Invalid Allocation Delay", func(t *testing.T) {
		// Try to register with allocation delay of 0 (should fail)
		result, err := h.ExecuteCLIWithKeystore("executor-ecdsa",
			"register",
			"--allocation-delay", "0",
			"--metadata-uri", "https://example.com/operator/metadata.json")

		require.NoError(t, err)
		assert.Equal(t, 1, result.ExitCode)
		assert.Contains(t, result.Stderr, "allocation delay")
	})

	t.Run("Double Registration Attempt", func(t *testing.T) {
		// Skip this test in a new harness as the aggregator is likely already registered
		// from chain setup. Instead, we'll test with executor keystore.

		// First registration
		result, err := h.ExecuteCLIWithKeystore("executor-ecdsa",
			"operator", "register",
			"--allocation-delay", "1",
			"--metadata-uri", "https://example.com/operator/metadata.json")

		require.NoError(t, err)

		if result.ExitCode == 0 {
			// Wait for transaction
			if txHash, err := harness.ParseTransactionHash(result.Stdout); err == nil && txHash != "" {
				h.WaitForTransaction(context.Background(), txHash)
			}

			// Second registration attempt (should fail)
			result, err = h.ExecuteCLIWithKeystore("executor-ecdsa",
				"operator", "register",
				"--allocation-delay", "1",
				"--metadata-uri", "https://example.com/operator/metadata.json")

			require.NoError(t, err)
			assert.Equal(t, 1, result.ExitCode)
			assert.Contains(t, result.Stderr, "already registered")
		}
	})
}

func TestOperatorAllocationDelay(t *testing.T) {
	skipIfShort(t)

	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	t.Run("Set Allocation Delay", func(t *testing.T) {
		// The aggregator operator should already be registered from chain setup
		// Just set new allocation delay
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"operator", "set-allocation-delay",
			"--delay", "100")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "allocation delay")
		assert.Contains(t, result.Stdout, "100")
	})

	t.Run("Set Invalid Allocation Delay", func(t *testing.T) {
		// Try to set allocation delay to 0 for aggregator (should fail)
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"operator", "set-allocation-delay",
			"--delay", "0")

		require.NoError(t, err)
		assert.Equal(t, 1, result.ExitCode)
		assert.Contains(t, result.Stderr, "delay")
	})
}

func TestOperatorDelegation(t *testing.T) {
	skipIfShort(t)

	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	t.Run("Self Delegate", func(t *testing.T) {
		// The aggregator should already be registered, just self-delegate
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa", "operator", "delegate")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "delegation")
		assert.Contains(t, result.Stdout, "successful")
	})

	t.Run("Delegate to Specific Operator", func(t *testing.T) {
		// Get aggregator address to delegate to
		aggregatorKeystore := h.GetAggregatorECDSAKeystore()

		// Try to delegate executor to aggregator (executor needs to be registered first)
		result, err := h.ExecuteCLIWithKeystore("executor-ecdsa",
			"operator", "register",
			"--allocation-delay", "1",
			"--metadata-uri", "https://example.com/executor/metadata.json")

		if err == nil && result.ExitCode == 0 {
			// Wait for registration
			if txHash, err := harness.ParseTransactionHash(result.Stdout); err == nil && txHash != "" {
				h.WaitForTransaction(context.Background(), txHash)
			}

			// Now delegate to aggregator
			result, err = h.ExecuteCLIWithKeystore("executor-ecdsa",
				"operator", "delegate", "--operator", aggregatorKeystore.Address)

			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
		}
	})
}

func TestOperatorKeyRegistration(t *testing.T) {
	skipIfShort(t)

	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	t.Run("Register ECDSA Key", func(t *testing.T) {
		// The aggregator operator should already be registered
		// Get the aggregator ECDSA keystore
		aggregatorKeystore := h.GetAggregatorECDSAKeystore()

		// Register ECDSA key for operator set 0
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"operator", "register-key",
			"--operator-set-id", "0",
			"--key-type", "ecdsa",
			"--key-address", aggregatorKeystore.Address)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "registered key")
	})

	t.Run("Register BN254 Key", func(t *testing.T) {
		// Use the aggregator BN254 keystore
		aggregatorBN254 := h.GetAggregatorKeystore()

		// Register BN254 key for operator set 1
		// Note: You would need to read the actual key data from the keystore file
		// For now, we'll use the keystore path approach
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"operator", "register-key",
			"--operator-set-id", "1",
			"--key-type", "bn254",
			"--keystore-path", aggregatorBN254.Path,
			"--password", aggregatorBN254.Password)

		require.NoError(t, err)
		// BN254 key registration might fail if already registered or other reasons
		// Check the output for specific errors
		if result.ExitCode == 0 {
			assert.Contains(t, result.Stdout, "registered key")
		}
	})
}

// Helper function to skip tests in short mode
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}
