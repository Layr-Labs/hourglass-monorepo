package performerPoolManager

import (
	"context"
	"sync"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerCapacityPlanner"
)

// MockPerformerPool implements IPerformerPool for testing
type MockPerformerPool struct {
	mutex sync.RWMutex

	// Configuration
	AvsAddr string

	// State tracking
	PerformerCount        int
	HealthyPerformerCount int
	Performers            map[string]avsPerformer.IAvsPerformer
	ExecutedPlans         []*performerCapacityPlanner.PerformerCapacityPlan
	HealthChecks          int
	IsShutdown            bool

	// For controlling return values in tests
	ExecutePlanFn         func(ctx context.Context, plan *performerCapacityPlanner.PerformerCapacityPlan) error
	CheckHealthFn         func(ctx context.Context) (int, error)
	ShutdownFn            func() error
	GetHealthyPerformerFn func() (avsPerformer.IAvsPerformer, bool)
	ReturnedPerformer     avsPerformer.IAvsPerformer
}

// NewMockPerformerPool creates a new mock performer pool
func NewMockPerformerPool(avsAddress string) *MockPerformerPool {
	return &MockPerformerPool{
		AvsAddr:               avsAddress,
		PerformerCount:        0,
		HealthyPerformerCount: 0,
		Performers:            make(map[string]avsPerformer.IAvsPerformer),
		ExecutedPlans:         make([]*performerCapacityPlanner.PerformerCapacityPlan, 0),
		HealthChecks:          0,
		IsShutdown:            false,
	}
}

// ExecutePlan implements the IPerformerPool interface for testing
func (m *MockPerformerPool) ExecutePlan(ctx context.Context, plan *performerCapacityPlanner.PerformerCapacityPlan) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Use custom function if provided
	if m.ExecutePlanFn != nil {
		return m.ExecutePlanFn(ctx, plan)
	}

	// Default implementation tracks the executed plan
	m.ExecutedPlans = append(m.ExecutedPlans, plan)

	// Set performer count based on the plan
	m.PerformerCount = plan.TargetCount
	m.HealthyPerformerCount = plan.TargetCount

	return nil
}

// CheckHealth implements the IPerformerPool interface for testing
func (m *MockPerformerPool) CheckHealth(ctx context.Context) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Use custom function if provided
	if m.CheckHealthFn != nil {
		return m.CheckHealthFn(ctx)
	}

	// Default implementation tracks health checks
	m.HealthChecks++

	return m.HealthyPerformerCount, nil
}

// Shutdown implements the IPerformerPool interface for testing
func (m *MockPerformerPool) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Use custom function if provided
	if m.ShutdownFn != nil {
		return m.ShutdownFn()
	}

	// Default implementation marks as shutdown
	m.IsShutdown = true
	m.PerformerCount = 0
	m.HealthyPerformerCount = 0

	return nil
}

// GetHealthyPerformer implements the IPerformerPool interface for testing
func (m *MockPerformerPool) GetHealthyPerformer() (avsPerformer.IAvsPerformer, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Use custom function if provided
	if m.GetHealthyPerformerFn != nil {
		return m.GetHealthyPerformerFn()
	}

	// Return configured performer or nil if none is healthy
	if m.HealthyPerformerCount > 0 && m.ReturnedPerformer != nil {
		return m.ReturnedPerformer, true
	}

	return nil, false
}

// GetPerformerCount implements the IPerformerPool interface for testing
func (m *MockPerformerPool) GetPerformerCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.PerformerCount
}

// GetAvsAddress implements the IPerformerPool interface for testing
func (m *MockPerformerPool) GetAvsAddress() string {
	return m.AvsAddr
}

// SetMockPerformer is a helper method for tests to set the returned performer
func (m *MockPerformerPool) SetMockPerformer(performer avsPerformer.IAvsPerformer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ReturnedPerformer = performer

	// Update counts if setting a performer
	if performer != nil {
		if m.PerformerCount == 0 {
			m.PerformerCount = 1
		}
		if m.HealthyPerformerCount == 0 {
			m.HealthyPerformerCount = 1
		}
	}
}
