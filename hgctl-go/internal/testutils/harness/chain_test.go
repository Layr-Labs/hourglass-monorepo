package harness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestChainManager(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("Find Hourglass Root", func(t *testing.T) {
		// Test the findHourglassRoot function
		root := findHourglassRoot()
		
		// Verify we found a valid directory
		assert.NotEmpty(t, root)
		
		// Check if it looks like the hourglass monorepo
		if root != "." {
			// Should contain expected directories
			_, err := os.Stat(filepath.Join(root, "hgctl-go"))
			assert.NoError(t, err, "Should find hgctl-go directory")
		}
	})

	t.Run("Create ChainManager", func(t *testing.T) {
		manager := NewChainManager(logger)
		require.NotNil(t, manager)
		assert.NotEmpty(t, manager.projectRoot)
	})

	t.Run("Load Chain Config", func(t *testing.T) {
		manager := NewChainManager(logger)
		
		// Try to load config - it might not exist yet
		config, err := manager.LoadChainConfig()
		
		if err != nil {
			// Config file doesn't exist, which is OK for this test
			t.Logf("Chain config not found (expected): %v", err)
			assert.Contains(t, err.Error(), "chain-config.json")
		} else {
			// If config exists, verify it has expected fields
			assert.NotNil(t, config)
			if config.AVSTaskRegistrarAddress != "" {
				assert.True(t, len(config.AVSTaskRegistrarAddress) > 0)
			}
		}
	})

	// Note: We don't test StartChains here because it requires:
	// 1. The generateTestChainState.sh script to exist
	// 2. Anvil to be installed
	// 3. Network ports to be available
	// These are tested in the integration tests
}