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

	t.Run("Context Prerequisites Documentation", func(t *testing.T) {
		// Verify the command description includes context prerequisites
		assert.Contains(t, cmd.Description, "Prerequisites:", "Description should have prerequisites section")
		assert.Contains(t, cmd.Description, "AVS address must be configured", "Description should mention AVS address requirement")
		assert.Contains(t, cmd.Description, "Operator set ID must be configured", "Description should mention operator set ID requirement")
		assert.Contains(t, cmd.Description, "Operator address must be configured", "Description should mention operator address requirement")
		assert.Contains(t, cmd.Description, "hgctl context set", "Description should include context set examples")
	})

	t.Run("Magnitude Flag Details", func(t *testing.T) {
		// Find the magnitude flag
		var magnitudeFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "magnitude" {
				magnitudeFlag = sf
				break
			}
		}

		assert.NotNil(t, magnitudeFlag, "magnitude flag should exist")
		assert.True(t, magnitudeFlag.Required, "magnitude flag should be required")
		assert.Contains(t, magnitudeFlag.Usage, "1e18", "magnitude usage should include example format")
		assert.Contains(t, magnitudeFlag.Usage, "Allocation magnitude", "magnitude usage should describe purpose")
	})

	t.Run("Strategy Flag Details", func(t *testing.T) {
		// Find the strategy flag
		var strategyFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "strategy" {
				strategyFlag = sf
				break
			}
		}

		assert.NotNil(t, strategyFlag, "strategy flag should exist")
		assert.True(t, strategyFlag.Required, "strategy flag should be required")
		assert.Contains(t, strategyFlag.Usage, "Strategy contract address", "strategy usage should describe purpose")
	})
}
