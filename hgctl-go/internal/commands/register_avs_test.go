package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestRegisterAVSCommand(t *testing.T) {
	cmd := RegisterAVSCommand()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "register-avs", cmd.Name)
		assert.Equal(t, "Register operator with an AVS", cmd.Usage)
		assert.NotNil(t, cmd.Action)
		assert.Contains(t, cmd.Description, "Actively Validated Service")
	})

	t.Run("Required Flags", func(t *testing.T) {
		requiredFlags := map[string]bool{
			"operator-set-ids": false,
			"socket":           false,
		}

		for _, flag := range cmd.Flags {
			switch f := flag.(type) {
			case *cli.StringFlag:
				if _, exists := requiredFlags[f.Name]; exists {
					requiredFlags[f.Name] = f.Required
				}
			case *cli.Uint64SliceFlag:
				if _, exists := requiredFlags[f.Name]; exists {
					requiredFlags[f.Name] = f.Required
				}
			}
		}

		assert.True(t, requiredFlags["operator-set-ids"], "operator-set-ids flag should be required")
		assert.True(t, requiredFlags["socket"], "socket flag should be required")
	})

	t.Run("Socket Format Example", func(t *testing.T) {
		// Check that the command description includes socket format example
		// The actual command shows https:// in the example, so test expects https://
		assert.Contains(t, cmd.Description, "https://operator.example.com:8080")

		// Check socket flag usage
		var socketFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "socket" {
				socketFlag = sf
				break
			}
		}

		assert.NotNil(t, socketFlag)
		assert.Contains(t, socketFlag.Usage, "endpoint")
	})

	t.Run("Multiple Operator Set IDs", func(t *testing.T) {
		// Verify operator-set-ids is a slice flag
		var operatorSetIDsFlag *cli.Uint64SliceFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.Uint64SliceFlag); ok && sf.Name == "operator-set-ids" {
				operatorSetIDsFlag = sf
				break
			}
		}

		assert.NotNil(t, operatorSetIDsFlag)
		assert.Contains(t, operatorSetIDsFlag.Usage, "multiple")
	})
}
