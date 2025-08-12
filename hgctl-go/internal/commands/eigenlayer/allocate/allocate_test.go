package allocate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestAllocateCommand(t *testing.T) {
	cmd := Command()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "allocate", cmd.Name)
		assert.Equal(t, "Modify operator allocations to AVS operator sets", cmd.Usage)
		assert.NotNil(t, cmd.Action)
		assert.Contains(t, cmd.Description, "operator sets")
	})

	t.Run("Required Flags", func(t *testing.T) {
		requiredFlags := map[string]bool{
			"operator-set-id": false,
			"strategy":        false,
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
		assert.True(t, requiredFlags["strategy"], "strategy flag should be required")
	})

	t.Run("Default Values", func(t *testing.T) {
		var magnitudeFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "magnitude" {
				magnitudeFlag = sf
				break
			}
		}

		assert.NotNil(t, magnitudeFlag)
		assert.Equal(t, "1e18", magnitudeFlag.Value, "magnitude should default to 1e18")
	})

	t.Run("Example in Description", func(t *testing.T) {
		assert.Contains(t, cmd.Description, "Example:")
		assert.Contains(t, cmd.Description, "hgctl allocate")
	})
}
