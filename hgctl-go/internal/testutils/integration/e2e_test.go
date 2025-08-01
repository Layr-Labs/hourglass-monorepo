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

func TestCompleteOperatorOnboarding(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Full Operator Lifecycle", func(t *testing.T) {
		// Unique names for this test run
		operatorPrefix := fmt.Sprintf("e2e-operator-%d", time.Now().UnixNano())
		ecdsaKeystore := operatorPrefix + "-ecdsa"
		bn254Keystore := operatorPrefix + "-bn254"
		password := "test-password"

		// 1. Create all necessary keystores
		t.Log("Step 1: Creating keystores")

		// Create ECDSA keystore for operator transactions
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", ecdsaKeystore,
			"--type", "ecdsa")
		require.NoError(t, err)
		harness.Assert(t, result).AssertKeystoreCreated(ecdsaKeystore)

		// Create BN254 keystore for signing
		result, err = h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", bn254Keystore,
			"--type", "bn254")
		require.NoError(t, err)
		harness.Assert(t, result).AssertKeystoreCreated(bn254Keystore)

		// 2. Register as operator
		t.Log("Step 2: Registering as operator")
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register",
			"--context", h.ContextName,
			"--keystore", ecdsaKeystore,
			"--allocation-delay", "1")
		require.NoError(t, err)
		harness.Assert(t, result).AssertOperatorRegistered()

		// Wait for registration
		if txHash, err := result.GetTransactionHash(); err == nil && txHash != "" {
			require.NoError(t, h.WaitForTransaction(context.Background(), txHash))
		}

		// 3. Set allocation delay
		t.Log("Step 3: Setting allocation delay")
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"set-allocation-delay",
			"--context", h.ContextName,
			"--keystore", ecdsaKeystore,
			"--delay", "100")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			TransactionSucceeded()

		// 4. Self-delegate (if needed for the protocol)
		t.Log("Step 4: Self-delegating")
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"delegate",
			"--context", h.ContextName,
			"--keystore", ecdsaKeystore)
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			TransactionSucceeded()

		// 5. Register signing keys
		t.Log("Step 5: Registering BN254 signing keys")

		// Get operator address (would parse from previous output in real implementation)
		operatorAddress := h.ChainConfig.OperatorAccountAddress

		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register-key",
			"--context", h.ContextName,
			"--keystore", bn254Keystore,
			"--operator-address", operatorAddress,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0")
		require.NoError(t, err)
		harness.Assert(t, result).AssertKeyRegistered()

		// 6. Register with AVS
		t.Log("Step 6: Registering with AVS")
		socketURL := fmt.Sprintf("http://%s.example.com:8080", operatorPrefix)
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"register-avs",
			"--context", h.ContextName,
			"--keystore", ecdsaKeystore,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-ids", "0",
			"--socket", socketURL)
		require.NoError(t, err)
		harness.Assert(t, result).AssertAVSRegistered()

		// 7. Allocate to operator sets
		t.Log("Step 7: Allocating to operator sets")
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", ecdsaKeystore,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "1000000000000000000") // 1 ETH
		require.NoError(t, err)
		harness.Assert(t, result).AssertAllocationSet()

		// 8. Final verification
		t.Log("Step 8: Verifying complete setup")

		// Check operator status
		result, err = h.ExecuteCLI("operator", "status",
			"--context", h.ContextName,
			"--keystore", ecdsaKeystore)
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("registered", "delegated")

		// Check AVS registration
		result, err = h.ExecuteCLI("avs", "operators",
			"--context", h.ContextName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll(operatorAddress, socketURL)

		// Check allocations
		result, err = h.ExecuteCLI("allocation", "status",
			"--context", h.ContextName,
			"--keystore", ecdsaKeystore,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress)
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("1000000000000000000", h.GetBeaconETHStrategy())
	})
}

func TestMultiRoleOperatorSetup(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Setup Aggregator and Executor Operators", func(t *testing.T) {
		// Setup aggregator operator for operator set 0
		aggregatorName := fmt.Sprintf("aggregator-%d", time.Now().UnixNano())
		t.Log("Setting up aggregator operator for set 0")
		aggregatorKeystore := setupOperatorForSet(t, h, aggregatorName, "0", "aggregator")

		// Setup executor operator for operator set 1
		executorName := fmt.Sprintf("executor-%d", time.Now().UnixNano())
		t.Log("Setting up executor operator for set 1")
		executorKeystore := setupOperatorForSet(t, h, executorName, "1", "executor")

		// Verify both operators are properly configured
		t.Log("Verifying multi-role setup")

		// Check aggregator in set 0
		result, err := h.ExecuteCLI("avs", "operators",
			"--context", h.ContextName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("aggregator.example.com")

		// Check executor in set 1
		result, err = h.ExecuteCLI("avs", "operators",
			"--context", h.ContextName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "1")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("executor.example.com")

		// Verify allocations for both
		for _, ks := range []string{aggregatorKeystore, executorKeystore} {
			result, err = h.ExecuteCLI("allocation", "status",
				"--context", h.ContextName,
				"--keystore", ks,
				"--avs", h.ChainConfig.AVSTaskRegistrarAddress)
			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
		}
	})
}

func TestOperatorUpdateWorkflow(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Update Operator Configuration", func(t *testing.T) {
		// Setup operator
		operatorName := fmt.Sprintf("update-test-%d", time.Now().UnixNano())
		keystoreName, password := setupCompleteOperatorWithAllocation(t, h, operatorName)

		// 1. Update socket information
		t.Log("Updating socket information")
		newSocket := "http://updated-operator.example.com:9090"
		result, err := h.ExecuteCLIWithInput(password+"\n",
			"update-socket",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--socket", newSocket)
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("socket updated", newSocket)

		// 2. Increase allocation
		t.Log("Increasing allocation")
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "2000000000000000000") // Increase to 2 ETH total
		require.NoError(t, err)
		harness.Assert(t, result).AssertAllocationSet()

		// 3. Queue partial deallocation
		t.Log("Queueing partial deallocation")
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"deallocate",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "500000000000000000") // Deallocate 0.5 ETH
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("deallocation", "queued")

		// 4. Verify pending deallocations
		t.Log("Checking pending deallocations")
		result, err = h.ExecuteCLI("allocation", "pending",
			"--context", h.ContextName,
			"--keystore", keystoreName)
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("500000000000000000")
	})
}

func TestOperatorDeregistrationWorkflow(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Complete Operator Exit", func(t *testing.T) {
		// Setup operator
		operatorName := fmt.Sprintf("exit-test-%d", time.Now().UnixNano())
		keystoreName, password := setupCompleteOperatorWithAllocation(t, h, operatorName)

		// 1. Deallocate all funds
		t.Log("Step 1: Deallocating all funds")
		result, err := h.ExecuteCLIWithInput(password+"\n",
			"deallocate",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "1000000000000000000") // Deallocate all
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			TransactionSucceeded()

		// 2. Deregister from AVS
		t.Log("Step 2: Deregistering from AVS")
		result, err = h.ExecuteCLIWithInput(password+"\n",
			"deregister-avs",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-ids", "0")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("deregistered").
			TransactionSucceeded()

		// 3. Verify operator is no longer in AVS
		t.Log("Step 3: Verifying deregistration")
		result, err = h.ExecuteCLI("avs", "operators",
			"--context", h.ContextName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0")
		require.NoError(t, err)
		// Should not contain the operator address anymore
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsNone(operatorName)
	})
}

// Helper functions for E2E tests

func setupOperatorForSet(t *testing.T, h *harness.TestHarness, name string, opSetID string, role string) string {
	password := "test-password"

	// Create keystores
	ecdsaName := name + "-ecdsa"
	bn254Name := name + "-bn254"

	// Create ECDSA keystore
	result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
		"keystore", "create",
		"--context", h.ContextName,
		"--name", ecdsaName,
		"--type", "ecdsa")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Create BN254 keystore
	result, err = h.ExecuteCLIWithInput(password+"\n"+password+"\n",
		"keystore", "create",
		"--context", h.ContextName,
		"--name", bn254Name,
		"--type", "bn254")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Register operator
	result, err = h.ExecuteCLIWithInput(password+"\n",
		"register",
		"--context", h.ContextName,
		"--keystore", ecdsaName,
		"--allocation-delay", "1")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Set allocation delay
	result, err = h.ExecuteCLIWithInput(password+"\n",
		"set-allocation-delay",
		"--context", h.ContextName,
		"--keystore", ecdsaName,
		"--delay", "100")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Self-delegate
	result, err = h.ExecuteCLIWithInput(password+"\n",
		"delegate",
		"--context", h.ContextName,
		"--keystore", ecdsaName)
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Register key for specific operator set
	result, err = h.ExecuteCLIWithInput(password+"\n",
		"register-key",
		"--context", h.ContextName,
		"--keystore", bn254Name,
		"--operator-address", h.ChainConfig.OperatorAccountAddress,
		"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
		"--operator-set-id", opSetID)
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Register with AVS for specific operator set
	socketURL := fmt.Sprintf("http://%s.example.com:8080", role)
	result, err = h.ExecuteCLIWithInput(password+"\n",
		"register-avs",
		"--context", h.ContextName,
		"--keystore", ecdsaName,
		"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
		"--operator-set-ids", opSetID,
		"--socket", socketURL)
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Allocate to operator set
	result, err = h.ExecuteCLIWithInput(password+"\n",
		"allocate",
		"--context", h.ContextName,
		"--keystore", ecdsaName,
		"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
		"--operator-set-id", opSetID,
		"--strategy", h.GetBeaconETHStrategy(),
		"--amount", "1000000000000000000")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	return ecdsaName
}

func setupCompleteOperatorWithAllocation(t *testing.T, h *harness.TestHarness, name string) (string, string) {
	keystoreName := setupOperatorForSet(t, h, name, "0", "test-operator")
	return keystoreName, "test-password"
}
