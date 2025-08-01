package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/require"
)

func TestKeyRegistration(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Register BN254 Key for Aggregator", func(t *testing.T) {
		// Setup: Register operator first
		operatorName := fmt.Sprintf("key-reg-operator-%d", time.Now().UnixNano())
		operatorName, _ = setupRegisteredOperator(t, h, operatorName)

		// Create BN254 keystore for signing
		bn254Name := fmt.Sprintf("bn254-key-%d", time.Now().UnixNano())
		bn254Password := "bn254-password"
		result, err := h.ExecuteCLIWithInput(bn254Password+"\n"+bn254Password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", bn254Name,
			"--type", "bn254")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Get operator address (would parse from keystore or previous output)
		// For now, use a placeholder approach
		operatorAddress := h.ChainConfig.OperatorAccountAddress

		// Register BN254 key for aggregator role (operator set 0)
		result, err = h.ExecuteCLIWithInput(bn254Password+"\n",
			"register-key",
			"--context", h.ContextName,
			"--keystore", bn254Name,
			"--operator-address", operatorAddress,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0")

		require.NoError(t, err)
		harness.Assert(t, result).
			AssertKeyRegistered().
			OutputContainsAll("key registered", "operator set 0")

		// TODO: Verify key registration on-chain
		// isRegistered, err := h.Contracts.IsKeyRegistered(...)
	})

	t.Run("Register Key for Multiple Operator Sets", func(t *testing.T) {
		// Setup operator
		operatorName := fmt.Sprintf("multi-set-operator-%d", time.Now().UnixNano())
		operatorName, _ = setupRegisteredOperator(t, h, operatorName)

		// Create BN254 keystore
		bn254Name := fmt.Sprintf("multi-bn254-%d", time.Now().UnixNano())
		bn254Password := "bn254-password"
		result, err := h.ExecuteCLIWithInput(bn254Password+"\n"+bn254Password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", bn254Name,
			"--type", "bn254")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		operatorAddress := h.ChainConfig.OperatorAccountAddress

		// Register for operator set 0
		result, err = h.ExecuteCLIWithInput(bn254Password+"\n",
			"register-key",
			"--context", h.ContextName,
			"--keystore", bn254Name,
			"--operator-address", operatorAddress,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Register for operator set 1
		result, err = h.ExecuteCLIWithInput(bn254Password+"\n",
			"register-key",
			"--context", h.ContextName,
			"--keystore", bn254Name,
			"--operator-address", operatorAddress,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "1")
		require.NoError(t, err)
		harness.Assert(t, result).
			AssertKeyRegistered().
			OutputContainsAll("key registered", "operator set 1")
	})

	t.Run("Invalid Operator Set ID", func(t *testing.T) {
		// Setup
		operatorName := fmt.Sprintf("invalid-set-operator-%d", time.Now().UnixNano())
		operatorName, _ = setupRegisteredOperator(t, h, operatorName)

		bn254Name := fmt.Sprintf("invalid-bn254-%d", time.Now().UnixNano())
		bn254Password := "bn254-password"
		result, err := h.ExecuteCLIWithInput(bn254Password+"\n"+bn254Password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", bn254Name,
			"--type", "bn254")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Try to register for non-existent operator set
		result, err = h.ExecuteCLIWithInput(bn254Password+"\n",
			"register-key",
			"--context", h.ContextName,
			"--keystore", bn254Name,
			"--operator-address", h.ChainConfig.OperatorAccountAddress,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "999")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("operator set.*not found|invalid.*operator set")
	})
}

func TestAVSRegistration(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Register Operator to AVS", func(t *testing.T) {
		// Complete operator setup
		operatorName := fmt.Sprintf("avs-reg-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupCompleteOperator(t, h, operatorName)

		// Register with AVS
		socketInfo := "http://operator1.example.com:8080"
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"register-avs",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-ids", "0",
			"--socket", socketInfo)

		require.NoError(t, err)
		harness.Assert(t, result).
			AssertAVSRegistered().
			OutputContainsAll("registered to AVS", socketInfo)

		// TODO: Verify registration on-chain
		// isRegistered, err := h.Contracts.IsOperatorRegisteredToAVS(...)
	})

	t.Run("Register to Multiple Operator Sets", func(t *testing.T) {
		// Complete operator setup with keys for multiple sets
		operatorName := fmt.Sprintf("multi-avs-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupCompleteOperatorMultiSet(t, h, operatorName)

		// Register with AVS for multiple operator sets
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"register-avs",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-ids", "0,1",
			"--socket", "http://multi-operator.example.com:8080")

		require.NoError(t, err)
		harness.Assert(t, result).
			AssertAVSRegistered().
			OutputContainsAll("operator set 0", "operator set 1")
	})

	t.Run("Update Socket Information", func(t *testing.T) {
		// Setup operator registered to AVS
		operatorName := fmt.Sprintf("update-socket-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupCompleteOperator(t, h, operatorName)

		// Initial registration
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"register-avs",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-ids", "0",
			"--socket", "http://old-socket.example.com:8080")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Update socket
		newSocket := "http://new-socket.example.com:9090"
		result, err = h.ExecuteCLIWithInput(operatorPassword+"\n",
			"update-socket",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--socket", newSocket)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("socket updated", newSocket).
			TransactionSucceeded()
	})

	t.Run("Deregister from AVS", func(t *testing.T) {
		// Setup and register operator
		operatorName := fmt.Sprintf("dereg-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupCompleteOperator(t, h, operatorName)

		// Register first
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"register-avs",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-ids", "0",
			"--socket", "http://dereg-test.example.com:8080")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Wait for registration
		if txHash, err := result.GetTransactionHash(); err == nil && txHash != "" {
			h.WaitForTransaction(context.Background(), txHash)
		}

		// Deregister
		result, err = h.ExecuteCLIWithInput(operatorPassword+"\n",
			"deregister-avs",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-ids", "0")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("deregistered from AVS").
			TransactionSucceeded()
	})
}

func TestAVSStatus(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Check AVS Status", func(t *testing.T) {
		// Get AVS information
		result, err := h.ExecuteCLI("avs", "status",
			"--context", h.ContextName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("AVS Status", "operator sets")
	})

	t.Run("List Operators in AVS", func(t *testing.T) {
		result, err := h.ExecuteCLI("avs", "operators",
			"--context", h.ContextName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputMatches("OPERATOR.*SOCKET.*STATUS")
	})

	t.Run("Check Operator Set Configuration", func(t *testing.T) {
		result, err := h.ExecuteCLI("avs", "operator-sets",
			"--context", h.ContextName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("Operator Set", "ID", "Status")

		// Verify we have at least operator sets 0 and 1
		harness.Assert(t, result).
			OutputContainsAll("0", "1")
	})
}

// Helper functions for AVS tests

func setupCompleteOperator(t *testing.T, h *harness.TestHarness, name string) (string, string) {
	// Create and register operator
	keystoreName, password := setupRegisteredOperator(t, h, name)

	// Create BN254 keystore
	bn254Name := name + "-bn254"
	bn254Password := "bn254-password"
	result, err := h.ExecuteCLIWithInput(bn254Password+"\n"+bn254Password+"\n",
		"keystore", "create",
		"--context", h.ContextName,
		"--name", bn254Name,
		"--type", "bn254")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Register key
	result, err = h.ExecuteCLIWithInput(bn254Password+"\n",
		"register-key",
		"--context", h.ContextName,
		"--keystore", bn254Name,
		"--operator-address", h.ChainConfig.OperatorAccountAddress,
		"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
		"--operator-set-id", "0")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	return keystoreName, password
}

func setupCompleteOperatorMultiSet(t *testing.T, h *harness.TestHarness, name string) (string, string) {
	// Create and register operator
	keystoreName, password := setupRegisteredOperator(t, h, name)

	// Create BN254 keystore
	bn254Name := name + "-bn254"
	bn254Password := "bn254-password"
	result, err := h.ExecuteCLIWithInput(bn254Password+"\n"+bn254Password+"\n",
		"keystore", "create",
		"--context", h.ContextName,
		"--name", bn254Name,
		"--type", "bn254")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Register key for multiple operator sets
	for _, opSetID := range []string{"0", "1"} {
		result, err = h.ExecuteCLIWithInput(bn254Password+"\n",
			"register-key",
			"--context", h.ContextName,
			"--keystore", bn254Name,
			"--operator-address", h.ChainConfig.OperatorAccountAddress,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", opSetID)
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)
	}

	return keystoreName, password
}
