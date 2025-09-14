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
	prefixProcessed      = "processed:%s"       // processed tasks
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

// MarkTaskProcessed marks a task as processed
func (s *BadgerExecutorStore) MarkTaskProcessed(ctx context.Context, taskId string) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	if taskId == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	key := fmt.Sprintf(prefixProcessed, taskId)
	processedTask := &storage.ProcessedTask{
		TaskId:      taskId,
		ProcessedAt: time.Now(),
	}

	value, err := json.Marshal(processedTask)
	if err != nil {
		return fmt.Errorf("failed to marshal processed task: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		return txn.Set([]byte(key), value)
	})

	if err != nil {
		return fmt.Errorf("failed to mark task as processed: %w", err)
	}

	return nil
}

// IsTaskProcessed checks if a task has been processed
func (s *BadgerExecutorStore) IsTaskProcessed(ctx context.Context, taskId string) (bool, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return false, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	key := fmt.Sprintf(prefixProcessed, taskId)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})

	if err != nil {
		if errors.Is(err, badgerv3.ErrKeyNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if task is processed: %w", err)
	}

	return true, nil
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
