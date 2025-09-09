package harness

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// AnvilInstance represents a running anvil instance
type AnvilInstance struct {
	Config  *AnvilConfig
	Process *os.Process
	Logger  *zap.SugaredLogger
	client  *ethclient.Client
}

// StartAnvil starts an anvil instance with the given configuration
func StartAnvil(ctx context.Context, config *AnvilConfig, log *zap.SugaredLogger) (*AnvilInstance, error) {
	args := []string{
		"--chain-id", fmt.Sprintf("%d", config.ChainID),
		"--port", fmt.Sprintf("%d", config.Port),
		"--block-time", "2",
	}
	
	// Add state loading if provided
	if config.StatePath != "" && fileExists(config.StatePath) {
		args = append(args, "--load-state", config.StatePath)
	} else if config.ForkURL != "" {
		args = append(args, "--fork-url", config.ForkURL)
		if config.ForkBlock > 0 {
			args = append(args, "--fork-block-number", fmt.Sprintf("%d", config.ForkBlock))
		}
	}
	
	// Add config output if specified
	if config.ConfigPath != "" {
		args = append(args, "--config-out", config.ConfigPath)
	}
	
	log.Infow("Starting anvil instance",
		"chainID", config.ChainID,
		"port", config.Port,
		"args", args,
	)
	
	cmd := exec.CommandContext(ctx, "anvil", args...)
	
	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start anvil: %w", err)
	}
	
	instance := &AnvilInstance{
		Config:  config,
		Process: cmd.Process,
		Logger:  log,
	}
	
	// Wait for anvil to be ready
	if err := instance.WaitForReady(ctx, 30*time.Second); err != nil {
		instance.Stop()
		return nil, fmt.Errorf("anvil failed to start: %w", err)
	}
	
	return instance, nil
}

// WaitForReady waits for the anvil instance to be ready
func (a *AnvilInstance) WaitForReady(ctx context.Context, timeout time.Duration) error {
	rpcURL := fmt.Sprintf("http://localhost:%d", a.Config.Port)
	
	ethClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   rpcURL,
		BlockType: ethereum.BlockType_Latest,
	}, &logger.Logger{Logger: a.Logger.Desugar()})
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			client, err := ethClient.GetEthereumContractCaller()
			if err == nil {
				// Try to get chain ID to verify it's working
				if chainID, err := client.ChainID(ctx); err == nil {
					a.Logger.Infow("Anvil instance ready",
						"chainID", chainID,
						"port", a.Config.Port,
					)
					a.client = client
					return nil
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	return fmt.Errorf("anvil failed to become ready within %s", timeout)
}

// DumpState dumps the current state of the anvil instance to a file
func (a *AnvilInstance) DumpState(ctx context.Context, outputPath string) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	rpcURL := fmt.Sprintf("http://localhost:%d", a.Config.Port)
	
	// Use cast to dump the state
	cmd := exec.CommandContext(ctx, "cast", "rpc", "--rpc-url", rpcURL, "anvil_dumpState")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to dump anvil state: %w", err)
	}
	
	// Write the state to file
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	
	a.Logger.Infow("Dumped anvil state",
		"path", outputPath,
		"size", len(output),
	)
	
	return nil
}

// Stop stops the anvil instance
func (a *AnvilInstance) Stop() error {
	if a.Process != nil {
		a.Logger.Info("Stopping anvil instance")
		return a.Process.Kill()
	}
	return nil
}

// GetClient returns the ethereum client for this anvil instance
func (a *AnvilInstance) GetClient() *ethclient.Client {
	return a.client
}

// LoadAnvilWithState starts an anvil instance with a pre-existing state
func LoadAnvilWithState(ctx context.Context, statePath string, port uint16, chainID uint64, log *zap.SugaredLogger) (*AnvilInstance, error) {
	config := &AnvilConfig{
		ChainID:   chainID,
		Port:      port,
		StatePath: statePath,
	}
	
	return StartAnvil(ctx, config, log)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}