package keystore

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "keystore",
		Usage: "Manage keystores for operators",
		Subcommands: []*cli.Command{
			createCommand(),
			importCommand(),
			listCommand(),
			showCommand(),
		},
	}
}
