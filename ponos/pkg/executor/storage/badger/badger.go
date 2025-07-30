package badger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	badgerv3 "github.com/dgraph-io/badger/v3"
)

// Key prefixes for different data types
const (
	prefixPerformer      = "performer:%s"
	prefixTask           = "task:%s"
	prefixDeployment     = "deployment:%s"
	prefixDeployByStatus = "deploystatus:%s:%s" // status:deploymentId
)

// BadgerExecutorStore implements the ExecutorStore interface using BadgerDB
type BadgerExecutorStore struct {
	db       *badgerv3.DB
	mu       sync.RWMutex
	closed   bool
	closeCh  chan struct{}
	gcTicker *time.Ticker
}

// NewBadgerExecutorStore creates a new BadgerDB-backed executor store
func NewBadgerExecutorStore(cfg *executorConfig.BadgerConfig) (*BadgerExecutorStore, error) {
	if cfg == nil {
		return nil, errors.New("badger config is nil")
	}

	opts := badgerv3.DefaultOptions(cfg.Dir)
	opts.Logger = nil // Disable BadgerDB's default logging

	// Apply custom options if needed
	if cfg.InMemory {
		opts.InMemory = true
	}
	if cfg.ValueLogFileSize > 0 {
		opts.ValueLogFileSize = int64(cfg.ValueLogFileSize)
	}
	if cfg.NumVersionsToKeep > 0 {
		opts.NumVersionsToKeep = cfg.NumVersionsToKeep
	}
	if cfg.NumLevelZeroTables > 0 {
		opts.NumLevelZeroTables = cfg.NumLevelZeroTables
	}
	if cfg.NumLevelZeroTablesStall > 0 {
		opts.NumLevelZeroTablesStall = cfg.NumLevelZeroTablesStall
	}

	db, err := badgerv3.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	s := &BadgerExecutorStore{
		db:      db,
		closeCh: make(chan struct{}),
	}

	// Start garbage collection routine
	s.gcTicker = time.NewTicker(5 * time.Minute)
	go s.runGC()

	return s, nil
}

// runGC runs periodic garbage collection
func (s *BadgerExecutorStore) runGC() {
	for {
		select {
		case <-s.gcTicker.C:
			s.mu.RLock()
			if s.closed {
				s.mu.RUnlock()
				return
			}
			s.mu.RUnlock()

			// Run value log GC
			_ = s.db.RunValueLogGC(0.5)
		case <-s.closeCh:
			return
		}
	}
}

// SavePerformerState stores a performer state
func (s *BadgerExecutorStore) SavePerformerState(ctx context.Context, performerId string, state *storage.PerformerState) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	if state == nil {
		return errors.New("state is nil")
	}

	key := fmt.Sprintf(prefixPerformer, performerId)
	value, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal performer state: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		return txn.Set([]byte(key), value)
	})

	if err != nil {
		return fmt.Errorf("failed to save performer state: %w", err)
	}

	return nil
}

// GetPerformerState retrieves a performer state
func (s *BadgerExecutorStore) GetPerformerState(ctx context.Context, performerId string) (*storage.PerformerState, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var state storage.PerformerState
	key := fmt.Sprintf(prefixPerformer, performerId)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &state)
		})
	})

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get performer state: %w", err)
	}

	return &state, nil
}

// ListPerformerStates returns all performer states
func (s *BadgerExecutorStore) ListPerformerStates(ctx context.Context) ([]*storage.PerformerState, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var states []*storage.PerformerState
	prefix := []byte("performer:")

	err := s.db.View(func(txn *badgerv3.Txn) error {
		opts := badgerv3.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			var state storage.PerformerState
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &state)
			})
			if err != nil {
				continue // Skip on unmarshal error
			}

			states = append(states, &state)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list performer states: %w", err)
	}

	return states, nil
}

// DeletePerformerState removes a performer state
func (s *BadgerExecutorStore) DeletePerformerState(ctx context.Context, performerId string) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	key := fmt.Sprintf(prefixPerformer, performerId)

	err := s.db.Update(func(txn *badgerv3.Txn) error {
		// Check if exists
		_, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return txn.Delete([]byte(key))
	})

	if err != nil {
		return fmt.Errorf("failed to delete performer state: %w", err)
	}

	return nil
}

// SaveInflightTask stores an inflight task
func (s *BadgerExecutorStore) SaveInflightTask(ctx context.Context, taskId string, task *storage.TaskInfo) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	if task == nil {
		return errors.New("task is nil")
	}

	key := fmt.Sprintf(prefixTask, taskId)
	value, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		return txn.Set([]byte(key), value)
	})

	if err != nil {
		return fmt.Errorf("failed to save inflight task: %w", err)
	}

	return nil
}

// GetInflightTask retrieves an inflight task
func (s *BadgerExecutorStore) GetInflightTask(ctx context.Context, taskId string) (*storage.TaskInfo, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var task storage.TaskInfo
	key := fmt.Sprintf(prefixTask, taskId)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &task)
		})
	})

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get inflight task: %w", err)
	}

	return &task, nil
}

// ListInflightTasks returns all inflight tasks
func (s *BadgerExecutorStore) ListInflightTasks(ctx context.Context) ([]*storage.TaskInfo, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var tasks []*storage.TaskInfo
	prefix := []byte("task:")

	err := s.db.View(func(txn *badgerv3.Txn) error {
		opts := badgerv3.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			var task storage.TaskInfo
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &task)
			})
			if err != nil {
				continue // Skip on unmarshal error
			}

			tasks = append(tasks, &task)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list inflight tasks: %w", err)
	}

	return tasks, nil
}

// DeleteInflightTask removes an inflight task
func (s *BadgerExecutorStore) DeleteInflightTask(ctx context.Context, taskId string) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	key := fmt.Sprintf(prefixTask, taskId)

	err := s.db.Update(func(txn *badgerv3.Txn) error {
		// Check if exists
		_, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return txn.Delete([]byte(key))
	})

	if err != nil {
		return fmt.Errorf("failed to delete inflight task: %w", err)
	}

	return nil
}

// SaveDeployment stores a deployment
func (s *BadgerExecutorStore) SaveDeployment(ctx context.Context, deploymentId string, deployment *storage.DeploymentInfo) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	if deployment == nil {
		return errors.New("deployment is nil")
	}

	key := fmt.Sprintf(prefixDeployment, deploymentId)
	value, err := json.Marshal(deployment)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		// Save deployment
		if err := txn.Set([]byte(key), value); err != nil {
			return err
		}

		// Add to status index if it's a new deployment
		if deployment.Status != "" {
			statusKey := fmt.Sprintf(prefixDeployByStatus, deployment.Status, deploymentId)
			return txn.Set([]byte(statusKey), []byte{})
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to save deployment: %w", err)
	}

	return nil
}

// GetDeployment retrieves a deployment
func (s *BadgerExecutorStore) GetDeployment(ctx context.Context, deploymentId string) (*storage.DeploymentInfo, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var deployment storage.DeploymentInfo
	key := fmt.Sprintf(prefixDeployment, deploymentId)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &deployment)
		})
	})

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return &deployment, nil
}

// UpdateDeploymentStatus updates the status of a deployment
func (s *BadgerExecutorStore) UpdateDeploymentStatus(ctx context.Context, deploymentId string, status storage.DeploymentStatus) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	return s.db.Update(func(txn *badgerv3.Txn) error {
		// Get current deployment
		key := fmt.Sprintf(prefixDeployment, deploymentId)
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		var deployment storage.DeploymentInfo
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &deployment)
		})
		if err != nil {
			return err
		}

		// Validate status transition
		if !isValidDeploymentStatusTransition(deployment.Status, status) {
			return fmt.Errorf("invalid status transition from %s to %s", deployment.Status, status)
		}

		// Remove old status index
		if deployment.Status != "" {
			oldStatusKey := fmt.Sprintf(prefixDeployByStatus, deployment.Status, deploymentId)
			if err := txn.Delete([]byte(oldStatusKey)); err != nil {
				return err
			}
		}

		// Update status
		deployment.Status = status
		now := time.Now()
		deployment.CompletedAt = &now
		value, err := json.Marshal(deployment)
		if err != nil {
			return err
		}

		// Save updated deployment
		if err := txn.Set([]byte(key), value); err != nil {
			return err
		}

		// Add new status index
		newStatusKey := fmt.Sprintf(prefixDeployByStatus, status, deploymentId)
		return txn.Set([]byte(newStatusKey), []byte{})
	})
}

// Close shuts down the store
func (s *BadgerExecutorStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	close(s.closeCh)
	s.gcTicker.Stop()

	return s.db.Close()
}

// isValidDeploymentStatusTransition checks if a deployment status transition is valid
func isValidDeploymentStatusTransition(from, to storage.DeploymentStatus) bool {
	switch from {
	case storage.DeploymentStatusPending:
		return to == storage.DeploymentStatusDeploying || to == storage.DeploymentStatusFailed
	case storage.DeploymentStatusDeploying:
		return to == storage.DeploymentStatusRunning || to == storage.DeploymentStatusFailed
	case storage.DeploymentStatusRunning:
		return to == storage.DeploymentStatusFailed // Can fail from running
	case storage.DeploymentStatusFailed:
		return false // Terminal state
	default:
		return false
	}
}
