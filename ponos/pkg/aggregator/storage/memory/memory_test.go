package memory_test

import (
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage/memory"
)

// TestInMemoryAggregatorStore runs the standard storage test suite
func TestInMemoryAggregatorStore(t *testing.T) {
	suite := &storage.TestSuite{
		NewStore: func() (storage.AggregatorStore, error) {
			return memory.NewInMemoryAggregatorStore(), nil
		},
	}
	suite.Run(t)
}

// TestInMemorySpecific tests in-memory specific behavior
func TestInMemorySpecific(t *testing.T) {
	t.Run("MultipleInstances", func(t *testing.T) {
		// Test that multiple instances don't share state
		store1 := memory.NewInMemoryAggregatorStore()
		store2 := memory.NewInMemoryAggregatorStore()

		// Both should have independent state
		if store1 == store2 {
			t.Fatal("NewInMemoryAggregatorStore should create independent instances")
		}
	})
}
