package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			"eigenlayer", "register-operator",
			"--metadata-uri", "https://example.com/operator/metadata.json",
			"--allocation-delay", allocationDelay,
		)

		// Check execution succeeded
		if err != nil || result.ExitCode != 0 {
			t.Logf("Command failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}
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
			"eigenlayer", "set-allocation-delay",
			"--delay", "0")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "allocation delay")
		assert.Contains(t, result.Stdout, "0")
	})

	t.Run("Set Invalid Allocation Delay", func(t *testing.T) {
		// Try to set allocation delay to -1 for aggregator (should fail)
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"eigenlayer", "set-allocation-delay",
			"--delay", "-1")

		require.NoError(t, err)
		assert.Equal(t, 1, result.ExitCode)
	})

	t.Run("Delegate to Specific Operator", func(t *testing.T) {
		// Get aggregator address to delegate to
		aggregatorKeystore := h.GetAggregatorECDSAKeystore()

		// Try to delegate executor to aggregator (executor needs to be registered first)
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"eigenlayer", "register-operator",
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
				"eigenlayer", "delegate", "--operator", aggregatorKeystore.Address)

			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
		}
	})

	t.Run("Register ECDSA Key", func(t *testing.T) {
		// The aggregator operator should already be registered

		// Set operator-set-id in context first
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-set-id", "0",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// First remove existing system signer configuration
		removeResult, err := h.ExecuteCLI("signer", "system", "remove")
		require.NoError(t, err)
		assert.Equal(t, 0, removeResult.ExitCode)

		// Configure system signer with ECDSA using private key
		systemPrivateKey := h.ChainConfig.OperatorAccountPk
		if !strings.HasPrefix(systemPrivateKey, "0x") {
			systemPrivateKey = "0x" + systemPrivateKey
		}
		t.Setenv("SYSTEM_PRIVATE_KEY", systemPrivateKey)

		// Also set the SYSTEM_ECDSA_ADDRESS which is required when using private key
		t.Setenv("SYSTEM_ECDSA_ADDRESS", h.ChainConfig.OperatorAccountAddress)

		signerResult, err := h.ExecuteCLI("signer", "system", "privatekey")
		require.NoError(t, err)
		assert.Equal(t, 0, signerResult.ExitCode)

		// Register ECDSA key (no flags needed - uses context)
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"eigenlayer", "register-key")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "registered key")
	})

	t.Run("Register_Executor_BN254_Key", func(t *testing.T) {
		// Register executor's BN254 key to a different operator set
		executorBN254 := h.GetExecutorBN254Keystore()

		// Set context for operator set 1
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-address",
			h.ChainConfig.ExecOperatorAccountAddress,
			"--operator-set-id", "1",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		_, err = h.ExecuteCLI("signer", "system", "remove")
		require.NoError(t, err)

		// Configure system signer with BN254 keystore
		// This assumes the keystore name has been configured
		_, err = h.ExecuteCLI(
			"signer", "system", "keystore",
			"--name", "executor",
			"--type", "bn254",
		)
		require.NoError(t, err)
		// Set the system keystore password
		t.Setenv("SYSTEM_KEYSTORE_PASSWORD", executorBN254.Password)

		// Register BN254 key (no flags needed - uses context)
		result, _ := h.ExecuteCLIWithKeystore("executor-bn254",
			"el", "register-key")

		// This might fail if executor is not registered as operator
		// but we're testing the BN254 key registration flow
		if result.ExitCode == 0 {
			assert.Contains(t, result.Stdout, "Successfully registered key")
		} else {
			// Executor might not be registered as operator yet
			t.Logf("Executor BN254 registration result: %s", result.Stdout)
			t.Fail()
		}
	})

	strategyAddress := h.GetBeaconETHStrategy()

	t.Run("Full Allocation", func(t *testing.T) {
		// Ensure we have passed the allocation delay.
		err := h.MineBlocks(activateAllocationDelayBlocks)
		require.NoError(t, err)

		// Set context with operator-set-id
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-address",
			h.ChainConfig.OperatorAccountAddress,
			"--operator-set-id", "0",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Allocate (no operator-set-id flag needed - uses context)
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"eigenlayer", "allocate",
			"--strategy", strategyAddress,
			"--magnitude", "1e18")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully modified allocations")
	})

	t.Run("Register to Single Operator Set", func(t *testing.T) {
		// Set operator-set-id in context (should already be 0 from previous test)
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-set-id", "0",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Register AVS (no operator-set-ids flag needed - uses context)
		result, err := h.ExecuteCLIWithKeystore("aggregator-ecdsa",
			"eigenlayer", "register-avs",
			"--socket", "https://operator.example.com:8080")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully registered operator with AVS")
	})
}

// Helper function to skip tests in short mode
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}
