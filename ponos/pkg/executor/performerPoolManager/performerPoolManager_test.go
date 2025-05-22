package performerPoolManager

import (
	"context"
	"sync"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
)

// MockPerformerPoolManager implements IPerformerPoolManager for testing
type MockPerformerPoolManager struct {
	mutex sync.RWMutex

	// State tracking
	Initialized bool
	Started     bool
	Pools       map[string]IPerformerPool

	// For controlling return values in tests
	InitializeFn   func() error
	StartFn        func(ctx context.Context) error
	GetPerformerFn func(avsAddress string) (avsPerformer.IAvsPerformer, bool)
}

// NewMockPerformerPoolManager creates a new mock performer pool manager
func NewMockPerformerPoolManager() *MockPerformerPoolManager {
	return &MockPerformerPoolManager{
		mutex:       sync.RWMutex{},
		Initialized: false,
		Started:     false,
		Pools:       make(map[string]IPerformerPool),
	}
}

// Initialize implements the IPerformerPoolManager interface for testing
func (m *MockPerformerPoolManager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Use custom function if provided
	if m.InitializeFn != nil {
		return m.InitializeFn()
	}

	// Default implementation marks as initialized
	m.Initialized = true
	return nil
}

// Start implements the IPerformerPoolManager interface for testing
func (m *MockPerformerPoolManager) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Use custom function if provided
	if m.StartFn != nil {
		return m.StartFn(ctx)
	}

	// Default implementation marks as started
	m.Started = true
	return nil
}

// GetPerformer implements the IPerformerPoolManager interface for testing
func (m *MockPerformerPoolManager) GetPerformer(avsAddress string) (avsPerformer.IAvsPerformer, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Use custom function if provided
	if m.GetPerformerFn != nil {
		return m.GetPerformerFn(avsAddress)
	}

	// Check if we have a pool for this AVS
	pool, exists := m.Pools[avsAddress]
	if !exists {
		return nil, false
	}

	// Return performer from the pool
	return pool.GetHealthyPerformer()
}

// AddMockPool is a helper method for tests to add a mock pool for an AVS
func (m *MockPerformerPoolManager) AddMockPool(avsAddress string, pool IPerformerPool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Pools[avsAddress] = pool
}
