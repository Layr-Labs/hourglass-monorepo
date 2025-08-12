package context

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

func useCommand() *cli.Command {
	return &cli.Command{
		Name:      "use",
		Usage:     "Switch to a different context",
		ArgsUsage: "<context-name>",
		Action:    contextUseAction,
	}
}

func contextUseAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	contextName := c.Args().Get(0)
	log := config.LoggerFromContext(c.Context)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if _, exists := cfg.Contexts[contextName]; !exists {
		return fmt.Errorf("context '%s' not found", contextName)
	}

	cfg.CurrentContext = contextName

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Info("Switched context", zap.String("context", contextName))
	fmt.Printf("Switched to context '%s'\n", contextName)

	return nil
}
