package integration

import (
	"context"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/tools"
	"github.com/ethereum/go-ethereum/common"

	//"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/tools"
	//"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
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
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
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

		// If you need to check for specific transaction hash format
		assert.Contains(t, result.Stdout, "0x",
			"Output should contain a transaction hash starting with 0x")
	})

	t.Run("Set Allocation Delay", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "set-allocation-delay",
			"--delay", "0")

		if err != nil || result.ExitCode != 0 {
			t.Logf("Command failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "allocation delay")
		assert.Contains(t, result.Stdout, "0")
	})

	t.Run("Set Invalid Allocation Delay", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "set-allocation-delay",
			"--delay", "-1")

		require.NoError(t, err)
		assert.Equal(t, 1, result.ExitCode)
	})

	t.Run("Delegate to Specific Operator", func(t *testing.T) {
		t.Skip("Revisit with staker and delegation tooling")
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered2ECDSA,
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

			// Now delegate operator 2 to operator 1
			unregOp1 := h.ChainConfig.UnregisteredOperator1AccountAddress
			result, err = h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered2ECDSA,
				"eigenlayer", "delegate", "--operator", unregOp1)

			if err != nil || result.ExitCode != 0 {
				t.Logf("Command failed with error: %v", err)
				t.Logf("Exit code: %d", result.ExitCode)
				t.Logf("Stdout: %s", result.Stdout)
				t.Logf("Stderr: %s", result.Stderr)
			}

			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			return
		}

		t.Logf("Command failed with error: %v", err)
		t.Logf("Exit code: %d", result.ExitCode)
		t.Logf("Stdout: %s", result.Stdout)
		t.Logf("Stderr: %s", result.Stderr)
	})

	t.Run("Register ECDSA Key", func(t *testing.T) {
		// Set context with unregistered operator 1's address
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-address", h.ChainConfig.UnregisteredOperator1AccountAddress,
			"--operator-set-id", "1",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Configure system signer with unregistered operator 1's system ECDSA keystore
		err = h.ConfigureSystemKey(harness.KeystoreUnregistered1SystemECDSA)
		require.NoError(t, err)

		// Register ECDSA key (no flags needed - uses context)
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "register-key")

		if err != nil || result.ExitCode != 0 {
			t.Logf("Command failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "registered key")
	})

	t.Run("Register_Executor_BN254_Key", func(t *testing.T) {
		t.Skip("BN254 operator set not configured")
		// Set context for unregistered operator 2 with operator set 1
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-address", h.ChainConfig.UnregisteredOperator2AccountAddress,
			"--operator-set-id", "1",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Set system keystore password for the test duration
		systemKeystore := h.ChainConfig.UnregisteredOperator2SystemBN254KeystorePassword
		t.Setenv("SYSTEM_KEYSTORE_PASSWORD", systemKeystore)

		// Configure system signer with unregistered operator 2's BN254 system keystore
		err = h.ConfigureSystemKey(harness.KeystoreUnregistered2SystemBN254)
		require.NoError(t, err)

		// Register BN254 key (no flags needed - uses context)
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered2ECDSA,
			"el", "register-key")

		if err != nil || result.ExitCode != 0 {
			t.Logf("Command failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully registered key")
	})

	strategyAddress := h.GetBeaconETHStrategy()

	t.Run("Full Allocation", func(t *testing.T) {
		// Ensure we have passed the allocation delay.
		err := h.MineBlocks(activateAllocationDelayBlocks)
		require.NoError(t, err)

		// Set context with unregistered operator 1's address
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-address", h.ChainConfig.UnregisteredOperator1AccountAddress,
			"--operator-set-id", "1",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Initialize AllocationDelayInfo storage to avoid UninitializedAllocationDelay error
		ctx := context.Background()
		header, err := h.L1Client.HeaderByNumber(ctx, nil)
		require.NoError(t, err)
		currentBlock := uint32(header.Number.Uint64())

		rpcClient, err := rpc.Dial(h.ChainConfig.L1RPC)
		require.NoError(t, err)
		defer rpcClient.Close()

		allocationManagerAddr := common.HexToAddress(h.ChainConfig.AllocationManagerAddress)
		operatorAddr := common.HexToAddress(h.ChainConfig.UnregisteredOperator1AccountAddress)

		err = tools.InitializeAllocationDelay(rpcClient, allocationManagerAddr, operatorAddr, currentBlock)
		require.NoError(t, err)

		// Allocate (no operator-set-id flag needed - uses context)
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "allocate",
			"--strategy", strategyAddress,
			"--magnitude", "1e18")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully modified allocations")
	})

	t.Run("Register to Single Operator Set", func(t *testing.T) {
		// Set context with unregistered operator 1 for operator set 1
		contextResult, err := h.ExecuteCLI(
			"context",
			"set",
			"--operator-address", h.ChainConfig.UnregisteredOperator1AccountAddress,
			"--operator-set-id", "1",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Register AVS (no operator-set-ids flag needed - uses context)
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "register-avs",
			"--socket", "https://operator.example.com:8080")

		if err != nil || result.ExitCode != 0 {
			t.Logf("Command failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}

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
