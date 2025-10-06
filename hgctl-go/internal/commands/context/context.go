package context

import (
	"github.com/urfave/cli/v2"
)

// Command returns the context command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "context",
		Usage: "Manage contexts",
		Subcommands: []*cli.Command{
			createCommand(),
			copyCommand(),
			deleteCommand(),
			listCommand(),
			useCommand(),
			setCommand(),
			removeCommand(),
			showCommand(),
		},
	}
}
