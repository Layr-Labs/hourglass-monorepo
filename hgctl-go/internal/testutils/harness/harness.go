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
	// Simple environment restoration using defer
	var envToRestore []struct{ key, value string }
	defer func() {
		for _, env := range envToRestore {
			if env.value == "" {
				os.Unsetenv(env.key)
			} else {
				os.Setenv(env.key, env.value)
			}
		}
	}()

	// Set OPERATOR_PRIVATE_KEY for operations that need it
	// The operator signer is configured with privateKey: true, so it needs this env var
	if keystoreName != "" {
		if _, exists := h.keystores[keystoreName]; !exists {
			return nil, fmt.Errorf("unknown keystore: %s", keystoreName)
		}

		// Determine which private key to use based on keystore name
		var operatorPrivateKey string
		switch keystoreName {
		case "aggregator-ecdsa", "aggregator-bn254":
			operatorPrivateKey = h.ChainConfig.OperatorAccountPk
		case "executor-ecdsa", "executor-bn254":
			operatorPrivateKey = h.ChainConfig.ExecOperatorAccountPk
		}

		if operatorPrivateKey != "" {
			if !strings.HasPrefix(operatorPrivateKey, "0x") {
				operatorPrivateKey = "0x" + operatorPrivateKey
			}
			// Set OPERATOR_PRIVATE_KEY for the operator signer configuration
			originalValue := os.Getenv("OPERATOR_PRIVATE_KEY")
			envToRestore = append(envToRestore, struct{ key, value string }{"OPERATOR_PRIVATE_KEY", originalValue})
			os.Setenv("OPERATOR_PRIVATE_KEY", operatorPrivateKey)
		}
		
		// Also set keystore passwords if needed for system signers
		if strings.Contains(keystoreName, "aggregator") {
			originalPwd := os.Getenv("SYSTEM_KEYSTORE_PASSWORD")
			envToRestore = append(envToRestore, struct{ key, value string }{"SYSTEM_KEYSTORE_PASSWORD", originalPwd})
			os.Setenv("SYSTEM_KEYSTORE_PASSWORD", h.ChainConfig.OperatorKeystorePassword)
		} else if strings.Contains(keystoreName, "executor") {
			originalPwd := os.Getenv("SYSTEM_KEYSTORE_PASSWORD")
			envToRestore = append(envToRestore, struct{ key, value string }{"SYSTEM_KEYSTORE_PASSWORD", originalPwd})
			os.Setenv("SYSTEM_KEYSTORE_PASSWORD", h.ChainConfig.ExecOperatorKeystorePassword)
		}
	}

	// Simple output capture
	outputBuf := &bytes.Buffer{}

	// Create fresh app instance
	app := commands.Hgctl()
	app.Writer = outputBuf
	app.ErrWriter = outputBuf

	// Run the command
	allArgs := append([]string{"hgctl"}, args...)
	err := app.RunContext(context.Background(), allArgs)

	// Build result
	result := &CLIResult{
		Stdout: outputBuf.String(),
		Stderr: "",
		Error:  err,
	}

	if err != nil {
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

// useTestContext creates and configures the test context
func (h *TestHarness) useTestContext() error {
	// Step 1: Create/switch to test context
	result, err := h.ExecuteCLI("context", "use", h.ContextName)
	if err != nil && result.ExitCode != 0 {
		return fmt.Errorf("failed to create/use context: %w", err)
	}

	// Step 2: Configure all required context values in one batch
	// The context needs these for contract client initialization
	contextConfig := []struct {
		flag  string
		value string
	}{
		{"--executor-address", "127.0.0.1:9090"},
		{"--avs-address", h.ChainConfig.AVSAccountAddress},
		{"--operator-address", h.ChainConfig.OperatorAccountAddress},
		{"--l1-rpc-url", h.ChainConfig.L1RPC},  // Critical: contract middleware needs this
		{"--l2-rpc-url", h.ChainConfig.L2RPC},  // Include L2 RPC for completeness
	}

	for _, cfg := range contextConfig {
		result, err := h.ExecuteCLI("context", "set", cfg.flag, cfg.value)
		if err != nil && result.ExitCode != 0 {
			return fmt.Errorf("failed to set %s: %w", cfg.flag, err)
		}
	}

	// Step 3: Configure signers (simplified environment handling)
	if err := h.configureSigners(); err != nil {
		return fmt.Errorf("failed to configure signers: %w", err)
	}

	// Step 4: Verify context is properly configured
	result, err = h.ExecuteCLI("context", "show")
	if err != nil && result.ExitCode != 0 {
		return fmt.Errorf("failed to show context: %w", err)
	}
	h.logger.Debug("Context configured", zap.String("output", result.Stdout))

	return nil
}

// configureSigners sets up operator and system signers
func (h *TestHarness) configureSigners() error {
	// Configure operator signer
	operatorPK := h.ChainConfig.OperatorAccountPk
	if !strings.HasPrefix(operatorPK, "0x") {
		operatorPK = "0x" + operatorPK
	}

	// Use helper to temporarily set environment variable
	if err := h.withEnv("OPERATOR_PRIVATE_KEY", operatorPK, func() error {
		result, err := h.ExecuteCLI("signer", "operator", "privatekey")
		if err != nil && result.ExitCode != 0 {
			return fmt.Errorf("failed to configure operator signer: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	// Configure system signers
	if err := h.withEnv("SYSTEM_KEYSTORE_PASSWORD", h.ChainConfig.OperatorKeystorePassword, func() error {
		// ECDSA signer
		result, err := h.ExecuteCLI("signer", "system", "keystore",
			"--name", "aggregator-ecdsa", "--type", "ecdsa")
		if err != nil && result.ExitCode != 0 {
			return fmt.Errorf("failed to configure ECDSA signer: %w", err)
		}

		// BN254 signer
		result, err = h.ExecuteCLI("signer", "system", "keystore",
			"--name", "aggregator", "--type", "bn254")
		if err != nil && result.ExitCode != 0 {
			return fmt.Errorf("failed to configure BN254 signer: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// withEnv temporarily sets an environment variable for the duration of fn
func (h *TestHarness) withEnv(key, value string, fn func() error) error {
	original := os.Getenv(key)
	os.Setenv(key, value)
	defer func() {
		if original == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, original)
		}
	}()
	return fn()
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
