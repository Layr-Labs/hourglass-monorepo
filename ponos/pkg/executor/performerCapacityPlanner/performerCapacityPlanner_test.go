package performerCapacityPlanner

import (
	"context"
	"sync"
)

// MockPerformerCapacityPlanner implements IPerformerCapacityPlanner for testing
type MockPerformerCapacityPlanner struct {
	mutex sync.RWMutex

	// Maps AVS addresses to capacity plans
	CapacityPlans map[string]*PerformerCapacityPlan

	// For controlling return values in tests
	GetCapacityPlanFn func(avsAddress string) (*PerformerCapacityPlan, error)

	// Lifecycle tracking
	Started bool
}

// NewMockPerformerCapacityPlanner creates a new mock capacity planner
func NewMockPerformerCapacityPlanner() *MockPerformerCapacityPlanner {
	return &MockPerformerCapacityPlanner{
		mutex:             sync.RWMutex{},
		CapacityPlans:     make(map[string]*PerformerCapacityPlan),
		GetCapacityPlanFn: nil,
		Started:           false,
	}
}

// GetCapacityPlan implements the IPerformerCapacityPlanner interface for testing
func (m *MockPerformerCapacityPlanner) GetCapacityPlan(avsAddress string) (*PerformerCapacityPlan, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Use custom function if provided
	if m.GetCapacityPlanFn != nil {
		return m.GetCapacityPlanFn(avsAddress)
	}

	// Return plan from map if exists
	if plan, exists := m.CapacityPlans[avsAddress]; exists {
		return plan, nil
	}

	// Return a default empty plan
	return &PerformerCapacityPlan{
		TargetCount: 0,
		Artifact: &ArtifactVersion{
			AvsAddress:    avsAddress,
			OperatorSetId: "0",
			Digest:        "",
			RegistryUrl:   "",
		},
	}, nil
}

// Start implements the IPerformerCapacityPlanner interface for testing
func (m *MockPerformerCapacityPlanner) Start(ctx context.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Started = true
}

// AddCapacityPlan is a helper method for tests to add a capacity plan
func (m *MockPerformerCapacityPlanner) AddCapacityPlan(avsAddress string, plan *PerformerCapacityPlan) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CapacityPlans[avsAddress] = plan
}
