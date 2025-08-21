package register

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterKeyCommand(t *testing.T) {
	cmd := RegisterKeyCommand()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "register-key", cmd.Name)
		assert.Equal(t, "Register system signing key with AVS", cmd.Usage)
		assert.NotNil(t, cmd.Action)
		assert.Contains(t, cmd.Description, "ECDSA")
		assert.Contains(t, cmd.Description, "BN254")
	})

	t.Run("No Flags Required", func(t *testing.T) {
		// The refactored command uses context instead of flags
		assert.Empty(t, cmd.Flags, "Command should have no flags as everything comes from context")
	})

	t.Run("Context Prerequisites in Description", func(t *testing.T) {
		// Verify the description mentions context prerequisites
		assert.Contains(t, cmd.Description, "Prerequisites:")
		assert.Contains(t, cmd.Description, "AVS address must be configured")
		assert.Contains(t, cmd.Description, "Operator set ID must be configured")
		assert.Contains(t, cmd.Description, "Operator signer must be configured")
		assert.Contains(t, cmd.Description, "System signer must be configured")
		assert.Contains(t, cmd.Description, "hgctl context set")
		assert.Contains(t, cmd.Description, "hgctl signer")
	})

	t.Run("Environment Variables in Description", func(t *testing.T) {
		// Verify the description mentions required environment variables
		assert.Contains(t, cmd.Description, "SYSTEM_KEYSTORE_PASSWORD")
		assert.Contains(t, cmd.Description, "environment variable must be set")
	})

	t.Run("Usage Examples in Description", func(t *testing.T) {
		// Verify usage example is present
		assert.Contains(t, cmd.Description, "Usage:")
		assert.Contains(t, cmd.Description, "hgctl register-key")
	})
}
