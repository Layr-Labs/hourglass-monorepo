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
			"strategy":  false,
			"magnitude": false,
		}

		for _, flag := range cmd.Flags {
			switch f := flag.(type) {
			case *cli.StringFlag:
				if _, exists := requiredFlags[f.Name]; exists {
					requiredFlags[f.Name] = f.Required
				}
			}
		}

		assert.True(t, requiredFlags["strategy"], "strategy flag should be required")
		assert.True(t, requiredFlags["magnitude"], "magnitude flag should be required")
		
		// Verify operator-set-id flag no longer exists (comes from context now)
		for _, flag := range cmd.Flags {
			switch f := flag.(type) {
			case *cli.Uint64Flag:
				assert.NotEqual(t, "operator-set-id", f.Name, "operator-set-id flag should not exist (comes from context)")
			case *cli.UintFlag:
				assert.NotEqual(t, "operator-set-id", f.Name, "operator-set-id flag should not exist (comes from context)")
			}
		}
	})
}
