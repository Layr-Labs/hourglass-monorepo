package context

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func showCommand() *cli.Command {
	return &cli.Command{
		Name:   "show",
		Usage:  "Show current context details",
		Action: contextShowAction,
	}
}

func contextShowAction(c *cli.Context) error {
	log := logger.FromContext(c.Context)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, exists := cfg.Contexts[cfg.CurrentContext]
	if !exists {
		return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
	}

	log.Info("Current context", zap.String("name", cfg.CurrentContext))

	// Output as YAML
	contextData := map[string]interface{}{
		"executor-address": ctx.ExecutorAddress,
		"avs-address":      ctx.AVSAddress,
		"operator-set-id":  ctx.OperatorSetID,
		"rpc-url":          ctx.RPCUrl,
		"release-manager":  ctx.ReleaseManagerAddress,
		"network-id":       ctx.NetworkID,
	}

	// Add environment variables if any
	if len(ctx.EnvironmentVars) > 0 {
		contextData["environment-vars"] = ctx.EnvironmentVars
	}

	data := map[string]interface{}{
		"current-context": cfg.CurrentContext,
		"context":         contextData,
	}

	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()

	return encoder.Encode(data)
}