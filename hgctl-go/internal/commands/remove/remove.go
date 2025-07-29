package remove

import (
	"github.com/urfave/cli/v2"
)

// Command returns the remove command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove resources",
		Subcommands: []*cli.Command{
			performerCommand(),
		},
	}
}
