package appointee

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "appointee",
		Usage:       "Appointee management operations",
		Description: `Manage Appointee Access Management operations.`,
		Subcommands: []*cli.Command{
			CanCallCommand(),
			ListAppointeesCommand(),
			ListPermissionsCommand(),
			RemoveAppointeeCommand(),
		},
	}
}
