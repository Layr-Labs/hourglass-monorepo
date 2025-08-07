package context

import (
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List all contexts",
		Action: contextListAction,
	}
}

func contextListAction(c *cli.Context) error {
	log := logger.FromContext(c.Context)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log.Info("Listing contexts", zap.Int("count", len(cfg.Contexts)))

	// Create table
	table := tablewriter.NewWriter(c.App.Writer)
	table.SetHeader([]string{"CURRENT", "NAME", "EXECUTOR ADDRESS", "AVS ADDRESS", "RPC URL"})

	for name, ctx := range cfg.Contexts {
		current := ""
		if name == cfg.CurrentContext {
			current = "*"
		}

		avsAddr := ctx.AVSAddress
		if avsAddr == "" {
			avsAddr = "-"
		}

		rpcURL := ctx.L1RPCUrl
		if rpcURL == "" {
			rpcURL = "-"
		}

		table.Append([]string{
			current,
			name,
			ctx.ExecutorAddress,
			avsAddr,
			rpcURL,
		})
	}

	table.Render()
	return nil
}
