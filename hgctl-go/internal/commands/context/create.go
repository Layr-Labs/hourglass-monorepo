package context

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a new context",
		ArgsUsage: "[context-name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "use",
				Usage: "Set as current context",
				Value: true,
			},
		},
		Action: contextCreateAction,
	}
}

func contextCreateAction(c *cli.Context) error {
	log := logger.FromContext(c.Context)

	name := c.Args().Get(0)

	// If still no name, show help
	if name == "" {
		return cli.ShowSubcommandHelp(c)
	}

	setCurrent := c.Bool("use")

	// Load existing config or create new one
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config doesn't exist, create a new one
		cfg = &config.Config{
			Contexts: make(map[string]*config.Context),
		}
	}

	// Check if context already exists
	if _, exists := cfg.Contexts[name]; exists {
		return fmt.Errorf("context '%s' already exists", name)
	}

	log.Info("Creating new context", zap.String("name", name))

	// Prompt for L1 RPC URL
	l1RPCURL, err := output.InputString(
		"Enter L1 RPC URL",
		"The RPC endpoint URL for the L1 network (e.g., http://localhost:8545)",
		"",
		validateRPCURL,
	)
	if err != nil {
		return fmt.Errorf("failed to get L1 RPC URL: %w", err)
	}

	// Connect to RPC and get chain ID
	log.Info("Retrieving Chain ID from L1 RPC.")
	ethClient, err := ethclient.Dial(l1RPCURL)
	if err != nil {
		return fmt.Errorf("failed to connect to L1 RPC: %w", err)
	}
	defer ethClient.Close()

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	log.Info("Connected! L1 Chain ID:", zap.String("ChainID", chainID.String()))

	// Prompt for operator address
	operatorAddress, err := output.InputString(
		"Enter operator address",
		"The Ethereum address of the operator (e.g., 0x1234...)",
		"",
		validateEthereumAddress,
	)
	if err != nil {
		return fmt.Errorf("failed to get operator address: %w", err)
	}

	// Create new context with the provided information
	ctx := &config.Context{
		L1ChainID:       chainID.Uint64(),
		L1RPCUrl:        l1RPCURL,
		OperatorAddress: operatorAddress,
	}

	// Add to config
	cfg.Contexts[name] = ctx

	// Set as current if requested
	if setCurrent {
		cfg.CurrentContext = name
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Info("Context created",
		zap.String("name", name),
		zap.Uint64("l1ChainId", ctx.L1ChainID),
		zap.String("operatorAddress", ctx.OperatorAddress),
		zap.Bool("current", setCurrent))

	log.Info("Context created successfully", zap.String("ContextName", name))
	log.Info("Saved L1 RPC URL", zap.String("L1RPCUrl", ctx.L1RPCUrl))
	log.Info("Saved Operator Address", zap.String("OperatorAddress", ctx.OperatorAddress))
	log.Info("Retrieved L1 ChainID", zap.Uint64("ChainID", ctx.L1ChainID))

	if setCurrent {
		fmt.Printf("\nCurrent context set to '%s'\n", name)
	}

	return nil
}

// validateRPCURL validates that the provided string is a valid RPC URL
func validateRPCURL(input string) error {
	if input == "" {
		return fmt.Errorf("RPC URL cannot be empty")
	}

	// Basic URL validation - must start with http:// or https:// or ws:// or wss://
	urlPattern := regexp.MustCompile(`^(https?|wss?)://`)
	if !urlPattern.MatchString(input) {
		return fmt.Errorf("RPC URL must start with http://, https://, ws://, or wss://")
	}

	return nil
}

// validateEthereumAddress validates that the provided string is a valid Ethereum address
func validateEthereumAddress(input string) error {
	if input == "" {
		return fmt.Errorf("operator address cannot be empty")
	}

	// Remove 0x prefix for validation if present
	address := strings.TrimPrefix(input, "0x")

	// Check if it's a valid hex string of correct length (40 characters = 20 bytes)
	if len(address) != 40 {
		return fmt.Errorf("invalid Ethereum address: must be 40 hex characters (20 bytes)")
	}

	// Check if it's a valid hex string
	if !regexp.MustCompile("^[0-9a-fA-F]+$").MatchString(address) {
		return fmt.Errorf("invalid Ethereum address: must contain only hexadecimal characters")
	}

	// Validate using go-ethereum's common.IsHexAddress
	if !common.IsHexAddress(input) {
		return fmt.Errorf("invalid Ethereum address format")
	}

	return nil
}
