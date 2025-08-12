package delegate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestDelegateCommand(t *testing.T) {
	cmd := Command()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "delegate", cmd.Name)
		assert.Equal(t, "Self-delegate as an operator", cmd.Usage)
		assert.NotNil(t, cmd.Action)
		assert.Contains(t, cmd.Description, "Self-delegate")
	})

	t.Run("Optional Flags", func(t *testing.T) {
		// All flags should be optional for delegate
		for _, flag := range cmd.Flags {
			switch f := flag.(type) {
			case *cli.StringFlag:
				assert.False(t, f.Required, "%s flag should be optional", f.Name)
			}
		}
	})

	t.Run("Operator Flag", func(t *testing.T) {
		var operatorFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "operator" {
				operatorFlag = sf
				break
			}
		}

		assert.NotNil(t, operatorFlag)
		assert.Contains(t, operatorFlag.Usage, "defaults to configured operator address")
		assert.Contains(t, operatorFlag.Usage, "self-delegation")
	})
}
