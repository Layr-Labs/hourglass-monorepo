package harness

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/stretchr/testify/assert"

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

	// Pre-generated operatorKeystores
	operatorKeystores map[string]*PreGeneratedKeystore

	// Pre-generated systemKeystores
	systemKeystores map[string]*PreGeneratedKeystore

	// CLI app instance
	app *cli.App
}

// Keystore name constants
const (
	KeystoreAggregatorBN254          = "aggregator-bn254"
	KeystoreAggregatorECDSA          = "aggregator-ecdsa"
	KeystoreAggregatorSystem         = "aggregator-system"
	KeystoreExecutorBN254            = "executor-bn254"
	KeystoreExecutorECDSA            = "executor-ecdsa"
	KeystoreExecutorSystem           = "executor-system"
	KeystoreExecutor2ECDSA           = "executor2-ecdsa"
	KeystoreExecutor2System          = "executor2-system"
	KeystoreExecutor3ECDSA           = "executor3-ecdsa"
	KeystoreExecutor3System          = "executor3-system"
	KeystoreExecutor4ECDSA           = "executor4-ecdsa"
	KeystoreExecutor4System          = "executor4-system"
	KeystoreUnregistered1ECDSA       = "unregistered-operator1-ecdsa"
	KeystoreUnregistered1SystemBN254 = "unregistered1-system-bn254"
	KeystoreUnregistered1SystemECDSA = "unregistered1-system-ecdsa"
	KeystoreUnregistered2ECDSA       = "unregistered-operator2-ecdsa"
	KeystoreUnregistered2SystemBN254 = "unregistered2-system-bn254"
	KeystoreUnregistered2SystemECDSA = "unregistered2-system-ecdsa"
)

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
		ContextName:       "test",
		logger:            l,
		t:                 t,
		cleanupFns:        []func(){},
		operatorKeystores: make(map[string]*PreGeneratedKeystore),
		systemKeystores:   make(map[string]*PreGeneratedKeystore),
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

	// 3. Initialize pre-generated operatorKeystores
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

// initializeKeystores sets up references to pre-generated operatorKeystores
func (h *TestHarness) initializeKeystores() {
	// Operator ECDSA keystores (for signing transactions as the operator)
	aggregatorECDSAPath := strings.Replace(h.ChainConfig.OperatorKeystorePath,
		"aggregator-keystore.json", "aggregator-ecdsa-keystore.json", 1)
	h.operatorKeystores[KeystoreAggregatorECDSA] = &PreGeneratedKeystore{
		Path:     aggregatorECDSAPath,
		Password: h.ChainConfig.OperatorKeystorePassword,
		Address:  h.ChainConfig.OperatorAccountAddress,
		Type:     "ecdsa",
	}

	executorECDSAPath := strings.Replace(h.ChainConfig.ExecOperatorKeystorePath,
		"executor-keystore.json", "executor-ecdsa-keystore.json", 1)
	h.operatorKeystores[KeystoreExecutorECDSA] = &PreGeneratedKeystore{
		Path:     executorECDSAPath,
		Password: h.ChainConfig.ExecOperatorKeystorePassword,
		Address:  h.ChainConfig.ExecOperatorAccountAddress,
		Type:     "ecdsa",
	}

	h.operatorKeystores[KeystoreExecutor2ECDSA] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperator2KeystorePath,
		Password: h.ChainConfig.ExecOperator2KeystorePassword,
		Address:  h.ChainConfig.ExecOperator2AccountAddress,
		Type:     "ecdsa",
	}

	h.operatorKeystores[KeystoreExecutor3ECDSA] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperator3KeystorePath,
		Password: h.ChainConfig.ExecOperator3KeystorePassword,
		Address:  h.ChainConfig.ExecOperator3AccountAddress,
		Type:     "ecdsa",
	}

	h.operatorKeystores[KeystoreExecutor4ECDSA] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperator4KeystorePath,
		Password: h.ChainConfig.ExecOperator4KeystorePassword,
		Address:  h.ChainConfig.ExecOperator4AccountAddress,
		Type:     "ecdsa",
	}

	h.operatorKeystores[KeystoreUnregistered1ECDSA] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.UnregisteredOperator1KeystorePath,
		Password: h.ChainConfig.UnregisteredOperator1KeystorePassword,
		Address:  h.ChainConfig.UnregisteredOperator1AccountAddress,
		Type:     "ecdsa",
	}

	h.operatorKeystores[KeystoreUnregistered2ECDSA] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.UnregisteredOperator2KeystorePath,
		Password: h.ChainConfig.UnregisteredOperator2KeystorePassword,
		Address:  h.ChainConfig.UnregisteredOperator2AccountAddress,
		Type:     "ecdsa",
	}

	// System keystores (BN254 is ONLY for system keys, not operator keys)
	// Aggregator system keys (BN254 + ECDSA)
	h.systemKeystores[KeystoreAggregatorBN254] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.OperatorKeystorePath,
		Password: h.ChainConfig.OperatorKeystorePassword,
		Address:  h.ChainConfig.OperatorAccountAddress,
		Type:     "bn254",
	}

	h.systemKeystores[KeystoreAggregatorSystem] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.OperatorSystemKeystorePath,
		Password: h.ChainConfig.OperatorSystemKeystorePassword,
		Address:  h.ChainConfig.OperatorSystemAddress,
		Type:     "ecdsa",
	}

	// Executor 1 system keys (BN254 + ECDSA)
	h.systemKeystores[KeystoreExecutorBN254] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperatorKeystorePath,
		Password: h.ChainConfig.ExecOperatorKeystorePassword,
		Address:  h.ChainConfig.ExecOperatorAccountAddress,
		Type:     "bn254",
	}

	h.systemKeystores[KeystoreExecutorSystem] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperatorSystemKeystorePath,
		Password: h.ChainConfig.ExecOperatorSystemKeystorePassword,
		Address:  h.ChainConfig.ExecOperatorSystemAddress,
		Type:     "ecdsa",
	}

	// Executor 2 system key (ECDSA only)
	h.systemKeystores[KeystoreExecutor2System] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperator2SystemKeystorePath,
		Password: h.ChainConfig.ExecOperator2SystemKeystorePassword,
		Address:  h.ChainConfig.ExecOperator2SystemAddress,
		Type:     "ecdsa",
	}

	// Executor 3 system key (ECDSA only)
	h.systemKeystores[KeystoreExecutor3System] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperator3SystemKeystorePath,
		Password: h.ChainConfig.ExecOperator3SystemKeystorePassword,
		Address:  h.ChainConfig.ExecOperator3SystemAddress,
		Type:     "ecdsa",
	}

	// Executor 4 system key (ECDSA only)
	h.systemKeystores[KeystoreExecutor4System] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.ExecOperator4SystemKeystorePath,
		Password: h.ChainConfig.ExecOperator4SystemKeystorePassword,
		Address:  h.ChainConfig.ExecOperator4SystemAddress,
		Type:     "ecdsa",
	}

	// Unregistered operator 1 system keys (BN254 + ECDSA for testing)
	h.systemKeystores[KeystoreUnregistered1SystemBN254] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.UnregisteredOperator1SystemBN254KeystorePath,
		Password: h.ChainConfig.UnregisteredOperator1SystemBN254KeystorePassword,
		Address:  "", // BN254 keystores don't have ECDSA addresses
		Type:     "bn254",
	}

	h.systemKeystores[KeystoreUnregistered1SystemECDSA] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.UnregisteredOperator1SystemECDSAKeystorePath,
		Password: h.ChainConfig.UnregisteredOperator1SystemECDSAKeystorePassword,
		Address:  h.ChainConfig.UnregisteredOperator1SystemECDSAAddress,
		Type:     "ecdsa",
	}

	// Unregistered operator 2 system keys (BN254 + ECDSA for testing)
	h.systemKeystores[KeystoreUnregistered2SystemBN254] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.UnregisteredOperator2SystemBN254KeystorePath,
		Password: h.ChainConfig.UnregisteredOperator2SystemBN254KeystorePassword,
		Address:  "", // BN254 keystores don't have ECDSA addresses
		Type:     "bn254",
	}

	h.systemKeystores[KeystoreUnregistered2SystemECDSA] = &PreGeneratedKeystore{
		Path:     h.ChainConfig.UnregisteredOperator2SystemECDSAKeystorePath,
		Password: h.ChainConfig.UnregisteredOperator2SystemECDSAKeystorePassword,
		Address:  h.ChainConfig.UnregisteredOperator2SystemECDSAAddress,
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
	return h.ExecuteCLIWithOperatorKeystore("", args...)
}

// ConfigureSystemKey configures a system keystore for use with hgctl signer system
func (h *TestHarness) ConfigureSystemKey(keystoreName string) error {
	// Fetch system keystore
	keystore, exists := h.systemKeystores[keystoreName]
	if !exists {
		return fmt.Errorf("unknown system keystore: %s", keystoreName)
	}

	if keystore.Type == "ecdsa" {
		privateKey, err := convertKeystoreToPrivateKey(keystore.Path, keystore.Password)
		if err != nil {
			return err
		}
		err = os.Setenv("SYSTEM_PRIVATE_KEY", privateKey)
		if err != nil {
			return err
		}
		signerArgs := []string{"hgctl", "signer", "system", "privatekey"}
		if err := h.app.RunContext(context.Background(), signerArgs); err != nil {
			return fmt.Errorf("failed to configure system keystore: %w", err)
		}
		return nil
	}

	// Set keystore password environment variable
	err := os.Setenv("SYSTEM_KEYSTORE_PASSWORD", keystore.Password)
	if err != nil {
		return err
	}

	// Configure the system signer with this keystore
	signerArgs := []string{"hgctl", "signer", "system", "keystore", "--name", keystoreName, "--type", keystore.Type}
	if err := h.app.RunContext(context.Background(), signerArgs); err != nil {
		return fmt.Errorf("failed to configure system keystore: %w", err)
	}

	return nil
}

// ExecuteCLIWithOperatorKeystore runs hgctl with a specific keystore
func (h *TestHarness) ExecuteCLIWithOperatorKeystore(keystoreName string, args ...string) (*CLIResult, error) {
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

	// Configure operator keystore if specified
	if keystoreName != "" {
		keystore, exists := h.operatorKeystores[keystoreName]
		if !exists {
			return nil, fmt.Errorf("unknown operator keystore: %s", keystoreName)
		}

		// Update context with the operator address for this keystore
		contextArgs := []string{"hgctl", "context", "set", "--operator-address", keystore.Address}
		if err := h.app.RunContext(context.Background(), contextArgs); err != nil {
			return nil, fmt.Errorf("failed to set operator address in context: %w", err)
		}

		// Set keystore password environment variable
		originalPwd := os.Getenv("OPERATOR_KEYSTORE_PASSWORD")
		envToRestore = append(envToRestore, struct{ key, value string }{"OPERATOR_KEYSTORE_PASSWORD", originalPwd})
		os.Setenv("OPERATOR_KEYSTORE_PASSWORD", keystore.Password)

		// Configure the operator signer with this keystore
		signerArgs := []string{"hgctl", "signer", "operator", "keystore", "--name", keystoreName}
		if err := h.app.RunContext(context.Background(), signerArgs); err != nil {
			return nil, fmt.Errorf("failed to configure operator keystore: %w", err)
		}
	}

	// Simple output capture
	outputBuf := &bytes.Buffer{}

	h.app.Writer = outputBuf
	h.app.ErrWriter = outputBuf

	// Run the command
	allArgs := append([]string{"hgctl"}, args...)
	err := h.app.RunContext(context.Background(), allArgs)

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
		result.Stderr = err.Error()
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
		{"--executor-endpoint", "127.0.0.1:9090"},
		{"--avs-address", h.ChainConfig.AVSAccountAddress},
		{"--operator-address", h.ChainConfig.OperatorAccountAddress},
		{"--l1-rpc-url", h.ChainConfig.L1RPC}, // Critical: contract middleware needs this
		{"--l2-rpc-url", h.ChainConfig.L2RPC}, // Include L2 RPC for completeness
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

	// Configure system BN254 keystore only
	if err := h.withEnv("SYSTEM_KEYSTORE_PASSWORD", h.ChainConfig.OperatorKeystorePassword, func() error {
		// BN254 signer
		result, err := h.ExecuteCLI("signer", "system", "keystore",
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

// GetAggregatorBN254Keystore returns the aggregator BN254 keystore
func (h *TestHarness) GetAggregatorBN254Keystore() *PreGeneratedKeystore {
	return h.systemKeystores["aggregator-bn254"]
}

// GetAggregatorECDSAKeystore returns the aggregator ECDSA keystore
func (h *TestHarness) GetAggregatorECDSAKeystore() *PreGeneratedKeystore {
	return h.operatorKeystores["aggregator-ecdsa"]
}

// GetExecutorBN254Keystore returns the executor BN254 keystore
func (h *TestHarness) GetExecutorBN254Keystore() *PreGeneratedKeystore {
	return h.systemKeystores["executor-bn254"]
}

// GetExecutorECDSAKeystore returns the executor ECDSA keystore
func (h *TestHarness) GetExecutorECDSAKeystore() *PreGeneratedKeystore {
	return h.operatorKeystores["executor-ecdsa"]
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

// ConvertKeystoreToPrivateKey converts an ECDSA keystore to a hex-encoded private key
func convertKeystoreToPrivateKey(keystorePath, password string) (string, error) {
	// Clean and validate the path
	keystorePath = filepath.Clean(keystorePath)

	// Read the keystore file
	keyStoreContents, err := os.ReadFile(keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore at %s: %w", keystorePath, err)
	}

	// Decrypt the keystore
	key, err := keystore.DecryptKey(keyStoreContents, password)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	// Convert to hex string with 0x prefix
	privateKeyHex := hex.EncodeToString(key.PrivateKey.D.Bytes())
	return "0x" + privateKeyHex, nil
}
