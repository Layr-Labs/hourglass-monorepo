package translate

import (
	"github.com/urfave/cli/v2"
)

// Command returns the translate command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "translate",
		Usage: "Translate runtime specs to different formats",
		Subcommands: []*cli.Command{
			composeCommand(),
			containerCommand(),
		},
	}
}