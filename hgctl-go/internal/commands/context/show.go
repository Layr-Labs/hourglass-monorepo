package context

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

func showCommand() *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show current context details",
		Action: contextShowAction,
	}
}

func contextShowAction(_ *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, exists := cfg.Contexts[cfg.CurrentContext]
	if !exists {
		return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
	}

	// Use the ToMap method to get the context data
	data := map[string]interface{}{
		"current-context": cfg.CurrentContext,
		"context":         ctx.ToMap(),
	}

	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()

	return encoder.Encode(data)
}
