package uam

import (
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/eigenlayer/uam/admin"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/eigenlayer/uam/appointee"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "user",
		Usage:       "User Access Management operations",
		Description: `Manage User Access Management operations for admins and appointees.`,
		Subcommands: []*cli.Command{
			admin.Command(),
			appointee.Command(),
		},
	}
}
