package eigenlayer

import (
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/eigenlayer/allocate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/eigenlayer/delegate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/eigenlayer/deposit"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/eigenlayer/register"
	"github.com/urfave/cli/v2"
)

// Command returns the eigenlayer command with all subcommands
func Command() *cli.Command {
	return &cli.Command{
		Name:    "eigenlayer",
		Aliases: []string{"el"},
		Usage:   "EigenLayer-specific operations",
		Description: `Manage EigenLayer operations including operator registration, delegation, 
deposits, allocations, and AVS registration.`,
		Subcommands: []*cli.Command{
			allocate.Command(),
			allocate.SetAllocationDelayCommand(),
			delegate.Command(),
			deposit.Command(),
			register.RegisterAVSCommand(),
			register.DeregisterAVSCommand(),
			register.RegisterOperatorCommand(),
			register.RegisterKeyCommand(),
		},
	}
}
