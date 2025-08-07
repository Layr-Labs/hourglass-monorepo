package describe

import (
	"github.com/urfave/cli/v2"
)

// Command returns the describe command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "describe",
		Usage: "Describe resources in detail",
		Subcommands: []*cli.Command{
			releaseCommand(),
			operatorSetCommand(),
		},
	}
}
