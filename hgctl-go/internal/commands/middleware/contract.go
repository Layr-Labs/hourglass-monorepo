package middleware

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

// ContractBeforeFunc initializes a contract client and stores it in the context
// for commands that need blockchain interaction
func ContractBeforeFunc(c *cli.Context) error {
	log := GetLogger(c)

	// Get current context
	currentCtx, ok := c.Context.Value(config.ContextKey).(*config.Context)
	if !ok || currentCtx == nil {
		log.Info("No context configured, contract client will not be initialized")
		return nil
	}
	// Check if we need a contract client based on the command
	if !needsContractClient(c) {
		return nil
	}

	// Get AVS address from context or environment
	avsAddress := currentCtx.AVSAddress
	if avsAddress == "" {
		avsAddress = os.Getenv("AVS_ADDRESS")
	}

	// Get operator address from context or environment
	operatorAddress := currentCtx.OperatorAddress
	if operatorAddress == "" {
		operatorAddress = os.Getenv("OPERATOR_ADDRESS")
	}

	// For some commands, we might not need both addresses
	if !validateAddresses(c, avsAddress, operatorAddress) {
		log.Debug("Required addresses not configured for command, skipping contract client initialization")
		return nil
	}

	// Get RPC URL
	rpcURL := currentCtx.L1RPCUrl
	if rpcURL == "" {
		rpcURL = os.Getenv("ETH_RPC_URL")
	}
	if rpcURL == "" {
		return fmt.Errorf("RPC URL not configured")
	}

	// Get private key (only from environment for security)
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		privateKey = os.Getenv("OPERATOR_PRIVATE_KEY")
	}

	// Build contract configuration
	contractConfig := &client.ContractConfig{
		AVSAddress:      avsAddress,
		OperatorAddress: operatorAddress,
	}

	// Apply contract overrides from context if available
	if currentCtx.ContractOverrides != nil {
		contractConfig.DelegationManager = currentCtx.ContractOverrides.DelegationManager
		contractConfig.AllocationManager = currentCtx.ContractOverrides.AllocationManager
		contractConfig.StrategyManager = currentCtx.ContractOverrides.StrategyManager
		contractConfig.KeyRegistrar = currentCtx.ContractOverrides.KeyRegistrar
		contractConfig.ReleaseManager = currentCtx.ContractOverrides.ReleaseManager
	}

	// Create contract client
	contractClient, err := client.NewContractClient(rpcURL, privateKey, log, contractConfig)
	if err != nil {
		return fmt.Errorf("failed to create contract client: %w", err)
	}

	// Store in context
	c.Context = context.WithValue(c.Context, config.ContractClientKey, contractClient)

	log.Debug("Contract client initialized",
		zap.String("avsAddress", avsAddress),
		zap.String("operatorAddress", operatorAddress),
		zap.String("rpcURL", rpcURL),
		zap.Bool("hasSigner", privateKey != ""),
	)

	return nil
}

// needsContractClient determines if a command needs a contract client
func needsContractClient(c *cli.Context) bool {
	// List of commands that need contract client
	contractCommands := map[string]bool{
		"register":             true,
		"register-avs":         true,
		"register-key":         true,
		"deposit":              true,
		"delegate":             true,
		"allocate":             true,
		"set-allocation-delay": true,
		"operatorset":          true,
		"get":                  true,
		"describe":             true,
		"deploy":               true,
		"translate":            true,
	}

	// Check if the current command needs a contract client
	if c.NArg() == 0 {
		return false
	}

	command := c.Args().Get(0)
	return contractCommands[command]
}

// validateAddresses checks if the required addresses are available for the command
func validateAddresses(c *cli.Context, avsAddress, operatorAddress string) bool {
	command := c.Args().Get(0)

	// Commands that only need operator address
	operatorOnlyCommands := map[string]bool{
		"register":             true,
		"delegate":             true,
		"deposit":              true,
		"set-allocation-delay": true,
	}

	// Commands that need both AVS and operator addresses
	bothAddressCommands := map[string]bool{
		"register-avs": true,
		"register-key": true,
		"allocate":     true,
		"operatorset":  true,
		"get":          true,
		"describe":     true,
		"deploy":       true,
	}

	if operatorOnlyCommands[command] {
		return operatorAddress != ""
	}

	if bothAddressCommands[command] {
		return avsAddress != "" && operatorAddress != ""
	}

	return false
}

// GetContractClient retrieves the contract client from the context
func GetContractClient(c *cli.Context) (*client.ContractClient, error) {
	contractClient, ok := c.Context.Value(config.ContractClientKey).(*client.ContractClient)
	if !ok || contractClient == nil {
		return nil, fmt.Errorf("contract client not initialized. Please configure AVS address, operator address, and RPC URL using `hgctl context set`")
	}
	return contractClient, nil
}

// CleanupContractClient closes the contract client if it exists
func CleanupContractClient(c *cli.Context) error {
	if contractClient, ok := c.Context.Value(config.ContractClientKey).(*client.ContractClient); ok && contractClient != nil {
		contractClient.Close()
	}
	return nil
}
