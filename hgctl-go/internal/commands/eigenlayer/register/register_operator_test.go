package register

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestRegisterOperatorCommand(t *testing.T) {
	cmd := RegisterOperatorCommand()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "register-operator", cmd.Name)
		assert.Equal(t, "Register operator with EigenLayer", cmd.Usage)
		assert.NotNil(t, cmd.Action)
	})

	t.Run("Required Flags", func(t *testing.T) {
		requiredFlags := map[string]bool{
			"allocation-delay": false,
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

		assert.False(t, requiredFlags["allocation-delay"], "allocation-delay flag is optional with default")

		// Verify metadata-uri exists
		var metadataFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "metadata-uri" {
				metadataFlag = sf
				break
			}
		}
		assert.NotNil(t, metadataFlag, "metadata-uri flag should exist")
	})

	t.Run("Flag Validation", func(t *testing.T) {
		// After middleware refactoring, private-key flag no longer exists
		// Context configuration provides the necessary values
		expectedFlags := []string{"metadata-uri", "allocation-delay"}
		foundFlags := make(map[string]bool)

		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok {
				foundFlags[sf.Name] = true
			} else if uf, ok := flag.(*cli.Uint64Flag); ok {
				foundFlags[uf.Name] = true
			}
		}

		// Check that expected flags exist
		for _, expected := range expectedFlags {
			assert.True(t, foundFlags[expected], "Expected flag not found: %s", expected)
		}

		// Check that no unexpected flags exist
		for flagName := range foundFlags {
			found := false
			for _, expected := range expectedFlags {
				if flagName == expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Unexpected flag found: %s", flagName)
		}
	})

	t.Run("Default Metadata URI", func(t *testing.T) {
		var metadataFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "metadata-uri" {
				metadataFlag = sf
				break
			}
		}

		assert.NotNil(t, metadataFlag)
		// metadata-uri is optional, no default required
		assert.False(t, metadataFlag.Required, "metadata-uri should be optional")
	})
}
