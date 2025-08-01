package operatorset

import (
	"github.com/urfave/cli/v2"
)

// Command returns the operator-set command with subcommands
func Command() *cli.Command {
	return &cli.Command{
		Name:  "operator-set",
		Usage: "Manage operator sets for an AVS",
		Description: `Commands for discovering and describing operator sets.
For Hourglass AVS, operator sets include the aggregator operator set and executor operator sets.

The AVS address must be configured in the context before running these commands.`,
		Subcommands: []*cli.Command{
			getCommand(),
			describeCommand(),
		},
	}
}