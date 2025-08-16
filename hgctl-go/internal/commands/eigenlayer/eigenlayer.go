package eigenlayer

import (
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/allocate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/delegate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/deposit"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/register"
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
			delegate.Command(),
			deposit.Command(),
			register.RegisterAVSCommand(),
			register.RegisterOperatorCommand(),
			register.RegisterKeyCommand(),
			register.SetAllocationDelayCommand(),
		},
	}
}