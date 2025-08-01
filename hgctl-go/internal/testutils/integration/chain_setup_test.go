package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

// TestChainSetup tests the basic chain setup without running full tests
func TestChainSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chain setup test in short mode")
	}

	t.Run("Verify Chains Start", func(t *testing.T) {
		// Get test harness - this will trigger chain setup
		h := getTestHarness(t)
		require.NotNil(t, h)

		// Give chains a moment to stabilize
		time.Sleep(2 * time.Second)

		// Test L1 connection
		l1Client, err := ethclient.Dial("http://localhost:8545")
		if err != nil {
			t.Logf("Failed to connect to L1: %v", err)
			t.Skip("L1 chain not available")
		}
		defer l1Client.Close()

		// Get chain ID
		ctx := context.Background()
		chainID, err := l1Client.ChainID(ctx)
		require.NoError(t, err)
		require.Equal(t, int64(31337), chainID.Int64())

		// Get latest block
		block, err := l1Client.BlockByNumber(ctx, nil)
		require.NoError(t, err)
		t.Logf("L1 latest block: %d", block.NumberU64())

		// Test L2 connection
		l2Client, err := ethclient.Dial("http://localhost:9545")
		if err != nil {
			t.Logf("Failed to connect to L2: %v", err)
			t.Skip("L2 chain not available")
		}
		defer l2Client.Close()

		// Get L2 chain ID
		chainID, err = l2Client.ChainID(ctx)
		require.NoError(t, err)
		require.Equal(t, int64(31338), chainID.Int64())

		// Get L2 latest block
		block, err = l2Client.BlockByNumber(ctx, nil)
		require.NoError(t, err)
		t.Logf("L2 latest block: %d", block.NumberU64())
	})
}