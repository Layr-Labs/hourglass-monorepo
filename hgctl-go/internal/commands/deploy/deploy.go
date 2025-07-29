package deploy

import (
	"github.com/urfave/cli/v2"
)

// Command returns the deploy command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "deploy",
		Usage: "Deploy resources",
		Subcommands: []*cli.Command{
			artifactCommand(),
			executorCommand(),
			aggregatorCommand(),
		},
	}
}