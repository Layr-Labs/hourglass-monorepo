package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
)

// InMemoryExecutorStore implements ExecutorStore interface with in-memory storage
type InMemoryExecutorStore struct {
	mu              sync.RWMutex
	closed          bool
	performerStates map[string]*storage.PerformerState
	inflightTasks   map[string]*storage.TaskInfo
	deployments     map[string]*storage.DeploymentInfo
}

// NewInMemoryExecutorStore creates a new in-memory executor store
func NewInMemoryExecutorStore() *InMemoryExecutorStore {
	return &InMemoryExecutorStore{
		performerStates: make(map[string]*storage.PerformerState),
		inflightTasks:   make(map[string]*storage.TaskInfo),
		deployments:     make(map[string]*storage.DeploymentInfo),
	}
}

// SavePerformerState saves the state of a performer
func (s *InMemoryExecutorStore) SavePerformerState(ctx context.Context, performerId string, state *storage.PerformerState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	if state == nil {
		return fmt.Errorf("performer state cannot be nil")
	}

	if performerId == "" {
		return fmt.Errorf("performer ID cannot be empty")
	}

	// Clone the state to avoid external modifications
	stateCopy := *state
	stateCopy.PerformerId = performerId // Ensure consistency
	s.performerStates[performerId] = &stateCopy

	return nil
}

// GetPerformerState retrieves the state of a performer
func (s *InMemoryExecutorStore) GetPerformerState(ctx context.Context, performerId string) (*storage.PerformerState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	state, exists := s.performerStates[performerId]
	if !exists {
		return nil, storage.ErrNotFound
	}

	// Return a copy to prevent external modifications
	stateCopy := *state
	return &stateCopy, nil
}

// ListPerformerStates returns all performer states
func (s *InMemoryExecutorStore) ListPerformerStates(ctx context.Context) ([]*storage.PerformerState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	states := make([]*storage.PerformerState, 0, len(s.performerStates))
	for _, state := range s.performerStates {
		// Return copies to prevent external modifications
		stateCopy := *state
		states = append(states, &stateCopy)
	}

	return states, nil
}

// DeletePerformerState removes a performer state
func (s *InMemoryExecutorStore) DeletePerformerState(ctx context.Context, performerId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	if _, exists := s.performerStates[performerId]; !exists {
		return storage.ErrNotFound
	}

	delete(s.performerStates, performerId)
	return nil
}

// Close closes the store
func (s *InMemoryExecutorStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	s.closed = true

	// Clear all maps
	s.performerStates = nil
	s.inflightTasks = nil
	s.deployments = nil

	return nil
}

// validateDeploymentStatusTransition validates deployment status transitions
func validateDeploymentStatusTransition(from, to storage.DeploymentStatus) error {
	// Define valid transitions
	validTransitions := map[storage.DeploymentStatus][]storage.DeploymentStatus{
		storage.DeploymentStatusPending:   {storage.DeploymentStatusDeploying, storage.DeploymentStatusFailed},
		storage.DeploymentStatusDeploying: {storage.DeploymentStatusRunning, storage.DeploymentStatusFailed},
		storage.DeploymentStatusRunning:   {storage.DeploymentStatusFailed}, // Can fail after running
		storage.DeploymentStatusFailed:    {},                               // Terminal state
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("%w: unknown status %s", storage.ErrInvalidDeploymentStatus, from)
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return nil
		}
	}

	return fmt.Errorf("%w: cannot transition from %s to %s", storage.ErrInvalidDeploymentStatus, from, to)
}
