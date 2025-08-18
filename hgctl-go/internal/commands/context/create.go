package context

import (
	"context"
	"fmt"
	"math/big"
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
		ArgsUsage: "hgctl context create <name>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "use",
				Usage: "Set as current context",
				Value: true,
			},
			&cli.StringFlag{
				Name:  "l1-rpc-url",
				Usage: "Set the L1 RPC URL for the context",
			},
			&cli.StringFlag{
				Name:  "l2-rpc-url",
				Usage: "Set the L2 RPC URL for the context",
			},
			&cli.StringFlag{
				Name:  "operator-address",
				Usage: "Set the operator address for the context",
			},
		},
		Action: contextCreateAction,
	}
}

func contextCreateAction(c *cli.Context) error {
	log := config.LoggerFromContext(c.Context)

	name := c.Args().Get(0)
	if name == "" {
		return cli.ShowSubcommandHelp(c)
	}

	// Load existing config or create new one
	cfg, err := loadOrCreateConfig()
	if err != nil {
		return err
	}

	// Check if context already exists
	if _, exists := cfg.Contexts[name]; exists {
		return fmt.Errorf("context '%s' already exists", name)
	}

	log.Info("Creating new context", zap.String("name", name))

	// Get L1 RPC URL
	l1RPCURL, err := getL1RPCURL(c)
	if err != nil {
		return err
	}

	// Get chain ID from RPC
	chainID, err := getChainID(l1RPCURL, log)
	if err != nil {
		return err
	}

	// Get L2 RPC URL
	l2RPCURL, err := getL2RPCURL(c)
	if err != nil {
		return err
	}

	// Get chain ID from L2 RPC
	l2ChainID, err := getChainIDFromRPC(l2RPCURL, log)
	if err != nil {
		return err
	}

	// Get operator address
	operatorAddress, err := getOperatorAddress(c)
	if err != nil {
		return err
	}

	// Create and save the context
	ctx := createContext(chainID.Uint64(), l1RPCURL, operatorAddress, l2ChainID.Uint64(), l2RPCURL)
	if err := saveContext(cfg, name, ctx, c.Bool("use")); err != nil {
		return err
	}

	// Log success
	logContextCreated(log, name, ctx, c.Bool("use"))

	return nil
}

// loadOrCreateConfig loads existing config or creates a new one
func loadOrCreateConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config doesn't exist, create a new one
		return &config.Config{
			Contexts: make(map[string]*config.Context),
		}, nil
	}
	return cfg, nil
}

// getL1RPCURL gets the L1 RPC URL from flag or prompts for it
func getL1RPCURL(c *cli.Context) (string, error) {
	l1RPCURL := c.String("l1-rpc-url")
	if l1RPCURL == "" {
		// Prompt for L1 RPC URL if not provided via flag
		url, err := output.InputString(
			"Enter L1 RPC URL",
			"The RPC endpoint URL for the L1 network (e.g., http://localhost:8545)",
			"",
			validateRPCURL,
		)
		if err != nil {
			return "", fmt.Errorf("failed to get L1 RPC URL: %w", err)
		}
		return url, nil
	}

	// Validate the provided L1 RPC URL
	if err := validateRPCURL(l1RPCURL); err != nil {
		return "", fmt.Errorf("invalid L1 RPC URL: %w", err)
	}
	return l1RPCURL, nil
}

// getChainID connects to the RPC and retrieves the chain ID
func getChainID(rpcURL string, log logger.Logger) (*big.Int, error) {
	return getChainIDFromRPC(rpcURL, log)
}

// getChainIDFromRPC connects to the RPC and retrieves the chain ID with a label
func getChainIDFromRPC(rpcURL string, log logger.Logger) (*big.Int, error) {
	log.Info(fmt.Sprintf("Retrieving Chain ID from RPC."))

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer ethClient.Close()

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	log.Info(fmt.Sprintf("Connected! Chain ID:"), zap.String("ChainID", chainID.String()))
	return chainID, nil
}

// getL2RPCURL gets the L2 RPC URL from flag or prompts for it
func getL2RPCURL(c *cli.Context) (string, error) {
	l2RPCURL := c.String("l2-rpc-url")
	if l2RPCURL == "" {
		// Prompt for L2 RPC URL if not provided via flag
		url, err := output.InputString(
			"Enter L2 RPC URL",
			"The RPC endpoint URL for the L2 network (e.g., http://localhost:9545)",
			"",
			validateRPCURL,
		)
		if err != nil {
			return "", fmt.Errorf("failed to get L2 RPC URL: %w", err)
		}
		return url, nil
	}

	// Validate the provided L2 RPC URL
	if err := validateRPCURL(l2RPCURL); err != nil {
		return "", fmt.Errorf("invalid L2 RPC URL: %w", err)
	}
	return l2RPCURL, nil
}

// getOperatorAddress gets the operator address from flag or prompts for it
func getOperatorAddress(c *cli.Context) (string, error) {
	operatorAddress := c.String("operator-address")
	if operatorAddress == "" {
		// Prompt for operator address if not provided via flag
		address, err := output.InputString(
			"Enter operator address",
			"The Ethereum address of the operator (e.g., 0x1234...)",
			"",
			validateEthereumAddress,
		)
		if err != nil {
			return "", fmt.Errorf("failed to get operator address: %w", err)
		}
		return address, nil
	}

	// Validate the provided operator address
	if err := validateEthereumAddress(operatorAddress); err != nil {
		return "", fmt.Errorf("invalid operator address: %w", err)
	}
	return operatorAddress, nil
}

// createContext creates a new context with the provided information
func createContext(l1ChainID uint64, l1RPCURL, operatorAddress string, l2ChainID uint64, l2RPCURL string) *config.Context {
	return &config.Context{
		L1ChainID:       l1ChainID,
		L1RPCUrl:        l1RPCURL,
		L2ChainID:       l2ChainID,
		L2RPCUrl:        l2RPCURL,
		OperatorAddress: operatorAddress,
	}
}

// saveContext saves the context to the configuration
func saveContext(cfg *config.Config, name string, ctx *config.Context, setCurrent bool) error {
	cfg.Contexts[name] = ctx

	if setCurrent {
		cfg.CurrentContext = name
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if setCurrent {
		fmt.Printf("\nCurrent context set to '%s'\n", name)
	}

	return nil
}

// logContextCreated logs the successful context creation
func logContextCreated(log logger.Logger, name string, ctx *config.Context, setCurrent bool) {
	log.Info("Context created",
		zap.String("name", name),
		zap.Uint64("l1ChainId", ctx.L1ChainID),
		zap.Uint64("l2ChainId", ctx.L2ChainID),
		zap.String("operatorAddress", ctx.OperatorAddress),
		zap.Bool("current", setCurrent))

	log.Info("Context created successfully", zap.String("ContextName", name))
	log.Info("Saved L1 RPC URL", zap.String("L1RPCUrl", ctx.L1RPCUrl))
	log.Info("Retrieved L1 ChainID", zap.Uint64("L1ChainID", ctx.L1ChainID))
	log.Info("Saved L2 RPC URL", zap.String("L2RPCUrl", ctx.L2RPCUrl))
	log.Info("Retrieved L2 ChainID", zap.Uint64("L2ChainID", ctx.L2ChainID))
	log.Info("Saved Operator Address", zap.String("OperatorAddress", ctx.OperatorAddress))
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
