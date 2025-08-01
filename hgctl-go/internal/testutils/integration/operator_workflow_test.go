package integration

import (
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/require"
)

// TestOperatorWorkflow tests a complete operator workflow
// This is an integration test that verifies commands work together
func TestOperatorWorkflow(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Complete Operator Setup Flow", func(t *testing.T) {
		// Step 1: Create operator keystores
		password := "test-password-123"
		
		// Create ECDSA keystore for operator
		result, err := h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--name", "operator-ecdsa-workflow",
			"--type", "ecdsa")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0)
		require.Contains(t, result.Stderr, "ECDSA keystore created successfully")
		
		// Create BN254 keystore for signing
		result, err = h.ExecuteCLIWithInput(password+"\n"+password+"\n",
			"keystore", "create",
			"--name", "operator-bn254-workflow",
			"--type", "bn254")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0)
		require.Contains(t, result.Stderr, "BLS keystore created successfully")
		
		// Step 2: Register as operator (will fail due to no contracts, but validates command)
		result, err = h.ExecuteCLI("register",
			"--address", h.ChainConfig.OperatorAccountAddress,
			"--allocation-delay", "100",
			"--metadata-uri", "https://example.com/operator/metadata.json")
		require.NoError(t, err)
		// Expect failure due to no contracts
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to register")
		
		// Step 3: Attempt to self-delegate (will fail due to no contracts)
		result, err = h.ExecuteCLI("delegate")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to delegate")
		
		// Step 4: Attempt to register with AVS (will fail due to no contracts)
		result, err = h.ExecuteCLI("register-avs",
			"--address", h.ChainConfig.OperatorAccountAddress,
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-ids", "0,1",
			"--socket", "http://operator.example.com:8080")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to register")
	})
}

// TestStakerWorkflow tests staker-specific operations
func TestStakerWorkflow(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Staker Deposit Flow", func(t *testing.T) {
		// Test deposit with various amount formats
		testCases := []struct {
			name   string
			amount string
		}{
			{"Wei amount", "1000000000000000000"},
			{"Ether amount", "1 ether"},
			{"Decimal ether", "0.5 ether"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := h.ExecuteCLI("deposit",
					"--strategy", "0x1234567890123456789012345678901234567890",
					"--amount", tc.amount)
				require.NoError(t, err)
				// Will fail due to no contracts, but validates amount parsing
				harness.Assert(t, result).
					HasExitCode(1).
					ErrorMatches("no contract code|failed to deposit")
			})
		}
	})
}

// TestOperatorAllocationWorkflow tests allocation management
func TestOperatorAllocationWorkflow(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Allocation Management Flow", func(t *testing.T) {
		// Step 1: Set allocation delay
		result, err := h.ExecuteCLI("set-allocation-delay",
			"--delay", "200")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to set")
			
		// Step 2: Allocate to operator set
		result, err = h.ExecuteCLI("allocate",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--strategy", "0xabcdef0123456789012345678901234567890123",
			"--magnitude", "5e17") // 50% allocation
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to modify")
			
		// Step 3: Allocate to another operator set
		result, err = h.ExecuteCLI("allocate",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "1",
			"--strategy", "0xabcdef0123456789012345678901234567890123",
			"--magnitude", "3e17") // 30% allocation
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to modify")
	})
}

// TestKeyRegistrationWorkflow tests different key registration scenarios
func TestKeyRegistrationWorkflow(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("ECDSA Key Registration", func(t *testing.T) {
		result, err := h.ExecuteCLI("register-key",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--key-type", "ecdsa",
			"--key-address", "0xabcdef0123456789012345678901234567890123")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to register key")
	})
	
	t.Run("BN254 Key Registration", func(t *testing.T) {
		// Example BN254 public key (hex encoded - 64 bytes)
		bn254Key := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		
		result, err := h.ExecuteCLI("register-key",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--key-type", "bn254",
			"--key-data", bn254Key)
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("no contract code|failed to register key")
	})
	
	t.Run("Invalid Key Type", func(t *testing.T) {
		result, err := h.ExecuteCLI("register-key",
			"--avs", "0x1234567890123456789012345678901234567890",
			"--operator-set-id", "0",
			"--key-type", "invalid")
		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("invalid key type")
	})
}