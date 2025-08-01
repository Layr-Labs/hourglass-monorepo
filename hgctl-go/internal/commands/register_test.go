package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestRegisterCommand(t *testing.T) {
	cmd := RegisterCommand()
	
	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "register", cmd.Name)
		assert.Equal(t, "Register operator with EigenLayer", cmd.Usage)
		assert.NotNil(t, cmd.Action)
	})
	
	t.Run("Required Flags", func(t *testing.T) {
		requiredFlags := map[string]bool{
			"address":          false,
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
		
		assert.True(t, requiredFlags["address"], "address flag should be required")
		assert.False(t, requiredFlags["allocation-delay"], "allocation-delay flag is optional with default")
	})
	
	t.Run("Environment Variable Support", func(t *testing.T) {
		// Check that the command supports env vars (based on our implementation)
		var privateKeyFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "private-key" {
				privateKeyFlag = sf
				break
			}
		}
		
		assert.NotNil(t, privateKeyFlag)
		assert.Contains(t, privateKeyFlag.Usage, "PRIVATE_KEY env var")
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