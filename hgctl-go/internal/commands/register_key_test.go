package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestRegisterKeyCommand(t *testing.T) {
	cmd := RegisterKeyCommand()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "register-key", cmd.Name)
		assert.Equal(t, "Register operator signing key with AVS", cmd.Usage)
		assert.NotNil(t, cmd.Action)
		assert.Contains(t, cmd.Description, "ECDSA")
		assert.Contains(t, cmd.Description, "BN254")
	})

	t.Run("Required Flags", func(t *testing.T) {
		requiredFlags := map[string]bool{
			"operator-set-id": false,
			"key-type":        false,
		}

		for _, flag := range cmd.Flags {
			switch f := flag.(type) {
			case *cli.StringFlag:
				if _, exists := requiredFlags[f.Name]; exists {
					requiredFlags[f.Name] = f.Required
				}
			case *cli.Uint64Flag:
				if _, exists := requiredFlags[f.Name]; exists {
					requiredFlags[f.Name] = f.Required
				}
			}
		}

		assert.True(t, requiredFlags["operator-set-id"], "operator-set-id flag should be required")
		assert.True(t, requiredFlags["key-type"], "key-type flag should be required")
	})

	t.Run("Conditional Flags", func(t *testing.T) {
		var keyAddressFlag, keyDataFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok {
				switch sf.Name {
				case "key-address":
					keyAddressFlag = sf
				case "key-data":
					keyDataFlag = sf
				}
			}
		}

		assert.NotNil(t, keyAddressFlag)
		assert.NotNil(t, keyDataFlag)
		assert.Contains(t, keyAddressFlag.Usage, "ecdsa")
		assert.Contains(t, keyDataFlag.Usage, "bn254")
		assert.False(t, keyAddressFlag.Required, "key-address should be optional")
		assert.False(t, keyDataFlag.Required, "key-data should be optional")
	})

	t.Run("Examples in Description", func(t *testing.T) {
		assert.Contains(t, cmd.Description, "For ECDSA keys:")
		assert.Contains(t, cmd.Description, "For BN254 keys:")
		assert.Contains(t, cmd.Description, "hgctl register-key")
	})
}
