package get

import (
	"github.com/urfave/cli/v2"
)

// Command returns the get command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Retrieve resources",
		Subcommands: []*cli.Command{
			performerCommand(),
			releaseCommand(),
			operatorSetCommand(),
		},
	}
}
