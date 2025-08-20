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
		log.Debug("No context configured, contract client will not be initialized")
		return nil
	}

	log.Debug("Context found, initializing contract client",
		zap.String("avsAddress", currentCtx.AVSAddress),
		zap.String("operatorAddress", currentCtx.OperatorAddress),
		zap.String("l1RpcUrl", currentCtx.L1RPCUrl))

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

	// Get RPC URL
	rpcURL := currentCtx.L1RPCUrl
	if rpcURL == "" {
		return fmt.Errorf("l1 RPC URL not configured")
	}

	// Get operator private key from configured source (env var or keystore)
	privateKey, err := config.GetOperatorPrivateKey(currentCtx)
	if err != nil {
		return fmt.Errorf("failed to get operator private key: %w", err)
	}

	// Build contract configuration
	contractConfig := &client.ContractConfig{
		AVSAddress:      avsAddress,
		OperatorAddress: operatorAddress,
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

// GetContractClient retrieves the contract client from the context
func GetContractClient(c *cli.Context) (*client.ContractClient, error) {
	contractClient, ok := c.Context.Value(config.ContractClientKey).(*client.ContractClient)
	if !ok || contractClient == nil {
		return nil, fmt.Errorf("contract client not initialized. Please configure AVS address, operator address and l1 RPC URL using `hgctl context set`")
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
