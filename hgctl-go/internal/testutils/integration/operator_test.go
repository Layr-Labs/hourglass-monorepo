package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperatorRegistration(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)
	
	// For test chains, we'll proceed without checking for EigenLayer contracts
	// as we're testing with mock/test addresses

	t.Run("Register New Operator", func(t *testing.T) {
		// Create operator keystore
		keystoreName := fmt.Sprintf("operator-%d", time.Now().UnixNano())
		password := "test-password"

		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		if result.ExitCode != 0 {
			t.Logf("Keystore creation failed:\nStdout: %s\nStderr: %s", result.Stdout, result.Stderr)
		}
		require.Equal(t, 0, result.ExitCode)

		// Get the operator address from the keystore creation output
		// For now, use the test operator address from chain config
		operatorAddress := h.ChainConfig.OperatorAccountAddress

		// Register as operator using environment variables for RPC and private key
		result, err = h.ExecuteCLI(
			"register",
			"--address", operatorAddress,
			"--allocation-delay", "1")

		require.NoError(t, err)
		// Debug output
		if result.ExitCode != 0 {
			t.Logf("Register command failed:\nStdout: %s\nStderr: %s", result.Stdout, result.Stderr)
		}
		harness.Assert(t, result).
			AssertOperatorRegistered().
			OutputContainsAll("operator registered", "transaction successful")

		// Get transaction hash and wait for confirmation
		txHash, err := result.GetTransactionHash()
		if err == nil && txHash != "" {
			err = h.WaitForTransaction(context.Background(), txHash)
			assert.NoError(t, err)
		}

		// TODO: Verify on-chain registration when contract client methods are available
		// isOperator, err := h.Contracts.IsOperatorRegistered(context.Background(), operatorAddress)
		// require.NoError(t, err)
		// require.True(t, isOperator)
	})

	t.Run("Register With Custom Metadata URI", func(t *testing.T) {
		keystoreName := fmt.Sprintf("operator-metadata-%d", time.Now().UnixNano())
		password := "test-password"

		// Create keystore
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Register with metadata URI
		metadataURI := "https://example.com/operator/metadata.json"
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--allocation-delay", "1",
			"--metadata-uri", metadataURI)

		require.NoError(t, err)
		harness.Assert(t, result).
			AssertOperatorRegistered().
			OutputContainsAll("metadata", metadataURI)
	})

	t.Run("Register With Invalid Allocation Delay", func(t *testing.T) {
		keystoreName := fmt.Sprintf("operator-invalid-%d", time.Now().UnixNano())
		password := "test-password"

		// Create keystore
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Try to register with allocation delay of 0 (should fail)
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--allocation-delay", "0")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("allocation delay|invalid|must be greater than")
	})

	t.Run("Double Registration Attempt", func(t *testing.T) {
		keystoreName := fmt.Sprintf("operator-double-%d", time.Now().UnixNano())
		password := "test-password"

		// Create keystore
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// First registration
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--allocation-delay", "1")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Wait for transaction
		if txHash, err := result.GetTransactionHash(); err == nil && txHash != "" {
			h.WaitForTransaction(context.Background(), txHash)
		}

		// Second registration attempt (should fail)
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--allocation-delay", "1")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("already registered|operator exists")
	})
}

func TestOperatorAllocationDelay(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Set Allocation Delay", func(t *testing.T) {
		// First register an operator
		keystoreName := fmt.Sprintf("operator-delay-%d", time.Now().UnixNano())
		password := "test-password"

		// Create and register operator
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--allocation-delay", "1")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Wait for registration
		if txHash, err := result.GetTransactionHash(); err == nil && txHash != "" {
			h.WaitForTransaction(context.Background(), txHash)
		}

		// Set new allocation delay
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"set-allocation-delay",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--delay", "100")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("allocation delay", "set", "100").
			TransactionSucceeded()
	})

	t.Run("Set Invalid Allocation Delay", func(t *testing.T) {
		// Use existing operator
		keystoreName := fmt.Sprintf("operator-invalid-delay-%d", time.Now().UnixNano())
		password := "test-password"

		// Create and register operator
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--allocation-delay", "1")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Try to set allocation delay to 0
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"set-allocation-delay",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--delay", "0")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("invalid.*delay|must be greater")
	})
}

func TestOperatorDelegation(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Self Delegate", func(t *testing.T) {
		// Register operator
		keystoreName := fmt.Sprintf("operator-delegate-%d", time.Now().UnixNano())
		password := "test-password"

		// Create and register operator
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--allocation-delay", "1")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Wait for registration
		if txHash, err := result.GetTransactionHash(); err == nil && txHash != "" {
			h.WaitForTransaction(context.Background(), txHash)
		}

		// Self-delegate
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"delegate",
			"--context", h.ContextName,
			"--keystore", keystoreName)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("delegation", "successful").
			TransactionSucceeded()
	})

	t.Run("Delegate Before Registration", func(t *testing.T) {
		// Create keystore but don't register
		keystoreName := fmt.Sprintf("operator-no-reg-%d", time.Now().UnixNano())
		password := "test-password"

		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Try to delegate without registration
		// Note: delegate command uses private key from env, not keystore
		result, err = h.ExecuteCLI("delegate")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("not registered|operator not found|no contract code")
	})
}

func TestOperatorStatus(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Check Operator Status", func(t *testing.T) {
		// Register an operator first
		keystoreName := fmt.Sprintf("operator-status-%d", time.Now().UnixNano())
		password := "test-password"

		// Create and register operator
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--allocation-delay", "1")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Get operator address (simplified - would parse from output)
		// operatorAddress := extractAddressFromOutput(result.Stdout)

		// Check status
		result, err = h.ExecuteCLI("operator", "status",
			"--context", h.ContextName,
			"--keystore", keystoreName)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("registered", "allocation delay")
	})

}

// Helper function to setup a registered operator for other tests
func setupRegisteredOperator(t *testing.T, h *harness.TestHarness, name string) (string, string) {
	password := "test-password"

	// Create keystore
	result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
		"keystore", "create",
		"--context", h.ContextName,
		"--name", name,
		"--type", "ecdsa")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Register operator
	// The register command doesn't use keystore, it uses address directly
	result, err = h.ExecuteCLI(
		"register",
		"--address", h.ChainConfig.OperatorAccountAddress,
		"--allocation-delay", "1")
	require.NoError(t, err)
	if result.ExitCode != 0 {
		t.Logf("Register command failed in setupRegisteredOperator:\nStdout: %s\nStderr: %s", result.Stdout, result.Stderr)
	}
	require.Equal(t, 0, result.ExitCode)

	// Wait for transaction
	if txHash, err := result.GetTransactionHash(); err == nil && txHash != "" {
		_ = h.WaitForTransaction(context.Background(), txHash)
	}

	// Return keystore name and password
	return name, password
}
