package middleware

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/urfave/cli/v2"
)

// RequireContext ensures a context is configured before running commands
func RequireContext(c *cli.Context) error {
	// Try to get current context from config
	_, err := config.GetCurrentContext()
	if err != nil {
		fmt.Fprintf(c.App.ErrWriter, "\nError: No context configured\n\n")
		fmt.Fprintf(c.App.ErrWriter, "To create a context, run:\n")
		fmt.Fprintf(c.App.ErrWriter, "  hgctl context create <name> --use\n\n")
		fmt.Fprintf(c.App.ErrWriter, "To list available contexts:\n")
		fmt.Fprintf(c.App.ErrWriter, "  hgctl context list\n\n")
		fmt.Fprintf(c.App.ErrWriter, "To use an existing context:\n")
		fmt.Fprintf(c.App.ErrWriter, "  hgctl context use <name>\n\n")
		return fmt.Errorf("no context configured - please create one first")
	}
	return nil
}