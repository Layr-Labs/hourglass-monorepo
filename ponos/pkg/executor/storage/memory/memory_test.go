package memory_test

import (
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
)

// TestInMemoryExecutorStore runs the standard storage test suite
func TestInMemoryExecutorStore(t *testing.T) {
	suite := &storage.TestSuite{
		NewStore: func() (storage.ExecutorStore, error) {
			return memory.NewInMemoryExecutorStore(), nil
		},
	}
	suite.Run(t)
}

// TestInMemorySpecific tests in-memory specific behavior
func TestInMemorySpecific(t *testing.T) {
	t.Run("MultipleInstances", func(t *testing.T) {
		// Test that multiple instances don't share state
		store1 := memory.NewInMemoryExecutorStore()
		store2 := memory.NewInMemoryExecutorStore()

		// Both should have independent state
		if store1 == store2 {
			t.Fatal("NewInMemoryExecutorStore should create independent instances")
		}
	})
}
