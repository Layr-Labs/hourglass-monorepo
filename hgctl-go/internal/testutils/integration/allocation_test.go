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

func TestAllocationManagement(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	// For test chains, we'll proceed without checking for EigenLayer contracts
	// as we're testing with mock/test addresses

	t.Run("Set Allocation to Operator Set", func(t *testing.T) {
		// Complete operator setup including AVS registration
		operatorName := fmt.Sprintf("alloc-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupOperatorWithAVS(t, h, operatorName)

		// Set allocation delay first
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"set-allocation-delay",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--delay", "100")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Allocate to operator set
		result, err = h.ExecuteCLIWithInput(operatorPassword+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "1000000000000000000") // 1e18

		require.NoError(t, err)
		harness.Assert(t, result).
			AssertAllocationSet().
			OutputContainsAll("allocation set", "1000000000000000000")

		// TODO: Verify allocation on-chain
		// allocation, err := h.Contracts.GetOperatorAllocation(...)
	})

	t.Run("Allocate Multiple Strategies", func(t *testing.T) {
		// Setup operator
		operatorName := fmt.Sprintf("multi-strat-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupOperatorWithAVS(t, h, operatorName)

		// Set allocation delay
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"set-allocation-delay",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--delay", "100")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Allocate BeaconETH
		result, err = h.ExecuteCLIWithInput(operatorPassword+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "500000000000000000") // 0.5e18
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Allocate another strategy (if available)
		// This is a placeholder - would need actual strategy address
		// result, err = h.ExecuteCLIWithInput(operatorPassword+"\n",
		//     "allocate",
		//     "--context", h.ContextName,
		//     "--keystore", operatorName,
		//     "--avs", h.ChainConfig.AVSTaskRegistrarAddress,
		//     "--operator-set-id", "0",
		//     "--strategy", "0xAnotherStrategy...",
		//     "--amount", "300000000000000000")
	})

	t.Run("Deallocate from Operator Set", func(t *testing.T) {
		// Setup operator with allocation
		operatorName := fmt.Sprintf("dealloc-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupOperatorWithAllocation(t, h, operatorName)

		// Deallocate
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"deallocate",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "500000000000000000") // Deallocate half

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("deallocation", "queued").
			TransactionSucceeded()
	})

	t.Run("Invalid Allocation Amount", func(t *testing.T) {
		// Setup operator
		operatorName := fmt.Sprintf("invalid-alloc-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupOperatorWithAVS(t, h, operatorName)

		// Try to allocate 0
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "0")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("invalid amount|must be greater than 0")
	})

	t.Run("Allocation Before Registration", func(t *testing.T) {
		// Create operator but don't register with AVS
		operatorName := fmt.Sprintf("unregistered-alloc-%d", time.Now().UnixNano())
		keystoreName, password := setupRegisteredOperator(t, h, operatorName)

		// Try to allocate without AVS registration
		result, err := h.ExecuteCLIWithInput(password+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", keystoreName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "1000000000000000000")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("not registered to AVS|operator not found")
	})
}

func TestAllocationStatus(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Check Operator Allocations", func(t *testing.T) {
		// Setup operator with allocations
		operatorName := fmt.Sprintf("status-operator-%d", time.Now().UnixNano())
		operatorName, _ = setupOperatorWithAllocation(t, h, operatorName)

		// Check allocation status
		result, err := h.ExecuteCLI("allocation", "status",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("Allocations", "Strategy", "Amount", "Operator Set")
	})

	t.Run("List All Allocations", func(t *testing.T) {
		// This assumes we have some operators with allocations from previous tests
		result, err := h.ExecuteCLI("allocation", "list",
			"--context", h.ContextName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputMatches("OPERATOR.*STRATEGY.*AMOUNT")

		// Parse table to verify format
		_, err = result.ParseTable()
		assert.NoError(t, err)
	})

	t.Run("Check Pending Deallocations", func(t *testing.T) {
		// Setup operator with pending deallocation
		operatorName := fmt.Sprintf("pending-dealloc-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupOperatorWithAllocation(t, h, operatorName)

		// Queue deallocation
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"deallocate",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "100000000000000000") // 0.1e18
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Check pending deallocations
		result, err = h.ExecuteCLI("allocation", "pending",
			"--context", h.ContextName,
			"--keystore", operatorName)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("Pending Deallocations", "100000000000000000")
	})
}

func TestAllocationConstraints(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Respect Allocation Delay", func(t *testing.T) {
		// Setup operator
		operatorName := fmt.Sprintf("delay-test-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupOperatorWithAVS(t, h, operatorName)

		// Set short allocation delay
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"set-allocation-delay",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--delay", "10") // 10 blocks
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// First allocation
		result, err = h.ExecuteCLIWithInput(operatorPassword+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "1000000000000000000")
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Try immediate reallocation (should fail due to delay)
		result, err = h.ExecuteCLIWithInput(operatorPassword+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", "2000000000000000000")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("allocation delay|too soon|wait")
	})

	t.Run("Over-allocation Prevention", func(t *testing.T) {
		// Setup operator
		operatorName := fmt.Sprintf("overalloc-operator-%d", time.Now().UnixNano())
		operatorName, operatorPassword := setupOperatorWithAVS(t, h, operatorName)

		// Try to allocate more than available balance
		hugeAmount := "999999999999999999999999999999999999"
		result, err := h.ExecuteCLIWithInput(operatorPassword+"\n",
			"allocate",
			"--context", h.ContextName,
			"--keystore", operatorName,
			"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
			"--operator-set-id", "0",
			"--strategy", h.GetBeaconETHStrategy(),
			"--amount", hugeAmount)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("insufficient.*balance|exceeds.*available")
	})
}

// Helper functions for allocation tests

func setupOperatorWithAVS(t *testing.T, h *harness.TestHarness, name string) (string, string) {
	// Complete operator setup through AVS registration
	keystoreName, password := setupCompleteOperator(t, h, name)

	// Register with AVS
	result, err := h.ExecuteCLIWithInput(password+"\n",
		"register-avs",
		"--context", h.ContextName,
		"--keystore", keystoreName,
		"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
		"--operator-set-ids", "0",
		"--socket", "http://alloc-test.example.com:8080")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Wait for transaction
	if txHash, err := result.GetTransactionHash(); err == nil && txHash != "" {
		h.WaitForTransaction(context.Background(), txHash)
	}

	return keystoreName, password
}

func setupOperatorWithAllocation(t *testing.T, h *harness.TestHarness, name string) (string, string) {
	// Setup operator with AVS
	keystoreName, password := setupOperatorWithAVS(t, h, name)

	// Set allocation delay
	result, err := h.ExecuteCLIWithInput(password+"\n",
		"set-allocation-delay",
		"--context", h.ContextName,
		"--keystore", keystoreName,
		"--delay", "100")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Make allocation
	result, err = h.ExecuteCLIWithInput(password+"\n",
		"allocate",
		"--context", h.ContextName,
		"--keystore", keystoreName,
		"--avs", h.ChainConfig.AVSTaskRegistrarAddress,
		"--operator-set-id", "0",
		"--strategy", h.GetBeaconETHStrategy(),
		"--amount", "1000000000000000000")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Wait for transaction
	if txHash, err := result.GetTransactionHash(); err == nil && txHash != "" {
		h.WaitForTransaction(context.Background(), txHash)
	}

	return keystoreName, password
}
