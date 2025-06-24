package avsExecutionManager

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateSeedFromMixHashAndAVS(t *testing.T) {
	tests := []struct {
		name       string
		mixHash    string
		avsAddress string
		expected   string // Expected seed as hex string for comparison
	}{
		{
			name:       "with 0x prefix",
			mixHash:    "0x1234567890abcdef",
			avsAddress: "0xabcdef1234567890",
			expected:   "TBD", // Will need to be calculated with keccak256
		},
		{
			name:       "without 0x prefix",
			mixHash:    "1234567890abcdef",
			avsAddress: "abcdef1234567890",
			expected:   "TBD", // Will need to be calculated with keccak256
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := calculateSeedFromMixHashAndAVS(tt.mixHash, tt.avsAddress)
			// Just verify it returns a valid big.Int for now
			assert.NotNil(t, seed)
			assert.True(t, seed.Cmp(big.NewInt(0)) >= 0)
		})
	}
}

func TestSelectStakeWeightedLeader(t *testing.T) {
	t.Run("single operator", func(t *testing.T) {
		operator1 := "0x1111111111111111111111111111111111111111"
		aggregators := &operatorManager.PeerWeight{
			Operators: []*peering.OperatorPeerInfo{
				{OperatorAddress: operator1},
			},
			Weights: map[string][]*big.Int{
				operator1: {big.NewInt(100)},
			},
		}

		leader, err := selectStakeWeightedLeader(aggregators, "0x1234", "0xavs123")
		require.NoError(t, err)
		assert.Equal(t, operator1, leader)
	})

	t.Run("multiple operators with equal weights", func(t *testing.T) {
		operator1 := "0x1111111111111111111111111111111111111111"
		operator2 := "0x2222222222222222222222222222222222222222"
		operator3 := "0x3333333333333333333333333333333333333333"

		aggregators := &operatorManager.PeerWeight{
			Operators: []*peering.OperatorPeerInfo{
				{OperatorAddress: operator1},
				{OperatorAddress: operator2},
				{OperatorAddress: operator3},
			},
			Weights: map[string][]*big.Int{
				operator1: {big.NewInt(100)},
				operator2: {big.NewInt(100)},
				operator3: {big.NewInt(100)},
			},
		}

		// Test with different mix hashes to ensure deterministic but varied results
		leader1, err := selectStakeWeightedLeader(aggregators, "0x0000", "0xavs123")
		require.NoError(t, err)

		leader2, err := selectStakeWeightedLeader(aggregators, "0xffff", "0xavs123")
		require.NoError(t, err)

		// Same mix hash should always return same leader
		leader3, err := selectStakeWeightedLeader(aggregators, "0x0000", "0xavs123")
		require.NoError(t, err)
		assert.Equal(t, leader1, leader3)

		// All leaders should be valid operators
		validOperators := []string{operator1, operator2, operator3}
		assert.Contains(t, validOperators, leader1)
		assert.Contains(t, validOperators, leader2)
	})

	t.Run("weighted selection favors higher stake", func(t *testing.T) {
		operator1 := "0x1111111111111111111111111111111111111111" // Low stake
		operator2 := "0x2222222222222222222222222222222222222222" // High stake

		aggregators := &operatorManager.PeerWeight{
			Operators: []*peering.OperatorPeerInfo{
				{OperatorAddress: operator1},
				{OperatorAddress: operator2},
			},
			Weights: map[string][]*big.Int{
				operator1: {big.NewInt(1)},  // 1% of total
				operator2: {big.NewInt(99)}, // 99% of total
			},
		}

		// Test multiple mix hashes and count selections
		operator2Count := 0
		testRuns := 100

		for i := 0; i < testRuns; i++ {
			mixHash := big.NewInt(int64(i)).Text(16)
			leader, err := selectStakeWeightedLeader(aggregators, mixHash, "0xavs123")
			require.NoError(t, err)
			if leader == operator2 {
				operator2Count++
			}
		}

		// operator2 should be selected most of the time (at least 80% due to 99x weight)
		assert.Greater(t, operator2Count, 80, "High stake operator should be selected more frequently")
	})

	t.Run("edge case: zero weights", func(t *testing.T) {
		operator1 := "0x1111111111111111111111111111111111111111"
		aggregators := &operatorManager.PeerWeight{
			Operators: []*peering.OperatorPeerInfo{
				{OperatorAddress: operator1},
			},
			Weights: map[string][]*big.Int{
				operator1: {big.NewInt(0)},
			},
		}

		_, err := selectStakeWeightedLeader(aggregators, "0x1234", "0xavs123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no operators with stake weight found")
	})

	t.Run("edge case: no operators", func(t *testing.T) {
		aggregators := &operatorManager.PeerWeight{
			Operators: []*peering.OperatorPeerInfo{},
			Weights:   map[string][]*big.Int{},
		}

		_, err := selectStakeWeightedLeader(aggregators, "0x1234", "0xavs123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no operators provided")
	})

	t.Run("edge case: nil aggregators", func(t *testing.T) {
		_, err := selectStakeWeightedLeader(nil, "0x1234", "0xavs123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no operators provided")
	})

	t.Run("operators without weights", func(t *testing.T) {
		operator1 := "0x1111111111111111111111111111111111111111"
		operator2 := "0x2222222222222222222222222222222222222222"

		aggregators := &operatorManager.PeerWeight{
			Operators: []*peering.OperatorPeerInfo{
				{OperatorAddress: operator1},
				{OperatorAddress: operator2},
			},
			Weights: map[string][]*big.Int{
				operator1: {big.NewInt(100)}, // Only operator1 has weight
				// operator2 has no weight entry
			},
		}

		leader, err := selectStakeWeightedLeader(aggregators, "0x1234", "0xavs123")
		require.NoError(t, err)
		assert.Equal(t, operator1, leader) // Should select the only operator with weight
	})

	t.Run("operators with empty weight arrays", func(t *testing.T) {
		operator1 := "0x1111111111111111111111111111111111111111"
		operator2 := "0x2222222222222222222222222222222222222222"

		aggregators := &operatorManager.PeerWeight{
			Operators: []*peering.OperatorPeerInfo{
				{OperatorAddress: operator1},
				{OperatorAddress: operator2},
			},
			Weights: map[string][]*big.Int{
				operator1: {big.NewInt(100)},
				operator2: {}, // Empty weight array
			},
		}

		leader, err := selectStakeWeightedLeader(aggregators, "0x1234", "0xavs123")
		require.NoError(t, err)
		assert.Equal(t, operator1, leader) // Should select the operator with valid weight
	})
}

func TestSelectStakeWeightedLeaderDeterministic(t *testing.T) {
	operator1 := "0x1111111111111111111111111111111111111111"
	operator2 := "0x2222222222222222222222222222222222222222"
	operator3 := "0x3333333333333333333333333333333333333333"

	aggregators := &operatorManager.PeerWeight{
		Operators: []*peering.OperatorPeerInfo{
			{OperatorAddress: operator1},
			{OperatorAddress: operator2},
			{OperatorAddress: operator3},
		},
		Weights: map[string][]*big.Int{
			operator1: {big.NewInt(30)},
			operator2: {big.NewInt(50)},
			operator3: {big.NewInt(20)},
		},
	}

	// Test that same mixHash always produces same result
	testCases := []string{
		"0x0000000000000000000000000000000000000000000000000000000000000000",
		"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	}

	for _, mixHash := range testCases {
		t.Run("mixHash_"+mixHash, func(t *testing.T) {
			// Run multiple times with same mixHash
			var leaders []string
			for i := 0; i < 10; i++ {
				leader, err := selectStakeWeightedLeader(aggregators, mixHash, "0xavs123")
				require.NoError(t, err)
				leaders = append(leaders, leader)
			}

			// All results should be identical
			for i := 1; i < len(leaders); i++ {
				assert.Equal(t, leaders[0], leaders[i], "Same mixHash should always produce same leader")
			}
		})
	}
}

func TestSelectStakeWeightedLeaderDistribution(t *testing.T) {
	// This test validates that the distribution roughly matches the stake weights
	operator1 := "0x1111111111111111111111111111111111111111" // 10% stake
	operator2 := "0x2222222222222222222222222222222222222222" // 60% stake
	operator3 := "0x3333333333333333333333333333333333333333" // 30% stake

	aggregators := &operatorManager.PeerWeight{
		Operators: []*peering.OperatorPeerInfo{
			{OperatorAddress: operator1},
			{OperatorAddress: operator2},
			{OperatorAddress: operator3},
		},
		Weights: map[string][]*big.Int{
			operator1: {big.NewInt(10)},
			operator2: {big.NewInt(60)},
			operator3: {big.NewInt(30)},
		},
	}

	counts := make(map[string]int)
	totalRuns := 1000

	// Run many selections with different seeds
	for i := 0; i < totalRuns; i++ {
		// Create more varied hex strings by using a longer format
		mixHash := fmt.Sprintf("%064x", i) // 64 character hex string
		leader, err := selectStakeWeightedLeader(aggregators, mixHash, "0xavs123")
		require.NoError(t, err)
		counts[leader]++
	}

	// Check that distribution roughly matches stake weights (within reasonable margin)
	tolerance := 0.05 // 5% tolerance

	expectedPercent1 := 0.10
	actualPercent1 := float64(counts[operator1]) / float64(totalRuns)
	assert.InDelta(t, expectedPercent1, actualPercent1, tolerance, "Operator1 selection frequency should match stake weight")

	expectedPercent2 := 0.60
	actualPercent2 := float64(counts[operator2]) / float64(totalRuns)
	assert.InDelta(t, expectedPercent2, actualPercent2, tolerance, "Operator2 selection frequency should match stake weight")

	expectedPercent3 := 0.30
	actualPercent3 := float64(counts[operator3]) / float64(totalRuns)
	assert.InDelta(t, expectedPercent3, actualPercent3, tolerance, "Operator3 selection frequency should match stake weight")

	// Ensure all operators were selected at least once
	assert.Greater(t, counts[operator1], 0, "Operator1 should be selected at least once")
	assert.Greater(t, counts[operator2], 0, "Operator2 should be selected at least once")
	assert.Greater(t, counts[operator3], 0, "Operator3 should be selected at least once")
}
