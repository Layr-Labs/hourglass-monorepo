package rewards

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "rewards",
		Usage: "Manage operator rewards",
		Subcommands: []*cli.Command{
			ShowCommand(),
			SetClaimerCommand(),
			ClaimCommand(),
		},
	}
}
