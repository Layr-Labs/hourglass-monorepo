package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeystoreOperations(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Create ECDSA Keystore", func(t *testing.T) {
		// Test creating ECDSA keystore with password input
		result, err := h.ExecuteCLIWithInput("test-password\ntest-password\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", "test-ecdsa",
			"--type", "ecdsa")

		require.NoError(t, err)
		if result.ExitCode != 0 {
			t.Logf("Keystore creation failed:\nStdout: %s\nStderr: %s", result.Stdout, result.Stderr)
		}
		harness.Assert(t, result).
			HasExitCode(0)
		// The keystore creation logs go to stderr
		require.Contains(t, result.Stderr, "ECDSA keystore created successfully")
	})

	t.Run("Create BN254 Keystore", func(t *testing.T) {
		// Test creating BN254 keystore
		result, err := h.ExecuteCLIWithInput("bn254-password\nbn254-password\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", "test-bn254",
			"--type", "bn254")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0)
		// The keystore creation logs go to stderr
		require.Contains(t, result.Stderr, "BLS keystore created successfully")
	})

	t.Run("List Keystores", func(t *testing.T) {
		result, err := h.ExecuteCLI("keystore", "list",
			"--context", h.ContextName)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("test-ecdsa", "test-bn254", "ecdsa", "bn254")
	})

	t.Run("Create Duplicate Keystore", func(t *testing.T) {
		// Attempt to create keystore with existing name
		result, err := h.ExecuteCLIWithInput("test-password\ntest-password\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", "test-ecdsa",
			"--type", "ecdsa")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("already exists|duplicate")
	})

	t.Run("Password Mismatch", func(t *testing.T) {
		// Test password confirmation mismatch
		result, err := h.ExecuteCLIWithInput("password1\npassword2\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", "test-mismatch",
			"--type", "ecdsa")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("password.*mismatch|do not match")
	})

	t.Run("Invalid Keystore Type", func(t *testing.T) {
		result, err := h.ExecuteCLI("keystore", "create",
			"--context", h.ContextName,
			"--name", "test-invalid",
			"--type", "invalid-type")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("invalid.*type|unsupported")
	})

	t.Run("Import Existing Key", func(t *testing.T) {
		// Test importing a specific private key
		privateKey := "1234567890123456789012345678901234567890123456789012345678901234"

		result, err := h.ExecuteCLIWithInput("import-password\nimport-password\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", "test-import",
			"--type", "ecdsa",
			"--key", privateKey)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0)
		// Check that the keystore was created
		require.Contains(t, result.Stderr, "ECDSA keystore created successfully")
	})


	t.Run("Delete Keystore", func(t *testing.T) {
		t.Skip("Delete command not implemented in hgctl")
		// Create a keystore to delete
		keystoreName := fmt.Sprintf("test-delete-%d", time.Now().UnixNano())
		result, err := h.ExecuteCLIWithInput("delete-password\ndelete-password\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--type", "ecdsa")

		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Delete it
		result, err = h.ExecuteCLI("keystore", "delete",
			"--context", h.ContextName,
			"--name", keystoreName,
			"--force") // Skip confirmation

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("deleted", keystoreName)

		// Verify it's gone
		result, err = h.ExecuteCLI("keystore", "list",
			"--context", h.ContextName)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsNone(keystoreName)
	})

	t.Run("Context Isolation", func(t *testing.T) {
		t.Skip("Context create command not implemented in hgctl")
		// Create a new context
		newContext := fmt.Sprintf("test-isolation-%d", time.Now().UnixNano())
		result, err := h.ExecuteCLI("context", "create",
			"--name", newContext,
			"--rpc-url", h.ChainConfig.L1RPC,
			"--chain-id", fmt.Sprintf("%d", h.ChainConfig.L1ChainID))

		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Create keystore in new context
		result, err = h.ExecuteCLIWithInput("isolated-password\nisolated-password\n",
			"keystore", "create",
			"--context", newContext,
			"--name", "isolated-keystore",
			"--type", "ecdsa")

		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		// Verify keystore is not visible in original context
		result, err = h.ExecuteCLI("keystore", "list",
			"--context", h.ContextName)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsNone("isolated-keystore")

		// Verify keystore is visible in new context
		result, err = h.ExecuteCLI("keystore", "list",
			"--context", newContext)

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0).
			OutputContainsAll("isolated-keystore")
	})
}

func TestKeystoreTableFormatting(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	// Create multiple keystores
	keystores := []struct {
		name     string
		keyType  string
		password string
	}{
		{"format-test-1", "ecdsa", "password1"},
		{"format-test-2", "bn254", "password2"},
		{"format-test-long-name-keystore", "ecdsa", "password3"},
	}

	for _, ks := range keystores {
		result, err := h.ExecuteCLIWithInput(ks.password+"\n"+ks.password+"\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", ks.name,
			"--type", ks.keyType)

		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)
	}

	// Test list command works - just verify it returns the expected keystores
	result, err := h.ExecuteCLI("keystore", "list",
		"--context", h.ContextName)

	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	// Verify all test keystores appear in the output
	for _, ks := range keystores {
		assert.Contains(t, result.Stdout, ks.name)
		assert.Contains(t, result.Stdout, ks.keyType) // Type is shown in lowercase
	}
}

func TestKeystorePasswordHandling(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)

	t.Run("Empty Password Rejected", func(t *testing.T) {
		result, err := h.ExecuteCLIWithInput("\n\n",
			"keystore", "create",
			"--context", h.ContextName,
			"--name", "test-empty-password",
			"--type", "ecdsa")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(1).
			ErrorMatches("password cannot be empty|empty password")
	})

	t.Run("Long Password Accepted", func(t *testing.T) {
		// Test with a very long password
		longPassword := strings.Repeat("a", 256)
		input := longPassword + "\n" + longPassword + "\n"

		result, err := h.ExecuteCLIWithInput(input,
			"keystore", "create",
			"--context", h.ContextName,
			"--name", "test-long-password",
			"--type", "ecdsa")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0)
		// Check stderr for success message
		require.Contains(t, result.Stderr, "Keystore created and registered")
	})

	t.Run("Special Characters in Password", func(t *testing.T) {
		// Test with special characters
		specialPassword := "p@ssw0rd!#$%^&*()"
		input := specialPassword + "\n" + specialPassword + "\n"

		result, err := h.ExecuteCLIWithInput(input,
			"keystore", "create",
			"--context", h.ContextName,
			"--name", "test-special-password",
			"--type", "ecdsa")

		require.NoError(t, err)
		harness.Assert(t, result).
			HasExitCode(0)
		// Check stderr for success message
		require.Contains(t, result.Stderr, "Keystore created and registered")

		// TODO: Verify we can use the keystore with the special password once export command is implemented
		// result, err = h.ExecuteCLIWithInput(specialPassword+"\n",
		// 	"keystore", "export",
		// 	"--context", h.ContextName,
		// 	"--name", "test-special-password")
		//
		// require.NoError(t, err)
		// harness.Assert(t, result).
		// 	HasExitCode(0).
		// 	OutputMatches("0x[a-fA-F0-9]{64}")
	})
}
