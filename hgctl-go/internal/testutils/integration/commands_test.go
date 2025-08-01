package integration

import (
	"fmt"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/require"
)

// TestCommandIntegration tests the integration of various commands
// These tests verify command behavior without requiring deployed contracts
func TestCommandIntegration(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Deposit Command Validation", func(t *testing.T) {
		// Test missing required flags
		result, err := h.ExecuteCLI("deposit")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("Required flag.*strategy|strategy.*required")

		// Test with strategy but missing amount
		result, err = h.ExecuteCLI("deposit", "--strategy", "0x1234567890123456789012345678901234567890")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("Required flag.*amount|amount.*required")

		// Test invalid amount format
		result, err = h.ExecuteCLI("deposit",
			"--strategy", "0x1234567890123456789012345678901234567890",
			"--amount", "invalid")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("invalid.*amount|amount.*invalid")

		// Test valid ether format
		result, err = h.ExecuteCLI("deposit",
			"--strategy", "0x1234567890123456789012345678901234567890",
			"--amount", "1.5 ether")
		require.NoError(t, err)
		// Will fail due to no contract, but validates amount parsing worked
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to deposit")
	})

	t.Run("Allocate Command Validation", func(t *testing.T) {
		// Test missing required flags
		result, err := h.ExecuteCLI("allocate")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("Required flag.*avs|avs.*required")

		// Test with partial flags
		result, err = h.ExecuteCLI("allocate",
			"--avs", "0x1234567890123456789012345678901234567890")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("Required flag.*operator-set-id|operator-set-id.*required")

		// Test with all required flags but invalid magnitude
		result, err = h.ExecuteCLI("allocate",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--strategy", "0xabcdef0123456789012345678901234567890123",
			"--magnitude", "invalid")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("invalid magnitude|magnitude.*invalid")

		// Test with valid scientific notation magnitude
		result, err = h.ExecuteCLI("allocate",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--strategy", "0xabcdef0123456789012345678901234567890123",
			"--magnitude", "1e18")
		require.NoError(t, err)
		// Will fail due to no contract, but validates parsing worked
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to modify allocations")
	})

	t.Run("Register Key Command Validation", func(t *testing.T) {
		// Test invalid key type
		result, err := h.ExecuteCLI("register-key",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--key-type", "invalid")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("invalid key type.*must be.*ecdsa.*bn254")

		// Test ECDSA without key-address
		result, err = h.ExecuteCLI("register-key",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--key-type", "ecdsa")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("key-address.*required.*ECDSA")

		// Test BN254 without key-data
		result, err = h.ExecuteCLI("register-key",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--key-type", "bn254")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("key-data.*required.*BN254")

		// Test with valid ECDSA parameters
		result, err = h.ExecuteCLI("register-key",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--key-type", "ecdsa",
			"--key-address", "0xabcdef0123456789012345678901234567890123")
		require.NoError(t, err)
		// Will fail due to no contract, but validates parsing worked
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to register key")
	})

	t.Run("Register AVS Command Validation", func(t *testing.T) {
		// Test missing required flags
		result, err := h.ExecuteCLI("register-avs")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("Required flag.*address|address.*required")

		// Test with multiple operator set IDs
		result, err = h.ExecuteCLI("register-avs",
			"--address", "0x1234567890123456789012345678901234567890",
			"--avs", "0xabcdef0123456789012345678901234567890123",
			"--operator-set-ids", "0,1,2",
			"--socket", "http://operator.example.com:8080")
		require.NoError(t, err)
		// Will fail due to no contract, but validates parsing worked
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to register operator to AVS")
	})

	t.Run("Delegate Command Validation", func(t *testing.T) {
		// Test delegate without any parameters (should use env vars)
		result, err := h.ExecuteCLI("delegate")
		require.NoError(t, err)
		// Will fail due to no contract, but shows env vars were used
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to delegate")

		// Test with explicit operator address
		result, err = h.ExecuteCLI("delegate",
			"--operator", "0x1234567890123456789012345678901234567890")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to delegate")
	})

	t.Run("Set Allocation Delay Validation", func(t *testing.T) {
		// Test missing required delay
		result, err := h.ExecuteCLI("set-allocation-delay")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("Required flag.*delay|delay.*required")

		// Test with delay value
		result, err = h.ExecuteCLI("set-allocation-delay", "--delay", "100")
		require.NoError(t, err)
		// Will fail due to no contract, but validates parsing worked
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to set allocation delay")
	})
}

func TestCommandHelp(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	commands := []string{
		"deposit",
		"allocate",
		"register",
		"register-key",
		"register-avs",
		"delegate",
		"set-allocation-delay",
	}

	for _, cmd := range commands {
		t.Run(fmt.Sprintf("%s Help", cmd), func(t *testing.T) {
			result, err := h.ExecuteCLI(cmd, "--help")
			require.NoError(t, err)
			harness.Assert(t, result).
				HasExitCode(0).
				OutputContainsAll("NAME", "USAGE", "OPTIONS")
		})
	}
}

func TestEnvironmentVariableSupport(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Commands Use Environment Variables", func(t *testing.T) {
		// Test that commands pick up ETH_RPC_URL and PRIVATE_KEY from environment
		// These are set by the test harness
		
		// Register command should work with env vars
		result, err := h.ExecuteCLI("register",
			"--address", h.ChainConfig.OperatorAccountAddress,
			"--allocation-delay", "1")
		require.NoError(t, err)
		// Will fail due to no contract, but should not complain about missing RPC or key
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to register")
		
		// Deposit command should work with env vars
		result, err = h.ExecuteCLI("deposit",
			"--strategy", "0x1234567890123456789012345678901234567890",
			"--amount", "1 ether")
		require.NoError(t, err)
		// Will fail due to no contract, but should not complain about missing RPC or key
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to deposit")
	})
}