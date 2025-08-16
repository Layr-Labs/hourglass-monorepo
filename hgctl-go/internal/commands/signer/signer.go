package signer

import (
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "signer",
		Usage: "Select a signer configuration",
		Subcommands: []*cli.Command{
			importCommand(),
			listCommand(),
		},
	}
}

func getContextName(c *cli.Context) string {
	contextName := c.String("context")
	if contextName == "" {
		// Get current context if not specified
		cfg, err := config.LoadConfig()
		if err == nil && cfg.CurrentContext != "" {
			contextName = cfg.CurrentContext
		} else {
			contextName = "default"
		}
	}
	return contextName
}

func addContextFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:    "context",
		Aliases: []string{"c"},
		Usage:   "Context to use (defaults to current context)",
	})
}
