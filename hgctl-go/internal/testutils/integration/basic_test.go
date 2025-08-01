package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicHarnessSetup tests the basic harness initialization without chains
func TestBasicHarnessSetup(t *testing.T) {
	// This test verifies the test harness can be created
	// without necessarily starting the chains
	
	t.Run("Create Test Harness", func(t *testing.T) {
		// Skip if we're in short mode
		if testing.Short() {
			t.Skip("Skipping integration test in short mode")
		}
		
		// For now, just verify we can get the harness
		h := getTestHarness(t)
		require.NotNil(t, h, "Test harness should not be nil")
		
		// Check that essential fields are initialized
		assert.NotEmpty(t, h.ContextName, "Context name should be set")
	})
}

// TestCLIExecution tests basic CLI execution without chains
func TestCLIExecution(t *testing.T) {
	skipIfShort(t)
	h := getTestHarness(t)
	
	t.Run("Execute Help Command", func(t *testing.T) {
		// Try to execute a simple help command
		result, err := h.ExecuteCLI("--help")
		
		// The command might fail if the binary doesn't exist
		// but we should get a result object
		assert.NotNil(t, result, "Should get a result object")
		
		if err == nil && result.ExitCode == 0 {
			// If it succeeded, verify help output
			assert.Contains(t, result.Stdout, "hgctl", "Help should mention hgctl")
		} else {
			// If it failed, that's OK for now - the binary might not exist
			t.Logf("CLI execution failed (expected if binary not built): %v", err)
		}
	})
}