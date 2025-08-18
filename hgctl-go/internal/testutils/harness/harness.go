package harness

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// CLIResult represents the result of a CLI command execution
type CLIResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
}

// TestHarness provides a complete testing environment for hgctl integration tests
type TestHarness struct {
	// Chain management
	chainManager *ChainManager
	ChainConfig  *config.ChainConfig

	// Test environment
	TestDir     string
	ContextName string

	// Clients for verification
	L1Client *ethclient.Client
	L2Client *ethclient.Client

	// Cleanup tracking
	cleanupFns []func()

	// Test logger
	logger logger.Logger
	t      *testing.T

	// Pre-generated keystores
	keystores map[string]*PreGeneratedKeystore

	// CLI app instance
	app *cli.App
}

// PreGeneratedKeystore represents a keystore from the chain setup
type PreGeneratedKeystore struct {
	Path     string
	Password string
	Address  string
	Type     string // "ecdsa" or "bn254"
}

// NewTestHarness creates a new test harness instance
func NewTestHarness(t *testing.T) *TestHarness {
	// Create test logger
	l := logger.NewLogger(true)

	h := &TestHarness{
		ContextName: "test",
		logger:      l,
		t:           t,
		cleanupFns:  []func(){},
		keystores:   make(map[string]*PreGeneratedKeystore),
	}

	// Create the CLI app instance
	h.app = commands.Hgctl()

	return h
}

// Setup initializes the test environment
func (h *TestHarness) Setup() error {
	h.logger.Info("Setting up test harness")

	// 1. Initialize chain manager
	h.chainManager = NewChainManager(h.logger)

	// 2. Load chain configuration first
	chainConfig, err := h.chainManager.LoadChainConfig()
	if err != nil {
		return fmt.Errorf("failed to load chain config: %w", err)
	}
	h.ChainConfig = chainConfig

	// 3. Initialize pre-generated keystores
	h.initializeKeystores()

	// 4. Start chains for this test
	if err := h.StartChains(); err != nil {
		return fmt.Errorf("failed to start chains: %w", err)
	}

	//5. Use and populate test context
	if err := h.useTestContext(); err != nil {
		return fmt.Errorf("failed to use test context: %w", err)
	}

	h.logger.Debug("Test harness setup complete")
	return nil
}

// StartChains starts the Anvil chains for testing
func (h *TestHarness) StartChains() error {
	// Ensure chains are running
	if err := h.chainManager.EnsureChainsRunning(); err != nil {
		return fmt.Errorf("failed to ensure chains are running: %w", err)
	}

	// Add chain cleanup to the cleanup functions
	h.cleanupFns = append(h.cleanupFns, func() {
		err := h.chainManager.Cleanup()
		if err != nil {
			return
		}
	})

	// Create clients
	h.logger.Info("Connecting to chains")
	l1Client, err := ethclient.Dial(h.ChainConfig.L1RPC)
	if err != nil {
		return fmt.Errorf("failed to connect to L1: %w", err)
	}
	h.L1Client = l1Client

	l2Client, err := ethclient.Dial(h.ChainConfig.L2RPC)
	if err != nil {
		return fmt.Errorf("failed to connect to L2: %w", err)
	}
	h.L2Client = l2Client

	// Wait for both chains
	if err := h.chainManager.WaitForAnvil(l1Client); err != nil {
		return fmt.Errorf("L1 not ready: %w", err)
	}
	if err := h.chainManager.WaitForAnvil(l2Client); err != nil {
		return fmt.Errorf("L2 not ready: %w", err)
	}

	return nil
}

// initializeKeystores sets up references to pre-generated keystores
func (h *TestHarness) initializeKeystores() {
	// Aggregator/Operator keystores
	h.keystores["aggregator-bn254"] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.OperatorKeystorePath,
		Password: h.ChainConfig.OperatorKeystorePassword,
		Address:  h.ChainConfig.OperatorAccountAddress,
		Type:     "bn254",
	}

	// Aggregator ECDSA keystore
	aggregatorECDSAPath := strings.Replace(h.ChainConfig.OperatorKeystorePath,
		"aggregator-keystore.json", "aggregator-ecdsa-keystore.json", 1)
	h.keystores["aggregator-ecdsa"] = &PreGeneratedKeystore{
		Path:     aggregatorECDSAPath,
		Password: h.ChainConfig.OperatorKeystorePassword,
		Address:  h.ChainConfig.OperatorAccountAddress,
		Type:     "ecdsa",
	}

	// Executor keystores
	h.keystores["executor-bn254"] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperatorKeystorePath,
		Password: h.ChainConfig.ExecOperatorKeystorePassword,
		Address:  h.ChainConfig.ExecOperatorAccountAddress,
		Type:     "bn254",
	}

	// Executor ECDSA keystore
	executorECDSAPath := strings.Replace(h.ChainConfig.ExecOperatorKeystorePath,
		"executor-keystore.json", "executor-ecdsa-keystore.json", 1)
	h.keystores["executor-ecdsa"] = &PreGeneratedKeystore{
		Path:     executorECDSAPath,
		Password: h.ChainConfig.ExecOperatorKeystorePassword,
		Address:  h.ChainConfig.ExecOperatorAccountAddress,
		Type:     "ecdsa",
	}
}

// Teardown cleans up all test resources
func (h *TestHarness) Teardown() {
	h.logger.Debug("Tearing down test harness")

	// Execute cleanup functions in reverse order
	for i := len(h.cleanupFns) - 1; i >= 0; i-- {
		h.cleanupFns[i]()
	}

	// Close clients
	if h.L1Client != nil {
		h.L1Client.Close()
	}
	if h.L2Client != nil {
		h.L2Client.Close()
	}
}

// ExecuteCLI runs hgctl with the given arguments
func (h *TestHarness) ExecuteCLI(args ...string) (*CLIResult, error) {
	return h.ExecuteCLIWithKeystore("", args...)
}

// ExecuteCLIWithKeystore runs hgctl with a specific keystore
func (h *TestHarness) ExecuteCLIWithKeystore(keystoreName string, args ...string) (*CLIResult, error) {
	// Set up environment variables
	originalEnv := make(map[string]string)
	defer func() {
		// Restore original environment
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Set ETH_RPC_URL
	originalEnv["ETH_RPC_URL"] = os.Getenv("ETH_RPC_URL")
	os.Setenv("ETH_RPC_URL", h.ChainConfig.L1RPC)

	// If keystore specified, set up the environment and args
	if keystoreName != "" {
		_, exists := h.keystores[keystoreName]
		if !exists {
			return nil, fmt.Errorf("unknown keystore: %s", keystoreName)
		}

		// Set private key for both ECDSA and BN254 keystores that need transaction signing
		var privateKey string
		switch keystoreName {
		case "aggregator-ecdsa", "aggregator-bn254":
			privateKey = h.ChainConfig.OperatorAccountPk
		case "executor-ecdsa", "executor-bn254":
			privateKey = h.ChainConfig.ExecOperatorAccountPk
		}

		if privateKey != "" {
			if !strings.HasPrefix(privateKey, "0x") {
				privateKey = "0x" + privateKey
			}
			originalEnv["PRIVATE_KEY"] = os.Getenv("PRIVATE_KEY")
			os.Setenv("PRIVATE_KEY", privateKey)
		}
	}

	// Capture stdout and stderr in a single buffer
	// This is important because the logger might write to either
	outputBuf := &bytes.Buffer{}

	// Create a fresh app instance to avoid state pollution
	app := commands.Hgctl()

	// CRITICAL: Set both Writer and ErrWriter to the same buffer
	// This ensures all output (including logger output) is captured
	app.Writer = outputBuf
	app.ErrWriter = outputBuf

	// Prepend app name to args
	allArgs := append([]string{"hgctl"}, args...)

	// Run the app with a fresh context
	ctx := context.Background()
	err := app.RunContext(ctx, allArgs)

	// Get the complete output
	output := outputBuf.String()

	result := &CLIResult{
		Stdout: output, // All output goes here
		Stderr: "",     // Empty since we're capturing everything in one buffer
		Error:  err,
	}

	// You might want to split stdout/stderr based on log levels
	// For now, everything goes to Stdout for simplicity

	if err != nil {
		// Check if it's a cli error with exit code
		if exitErr, ok := err.(cli.ExitCoder); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.ExitCode = 0
	}

	return result, nil
}

// Alternative: Create a helper method that properly checks for output
func (h *TestHarness) AssertCLIOutput(t *testing.T, result *CLIResult, expectedStrings ...string) {
	require.NotNil(t, result, "CLI result should not be nil")

	// Combine stdout and stderr for checking
	allOutput := result.Stdout + result.Stderr

	for _, expected := range expectedStrings {
		assert.Contains(t, allOutput, expected,
			"Expected to find '%s' in output:\n%s", expected, allOutput)
	}
}

// ExecuteCLIWithInput runs hgctl with stdin input
func (h *TestHarness) ExecuteCLIWithInput(input string, args ...string) (*CLIResult, error) {
	// For now, since we're using the CLI directly, we need to handle input differently
	// This might require modifying commands to accept passwords via flags or env vars
	// For the test harness, we'll use password flags instead of stdin
	return h.ExecuteCLI(args...)
}

// WaitForTransaction waits for a transaction to be mined
func (h *TestHarness) WaitForTransaction(ctx context.Context, txHash string) error {
	return h.waitForTransactionOnChain(ctx, h.L1Client, txHash)
}

func (h *TestHarness) waitForTransactionOnChain(ctx context.Context, client *ethclient.Client, txHash string) error {
	h.logger.Info("Waiting for transaction", zap.String("hash", txHash))

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for transaction %s", txHash)
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
			if err == nil && receipt != nil {
				if receipt.Status == 0 {
					return fmt.Errorf("transaction %s failed", txHash)
				}
				h.logger.Info("Transaction mined",
					zap.String("hash", txHash),
					zap.Uint64("block", receipt.BlockNumber.Uint64()))
				return nil
			}
		}
	}
}

// useTestContext creates a minimal context config file for testing
func (h *TestHarness) useTestContext() error {
	// Create the context first
	if err := h.app.Run([]string{"hgctl", "context", "use", h.ContextName}); err != nil {
		return fmt.Errorf("failed to create context: %w", err)
	}
	// Set each configuration value using separate context set calls
	contextSetCalls := [][]string{
		{"hgctl", "context", "set", "--executor-address", "127.0.0.1:9090"},
		{"hgctl", "context", "set", "--delegation-manager", h.ChainConfig.DelegationManagerAddress},
		{"hgctl", "context", "set", "--allocation-manager", h.ChainConfig.AllocationManagerAddress},
		{"hgctl", "context", "set", "--strategy-manager", h.ChainConfig.StrategyManagerAddress},
		{"hgctl", "context", "set", "--release-manager", h.ChainConfig.ReleaseManagerAddress},
		{"hgctl", "context", "set", "--key-registrar", h.ChainConfig.KeyRegistrarAddress},
		{"hgctl", "context", "set", "--avs-address", h.ChainConfig.AVSAccountAddress},
		{"hgctl", "context", "set", "--operator-address", h.ChainConfig.OperatorAccountAddress},
		{"hgctl", "context", "show"},
	}

	// Execute each context set call
	for _, args := range contextSetCalls {
		if err := h.app.Run(args); err != nil {
			return fmt.Errorf("failed to set context property %v: %w", args, err)
		}
	}

	h.logger.Debug("Populated test context config", zap.String("context", h.ContextName))

	return nil
}

// GetBeaconETHStrategy returns the address of the BeaconETH strategy
func (h *TestHarness) GetBeaconETHStrategy() string {
	// This is a well-known address for BeaconETH strategy on testnet
	return "0xbeaC0eeEeeeeEEeEeEEEEeeEEeEeeeEeeEEBEaC0"
}

// GetAggregatorKeystore returns the aggregator BN254 keystore
func (h *TestHarness) GetAggregatorKeystore() *PreGeneratedKeystore {
	return h.keystores["aggregator-bn254"]
}

// GetAggregatorECDSAKeystore returns the aggregator ECDSA keystore
func (h *TestHarness) GetAggregatorECDSAKeystore() *PreGeneratedKeystore {
	return h.keystores["aggregator-ecdsa"]
}

// GetExecutorKeystore returns the executor BN254 keystore
func (h *TestHarness) GetExecutorKeystore() *PreGeneratedKeystore {
	return h.keystores["executor-bn254"]
}

// GetExecutorECDSAKeystore returns the executor ECDSA keystore
func (h *TestHarness) GetExecutorECDSAKeystore() *PreGeneratedKeystore {
	return h.keystores["executor-ecdsa"]
}

// MineBlocks mines blocks on the L1 chain
func (h *TestHarness) MineBlocks(count int) error {
	return h.chainManager.MineBlocks(h.ChainConfig.L1RPC, count)
}

// ParseTransactionHash extracts a transaction hash from CLI output
func ParseTransactionHash(output string) (string, error) {
	// Look for patterns like "Transaction: 0x..." or "tx: 0x..."
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)
		if strings.Contains(line, "transaction") || strings.Contains(line, "tx") {
			// Look for hex string
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "0x") && len(part) == 66 {
					return part, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no transaction hash found in output")
}
