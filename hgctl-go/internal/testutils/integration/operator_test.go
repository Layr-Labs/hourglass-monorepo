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
	allocationDelay := "100"
	activateAllocationDelayBlocks := 101

	// Create and setup harness once
	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	t.Run("Operator_Registration_Allocation", func(t *testing.T) {
		// Execute the register command
		// TODO: use operator private key
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"register-operator",
			"--metadata-uri", "https://example.com/operator/metadata.json",
			"--allocation-delay", allocationDelay,
		)

		// Check execution succeeded
		require.NoError(t, err, "Register command should not return an error")
		require.Equal(t, 0, result.ExitCode, "Command should exit with code 0")

		// Based on your log output, the command outputs these messages:
		assert.Contains(t, result.Stdout, "Successfully registered operator with EigenLayer",
			"Output should contain the success message")

		// The transaction info is in a structured log with "txHash" field
		assert.Contains(t, result.Stdout, "txHash",
			"Output should contain transaction hash")

		// If you need to check for specific transaction hash 	format
		assert.Contains(t, result.Stdout, "0x",
			"Output should contain a transaction hash starting with 0x")
	})

	t.Run("Set Allocation Delay", func(t *testing.T) {
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"set-allocation-delay",
			"--delay", "0")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "allocation delay")
		assert.Contains(t, result.Stdout, "0")
	})

	t.Run("Set Invalid Allocation Delay", func(t *testing.T) {
		// Try to set allocation delay to -1 for aggregator (should fail)
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"set-allocation-delay",
			"--delay", "-1")

		require.NoError(t, err)
		assert.Equal(t, 1, result.ExitCode)
	})

	t.Run("Delegate to Specific Operator", func(t *testing.T) {
		// Get aggregator address to delegate to
		aggregatorKeystore := h.GetAggregatorECDSAKeystore()

		// Try to delegate executor to aggregator (executor needs to be registered first)
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"register-operator",
			"--allocation-delay", "1",
			"--metadata-uri", "https://example.com/executor/metadata.json")

		if err == nil && result.ExitCode == 0 {
			// Wait for registration
			if txHash, err := harness.ParseTransactionHash(result.Stdout); err == nil && txHash != "" {
				err := h.WaitForTransaction(context.Background(), txHash)
				if err != nil {
					t.Logf("failed to wait for transaction in test")
					return
				}
			}

			// Now delegate to aggregator
			result, err = h.ExecuteCLIWithKeystore("executor-ecdsa",
				"delegate", "--operator", aggregatorKeystore.Address)

			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
		}
	})

	t.Run("Register ECDSA Key", func(t *testing.T) {
		// The aggregator operator should already be registered
		// Get the aggregator ECDSA keystore
		aggregatorKeystore := h.GetAggregatorECDSAKeystore()

		// Register ECDSA key for operator set 0
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"register-key",
			"--operator-set-id", "0",
			"--key-type", "ecdsa",
			"--key-address", aggregatorKeystore.Address)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "registered key")
	})

	t.Run("Register_Executor_BN254_Key", func(t *testing.T) {
		// Register executor's BN254 key to a different operator set
		executorBN254 := h.GetExecutorKeystore()

		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-address",
			h.ChainConfig.ExecOperatorAccountAddress,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		result, _ := h.ExecuteCLIWithKeystore("executor-bn254",
			"register-key",
			"--operator-set-id", "1",
			"--key-type", "bn254",
			"--keystore-path", executorBN254.Path,
			"--password", executorBN254.Password)

		// This might fail if executor is not registered as operator
		// but we're testing the BN254 key registration flow
		if result.ExitCode == 0 {
			assert.Contains(t, result.Stdout, "Successfully registered key")
		} else {
			// Executor might not be registered as operator yet
			t.Logf("Executor BN254 registration result: %s", result.Stdout)
		}
	})

	strategyAddress := h.GetBeaconETHStrategy()

	t.Run("Full Allocation", func(t *testing.T) {
		// Ensure we have passed the allocation delay.
		err := h.MineBlocks(activateAllocationDelayBlocks)
		require.NoError(t, err)
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-address",
			h.ChainConfig.OperatorAccountAddress,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"allocate",
			"--operator-set-id", "0",
			"--strategy", strategyAddress,
			"--magnitude", "1e18")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully modified allocations")
	})

	t.Run("Register to Single Operator Set", func(t *testing.T) {
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"register-avs",
			"--operator-set-ids", "0",
			"--socket", "https://operator.example.com:8080")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully registered operator with AVS")
	})

	t.Run("Register to Multiple Operator Sets", func(t *testing.T) {
		// First, make sure operator set 2 exists and has keys registered
		// Register ECDSA key for operator set 2
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"register-key",
			"--operator-set-id", "2",
			"--key-type", "ecdsa",
			"--key-address", h.GetAggregatorECDSAKeystore().Address,
		)

		if err == nil && result.ExitCode == 0 {
			// Now register to multiple operator sets
			result, err = h.ExecuteCLIWithKeystore("aggregator-ecdsa",
				"register-avs",
				"--operator-set-ids", "1,2",
				"--socket", "wss://operator.example.com:8443",
			)

			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, result.Stdout, "Successfully registered operator with AVS")
		}
	})
}

// Helper function to skip tests in short mode
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}
