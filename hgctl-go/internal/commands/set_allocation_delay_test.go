package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestSetAllocationDelayCommand(t *testing.T) {
	cmd := SetAllocationDelayCommand()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "set-allocation-delay", cmd.Name)
		assert.Equal(t, "Set allocation delay for an operator", cmd.Usage)
		assert.NotNil(t, cmd.Action)
		assert.Contains(t, cmd.Description, "allocation delay")
	})

	t.Run("Required Flags", func(t *testing.T) {
		requiredFlags := map[string]bool{
			"delay": false,
		}

		for _, flag := range cmd.Flags {
			if uf, ok := flag.(*cli.Uint64Flag); ok {
				if _, exists := requiredFlags[uf.Name]; exists {
					requiredFlags[uf.Name] = uf.Required
				}
			}
		}

		assert.True(t, requiredFlags["delay"], "delay flag should be required")
	})

	t.Run("Delay Flag Details", func(t *testing.T) {
		var delayFlag *cli.Uint64Flag
		for _, flag := range cmd.Flags {
			if uf, ok := flag.(*cli.Uint64Flag); ok && uf.Name == "delay" {
				delayFlag = uf
				break
			}
		}

		assert.NotNil(t, delayFlag)
		assert.Contains(t, delayFlag.Usage, "seconds")
	})
}
