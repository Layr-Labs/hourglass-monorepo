package integration

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
)

var (
	// Global test harness instance
	testHarness *harness.TestHarness

	// Flags for test control
	skipSetup    = flag.Bool("skip-setup", false, "Skip chain setup (use existing chains)")
	keepChains   = flag.Bool("keep-chains", false, "Keep chains running after tests")
	verboseTests = flag.Bool("verbose", false, "Enable verbose test output")
)

// TestMain sets up the test environment once for all tests
func TestMain(m *testing.M) {
	flag.Parse()

	// Set up logging
	var logger *zap.Logger
	if *verboseTests {
		config := zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		logger, _ = config.Build()
	} else {
		logger = zap.NewNop()
	}

	defer logger.Sync()

	// Run tests
	exitCode := runTests(m, logger)
	os.Exit(exitCode)
}

func runTests(m *testing.M, logger *zap.Logger) int {
	if *skipSetup {
		logger.Info("Skipping chain setup, using existing chains")
		return m.Run()
	}

	// Create test harness
	t := &testing.T{}
	testHarness = harness.NewTestHarness(t)

	// Setup test environment
	logger.Info("Setting up test environment")
	if err := testHarness.Setup(); err != nil {
		logger.Error("Failed to setup test environment", zap.Error(err))
		return 1
	}

	// Run tests
	exitCode := m.Run()

	// Cleanup
	if !*keepChains {
		logger.Info("Tearing down test environment")
		testHarness.Teardown()
	} else {
		logger.Info("Keeping chains running (--keep-chains flag set)")
		fmt.Println("\nChains are still running:")
		fmt.Println("  L1 RPC: http://localhost:8545")
		fmt.Println("  L2 RPC: http://localhost:9545")
		fmt.Println("\nRemember to stop them manually!")
	}

	return exitCode
}

// Helper function to get a fresh test harness for isolated tests
func getTestHarness(t *testing.T) *harness.TestHarness {
	if testHarness == nil {
		t.Skip("Test harness not initialized. Run tests without -skip-setup flag.")
	}
	return testHarness
}

// Common test helpers

// skipIfShort skips the test if -short flag is set
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// requiresDocker skips the test if Docker is not available
func requiresDocker(t *testing.T) {
	// Check if Docker is available
	// This is a simplified check - could be made more robust
	if _, err := os.Stat("/var/run/docker.sock"); os.IsNotExist(err) {
		t.Skip("Test requires Docker")
	}
}

// requiresAnvil skips the test if Anvil is not installed
func requiresAnvil(t *testing.T) {
	// Check if anvil command exists
	// This is a simplified check
	_, err := exec.LookPath("anvil")
	if err != nil {
		t.Skip("Test requires Anvil (Foundry) to be installed")
	}
}

// Test environment validation
func TestEnvironmentSetup(t *testing.T) {
	skipIfShort(t)

	h := getTestHarness(t)

	// Verify chains are running
	t.Run("L1 Chain Running", func(t *testing.T) {
		chainID, err := h.L1Client.ChainID(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(31337), chainID.Int64())
	})

	t.Run("L2 Chain Running", func(t *testing.T) {
		chainID, err := h.L2Client.ChainID(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(31338), chainID.Int64())
	})

	// Verify contracts deployed
	t.Run("AVS Contract Deployed", func(t *testing.T) {
		require.NotEmpty(t, h.ChainConfig.AVSTaskRegistrarAddress)

		// Check contract code exists
		code, err := h.L1Client.CodeAt(context.Background(),
			common.HexToAddress(h.ChainConfig.AVSTaskRegistrarAddress), nil)
		require.NoError(t, err)
		require.NotEmpty(t, code, "AVS contract should have code")
	})

	// Verify CLI binary exists
	t.Run("CLI Binary Available", func(t *testing.T) {
		// CLI path is internal to harness, skip this check for now
		t.Skip("CLI path check not implemented")
	})

	// Verify test context created
	t.Run("Test Context Created", func(t *testing.T) {
		// Skip context check since we're using environment variables instead
		t.Skip("Context creation skipped - using environment variables")
	})
}

// Test data generators for common test scenarios

// generateTestKeystore creates a test keystore with a unique name
func generateTestKeystore(t *testing.T, h *harness.TestHarness, prefix string) string {
	name := fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())

	result, err := h.ExecuteCLIWithInput("test-password\ntest-password\n",
		"keystore", "create",
		"--context", h.ContextName,
		"--name", name,
		"--type", "ecdsa")

	require.NoError(t, err)
	harness.Assert(t, result).AssertKeystoreCreated(name)

	return name
}

// generateTestOperator creates and registers a test operator
func generateTestOperator(t *testing.T, h *harness.TestHarness) string {
	// Create keystore
	keystoreName := generateTestKeystore(t, h, "test-operator")

	// Register as operator
	result, err := h.ExecuteCLIWithInput("test-password\n",
		"register",
		"--context", h.ContextName,
		"--keystore", keystoreName,
		"--allocation-delay", "1")

	require.NoError(t, err)
	harness.Assert(t, result).AssertOperatorRegistered()

	// Extract operator address from output
	// This is a simplified example - real implementation would parse properly
	return keystoreName
}

// Benchmark helpers for performance testing

// benchmarkCLICommand benchmarks a CLI command execution
func benchmarkCLICommand(b *testing.B, h *harness.TestHarness, args ...string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := h.ExecuteCLI(args...)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Example benchmark
func BenchmarkContextList(b *testing.B) {
	h := getTestHarness(&testing.T{})
	benchmarkCLICommand(b, h, "context", "list")
}

// hasEigenLayerContracts checks if EigenLayer core contracts are deployed
func hasEigenLayerContracts(t *testing.T, h *harness.TestHarness) bool {
	// Try to get contract code at the expected DelegationManager address
	// This is the Sepolia address that our test config expects
	delegationManagerAddr := common.HexToAddress("0xd4a7e1bd8015057293f0d0a557088c286942e84b")
	
	code, err := h.L1Client.CodeAt(context.Background(), delegationManagerAddr, nil)
	if err != nil {
		t.Logf("Failed to check contract code: %v", err)
		return false
	}
	
	// If there's no code at the address, contracts aren't deployed
	if len(code) == 0 {
		t.Logf("No EigenLayer contracts found at expected addresses")
		return false
	}
	
	return true
}
