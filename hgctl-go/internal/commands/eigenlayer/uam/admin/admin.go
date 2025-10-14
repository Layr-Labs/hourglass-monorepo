package admin

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:    "admin",
		Usage:   "Admin management operations",
		Description: `Manage Admin Access Management operations.`,
		Subcommands: []*cli.Command{
			AcceptAdminCommand(),
			AddPendingAdminCommand(),
			IsAdminCommand(),
			IsPendingCommand(),
			ListAdminsCommand(),
			ListPendingCommand(),
			RemoveAdminCommand(),
			RemovePendingCommand(),
		},
	}
}
