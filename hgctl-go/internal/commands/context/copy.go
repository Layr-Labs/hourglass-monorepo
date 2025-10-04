package context

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
)

func copyCommand() *cli.Command {
	return &cli.Command{
		Name:      "copy",
		Usage:     "Copy an existing context",
		ArgsUsage: "<source-context-name>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "copy-name",
				Aliases: []string{"n"},
				Usage:   "Name for the copied context",
			},
			&cli.BoolFlag{
				Name:  "use",
				Usage: "Set the copied context as current",
				Value: false,
			},
		},
		Action: contextCopyAction,
	}
}

func contextCopyAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.ShowSubcommandHelp(c)
	}

	sourceName := c.Args().Get(0)
	copyName := c.String("copy-name")
	setCurrent := c.Bool("use")
	log := config.LoggerFromContext(c.Context)

	// Load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get source context
	sourceCtx, err := getSourceContext(cfg, sourceName)
	if err != nil {
		return err
	}

	// Get or prompt for copy name
	copyName, err = getCopyName(cfg, copyName)
	if err != nil {
		return err
	}

	// Perform the copy
	if err := performContextCopy(cfg, sourceCtx, copyName, setCurrent); err != nil {
		return err
	}

	// Log and display success
	logCopySuccess(log, sourceName, copyName, setCurrent)
	displayCopySuccess(sourceName, copyName, setCurrent)

	return nil
}

// getSourceContext validates and retrieves the source context
func getSourceContext(cfg *config.Config, sourceName string) (*config.Context, error) {
	sourceCtx, exists := cfg.Contexts[sourceName]
	if !exists {
		return nil, fmt.Errorf("source context '%s' not found", sourceName)
	}
	return sourceCtx, nil
}

// getCopyName gets the copy name from flag or prompts for it
func getCopyName(cfg *config.Config, copyName string) (string, error) {
	if copyName != "" {
		// Validate the provided copy-name
		if err := validateContextName(cfg, copyName); err != nil {
			return "", err
		}
		return copyName, nil
	}

	// Prompt for copy name
	return promptForCopyName(cfg)
}

// promptForCopyName prompts the user for a context name
func promptForCopyName(cfg *config.Config) (string, error) {
	return output.InputString(
		"Enter name for the copied context",
		"The name for the new context copy",
		"",
		func(input string) error {
			return validateContextName(cfg, input)
		},
	)
}

// validateContextName validates that a context name is valid and doesn't exist
func validateContextName(cfg *config.Config, name string) error {
	if name == "" {
		return fmt.Errorf("context name cannot be empty")
	}
	if _, exists := cfg.Contexts[name]; exists {
		return fmt.Errorf("context '%s' already exists", name)
	}
	return nil
}

// performContextCopy creates the copy and saves the config
func performContextCopy(cfg *config.Config, sourceCtx *config.Context, copyName string, setCurrent bool) error {
	// Deep copy the source context
	copiedCtx := deepCopyContext(sourceCtx)
	copiedCtx.Name = copyName

	// Add to config
	cfg.Contexts[copyName] = copiedCtx

	// Set as current if requested
	if setCurrent {
		cfg.CurrentContext = copyName
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// logCopySuccess logs the successful copy operation
func logCopySuccess(log logger.Logger, sourceName, copyName string, setCurrent bool) {
	log.Info("Context copied",
		zap.String("source", sourceName),
		zap.String("copy", copyName),
		zap.Bool("current", setCurrent))
}

// displayCopySuccess shows success message to the user
func displayCopySuccess(sourceName, copyName string, setCurrent bool) {
	fmt.Printf("Successfully copied context '%s' to '%s'\n", sourceName, copyName)
	if setCurrent {
		fmt.Printf("Current context set to '%s'\n", copyName)
	}
}

// deepCopyContext creates a deep copy of a context
func deepCopyContext(src *config.Context) *config.Context {
	dst := &config.Context{
		Name:             src.Name,
		ExecutorEndpoint: src.ExecutorEndpoint,
		AVSAddress:       src.AVSAddress,
		OperatorAddress:  src.OperatorAddress,
		OperatorSetID:    src.OperatorSetID,
		L1ChainID:        src.L1ChainID,
		L1RPCUrl:         src.L1RPCUrl,
		L2ChainID:        src.L2ChainID,
		L2RPCUrl:         src.L2RPCUrl,
		PrivateKey:       src.PrivateKey,
		EnvSecretsPath:   src.EnvSecretsPath,
	}

	// Copy slices and nested structures
	copyKeystores(src, dst)

	return dst
}

// copyKeystores deep copies the keystores slice
func copyKeystores(src, dst *config.Context) {
	if src.Keystores != nil {
		dst.Keystores = make([]signer.KeystoreReference, len(src.Keystores))
		copy(dst.Keystores, src.Keystores)
	}
}
