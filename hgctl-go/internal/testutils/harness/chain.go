package harness

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/config"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// ChainManager handles Anvil chain lifecycle for tests
type ChainManager struct {
	projectRoot string
	logger      logger.Logger
	l1Cmd       *exec.Cmd
	l2Cmd       *exec.Cmd
}

// NewChainManager creates a new chain manager instance
func NewChainManager(logger logger.Logger) *ChainManager {
	// Find project root by looking for the generateTestChainState.sh script
	projectRoot := findHourglassRoot()
	return &ChainManager{
		projectRoot: projectRoot,
		logger:      logger,
	}
}

// StartChains starts L1 and L2 Anvil chains
// If state files exist, it uses them. Otherwise, it runs the setup script first.
func (c *ChainManager) StartChains() (*exec.Cmd, *exec.Cmd, error) {
	// Check if state files exist
	l1StatePath := filepath.Join(c.projectRoot, "hgctl-go", "internal", "testdata", "anvil-l1-state.json")
	l2StatePath := filepath.Join(c.projectRoot, "hgctl-go", "internal", "testdata", "anvil-l2-state.json")
	configPath := filepath.Join(c.projectRoot, "hgctl-go", "internal", "testutils", "chainData", "chain-config.json")

	_, l1Err := os.Stat(l1StatePath)
	_, l2Err := os.Stat(l2StatePath)
	_, configErr := os.Stat(configPath)

	// If any required file is missing, run setup
	if os.IsNotExist(l1Err) || os.IsNotExist(l2Err) || os.IsNotExist(configErr) {
		c.logger.Info("State files missing, running chain setup script")
		if err := c.runSetupScript(); err != nil {
			return nil, nil, fmt.Errorf("failed to run setup script: %w", err)
		}
	} else {
		c.logger.Info("Using existing chain state files")
	}

	// Start anvil chains with the saved state
	l1Cmd, err := c.startAnvilL1()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start L1: %w", err)
	}

	l2Cmd, err := c.startAnvilL2()
	if err != nil {
		err := c.stopProcess(l1Cmd)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("failed to start L2: %w", err)
	}

	// Wait for both chains to be ready
	l1Client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		err := c.StopChains(l1Cmd, l2Cmd)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("failed to connect to L1: %w", err)
	}
	defer l1Client.Close()

	l2Client, err := ethclient.Dial("http://localhost:9545")
	if err != nil {
		err := c.StopChains(l1Cmd, l2Cmd)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("failed to connect to L2: %w", err)
	}
	defer l2Client.Close()

	if err := c.WaitForAnvil(l1Client); err != nil {
		err := c.StopChains(l1Cmd, l2Cmd)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("L1 failed to start: %w", err)
	}

	if err := c.WaitForAnvil(l2Client); err != nil {
		err := c.StopChains(l1Cmd, l2Cmd)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("L2 failed to start: %w", err)
	}

	c.logger.Info("Both chains are ready")
	return l1Cmd, l2Cmd, nil
}

// runSetupScript runs the chain setup script
func (c *ChainManager) runSetupScript() error {
	scriptPath := filepath.Join(c.projectRoot, "hgctl-go", "internal", "testutils", "scripts", "setup-test-chains.sh")

	// Make sure the script is executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Run the script which starts chains and generates config
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = c.projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run chain setup script: %w", err)
	}

	return nil
}

func (c *ChainManager) startAnvilL1() (*exec.Cmd, error) {
	statePath := filepath.Join(c.projectRoot, "hgctl-go", "internal", "testdata", "anvil-l1-state.json")

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("L1 state file not found at %s", statePath)
	}

	cmd := exec.Command("anvil",
		"--load-state", statePath,
		"--chain-id", "31337",
		"--port", "8545",
		"--block-time", "2",
		"--fork-url", "https://late-crimson-dew.quiknode.pro/56c000eadf175378343de407c56e0ccd62801fe9",
		"--fork-block-number", "23477799",
		"--silent",
	)

	// Capture output for debugging
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start L1 anvil: %w", err)
	}

	go logPipe("anvil-l1 stdout", stdoutPipe, c.logger)
	go logPipe("anvil-l1 stderr", stderrPipe, c.logger)

	return cmd, nil
}

func logPipe(name string, pipe io.ReadCloser, logger logger.Logger) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		logger.Info(name, zap.String("line", scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		logger.Warn("error reading from pipe", zap.String("pipe", name), zap.Error(err))
	}
}

// startAnvilL2 starts the L2 Anvil chain with saved state
func (c *ChainManager) startAnvilL2() (*exec.Cmd, error) {
	statePath := filepath.Join(c.projectRoot, "hgctl-go", "internal", "testdata", "anvil-l2-state.json")

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("L2 state file not found at %s", statePath)
	}

	cmd := exec.Command("anvil",
		"--load-state", statePath,
		"--chain-id", "31338",
		"--port", "9545",
		"--block-time", "2",
		//TODO: add fork url and height
	)

	c.logger.Info("Starting L2 Anvil", zap.String("state", statePath))

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start L2 anvil: %w", err)
	}

	// Give it time to start
	time.Sleep(2 * time.Second)

	return cmd, nil
}

// StopChains stops the running Anvil processes
func (c *ChainManager) StopChains(l1Cmd, l2Cmd *exec.Cmd) error {
	c.logger.Debug("Stopping test chains")

	var firstErr error

	if err := c.stopProcess(l1Cmd); err != nil && firstErr == nil {
		firstErr = err
	}

	if err := c.stopProcess(l2Cmd); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

func (c *ChainManager) stopProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// Try graceful shutdown first
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		// If that fails, force kill
		return cmd.Process.Kill()
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		// Force kill if it doesn't exit gracefully
		return cmd.Process.Kill()
	}
}

// LoadChainConfig loads the generated chain configuration
func (c *ChainManager) LoadChainConfig() (*config.ChainConfig, error) {
	configPath := filepath.Join(c.projectRoot, "hgctl-go", "internal", "testutils", "chainData", "chain-config.json")

	c.logger.Info("Loading chain config", zap.String("path", configPath))

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chain config: %w", err)
	}

	var cc config.ChainConfig
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, fmt.Errorf("failed to parse chain config: %w", err)
	}

	return &cc, nil
}

// WaitForAnvil waits for an Anvil instance to be ready
func (c *ChainManager) WaitForAnvil(client *ethclient.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for anvil to be ready")
		case <-ticker.C:
			// Try to get the chain ID
			_, err := client.ChainID(ctx)
			if err == nil {
				c.logger.Info("Anvil is ready")
				return nil
			}
			c.logger.Debug("Waiting for anvil", zap.Error(err))
		}
	}
}

// MineBlocks mines a specified number of blocks on the given chain
func (c *ChainManager) MineBlocks(rpcURL string, count int) error {
	c.logger.Info("Mining blocks", zap.String("rpc", rpcURL), zap.Int("count", count))

	cmd := exec.Command("cast", "rpc", "--rpc-url", rpcURL, "anvil_mine", fmt.Sprintf("%d", count))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mine blocks: %w, output: %s", err, string(output))
	}

	return nil
}

// SetBalance sets the ETH balance of an address
func (c *ChainManager) SetBalance(rpcURL string, address string, amount string) error {
	c.logger.Info("Setting balance", zap.String("address", address), zap.String("amount", amount))

	cmd := exec.Command("cast", "rpc", "--rpc-url", rpcURL, "anvil_setBalance", address, amount)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set balance: %w, output: %s", err, string(output))
	}

	return nil
}

// Helper function to find the hourglass-monorepo root
func findHourglassRoot() string {
	// Start from current directory and look for hourglass-monorepo
	dir, _ := os.Getwd()

	for {
		// Check if we're in hourglass-monorepo by looking for characteristic files
		if filepath.Base(dir) == "hourglass-monorepo" {
			// Verify by checking for expected subdirectories
			if _, err := os.Stat(filepath.Join(dir, "contracts")); err == nil {
				if _, err := os.Stat(filepath.Join(dir, "ponos")); err == nil {
					if _, err := os.Stat(filepath.Join(dir, "hgctl-go")); err == nil {
						return dir
					}
				}
			}
		}

		// Also check if we're inside hourglass-monorepo
		parentDir := filepath.Dir(dir)
		if filepath.Base(parentDir) == "hourglass-monorepo" {
			return parentDir
		}

		// Move up
		newDir := filepath.Dir(dir)
		if newDir == dir {
			// Reached root, default to relative path
			return "."
		}
		dir = newDir
	}
}

// EnsureChainsRunning ensures chains are running, starting them if needed
func (c *ChainManager) EnsureChainsRunning() error {
	// Check if chains are already running
	if c.l1Cmd != nil && c.l1Cmd.Process != nil {
		// Check if process is still alive
		if err := c.l1Cmd.Process.Signal(syscall.Signal(0)); err == nil {
			c.logger.Debug("L1 chain already running")
			return nil
		}
	}

	// Start chains if not running
	l1Cmd, l2Cmd, err := c.StartChains()
	if err != nil {
		return fmt.Errorf("failed to start chains: %w", err)
	}

	c.l1Cmd = l1Cmd
	c.l2Cmd = l2Cmd
	return nil
}

// Cleanup stops any running chains
func (c *ChainManager) Cleanup() error {
	return c.StopChains(c.l1Cmd, c.l2Cmd)
}
