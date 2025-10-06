package context

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func deleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a context",
		ArgsUsage: "<context-name>",
		Description: `Delete an existing context by name.

Examples:
  # Delete a context
  hgctl context delete my-context

Note: You cannot delete the currently active context. Switch to another context first.`,
		Action: contextDeleteAction,
	}
}

func contextDeleteAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	contextName := c.Args().Get(0)
	log := config.LoggerFromContext(c.Context)

	// Load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate context exists
	if err := validateContextExists(cfg, contextName); err != nil {
		return err
	}

	// Prevent deleting the current context
	if err := validateNotCurrentContext(cfg, contextName); err != nil {
		return err
	}

	// Delete the context
	if err := deleteContext(cfg, contextName); err != nil {
		return err
	}

	// Log and display success
	logDeleteSuccess(log, contextName)
	displayDeleteSuccess(contextName)

	return nil
}

// validateContextExists checks if the context exists
func validateContextExists(cfg *config.Config, contextName string) error {
	if _, exists := cfg.Contexts[contextName]; !exists {
		return fmt.Errorf("context '%s' not found", contextName)
	}
	return nil
}

// validateNotCurrentContext ensures we're not deleting the active context
func validateNotCurrentContext(cfg *config.Config, contextName string) error {
	if cfg.CurrentContext == contextName {
		return fmt.Errorf("cannot delete current context '%s'. Switch to another context first using 'hgctl context use <context-name>'", contextName)
	}
	return nil
}

// deleteContext removes the context and saves the config
func deleteContext(cfg *config.Config, contextName string) error {
	// Remove from map
	delete(cfg.Contexts, contextName)

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// logDeleteSuccess logs the successful deletion
func logDeleteSuccess(log logger.Logger, contextName string) {
	log.Info("Context deleted",
		zap.String("context", contextName))
}

// displayDeleteSuccess shows success message to the user
func displayDeleteSuccess(contextName string) {
	fmt.Printf("Successfully deleted context '%s'\n", contextName)
}