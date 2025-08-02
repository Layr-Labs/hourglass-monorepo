package badger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/aggregatorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	badgerv3 "github.com/dgraph-io/badger/v3"
)

// Key prefixes for different data types
const (
	prefixChainBlock     = "chain:%d:lastBlock"
	prefixTask           = "task:%s"
	prefixTaskByStatus   = "taskstatus:%s:%s" // status:taskId
	prefixOperatorConfig = "opset:%s:%d"      // avsAddress:operatorSetId
	prefixAVSConfig      = "avs:%s"           // avsAddress
)

// BadgerAggregatorStore implements the AggregatorStore interface using BadgerDB
type BadgerAggregatorStore struct {
	db       *badgerv3.DB
	mu       sync.RWMutex
	closed   bool
	closeCh  chan struct{}
	gcTicker *time.Ticker
}

// NewBadgerAggregatorStore creates a new BadgerDB-backed aggregator store
func NewBadgerAggregatorStore(cfg *aggregatorConfig.BadgerConfig) (*BadgerAggregatorStore, error) {
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

	s := &BadgerAggregatorStore{
		db:      db,
		closeCh: make(chan struct{}),
	}

	// Start garbage collection routine
	s.gcTicker = time.NewTicker(5 * time.Minute)
	go s.runGC()

	return s, nil
}

// runGC runs periodic garbage collection
func (s *BadgerAggregatorStore) runGC() {
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

// GetLastProcessedBlock retrieves the last processed block for a chain
func (s *BadgerAggregatorStore) GetLastProcessedBlock(ctx context.Context, chainId config.ChainId) (uint64, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return 0, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var blockNum uint64
	key := fmt.Sprintf(prefixChainBlock, chainId)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &blockNum)
		})
	})

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return 0, err
		}
		return 0, fmt.Errorf("failed to get last processed block: %w", err)
	}

	return blockNum, nil
}

// SetLastProcessedBlock updates the last processed block for a chain
func (s *BadgerAggregatorStore) SetLastProcessedBlock(ctx context.Context, chainId config.ChainId, blockNum uint64) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	key := fmt.Sprintf(prefixChainBlock, chainId)
	value, err := json.Marshal(blockNum)
	if err != nil {
		return fmt.Errorf("failed to marshal block number: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		return txn.Set([]byte(key), value)
	})

	if err != nil {
		return fmt.Errorf("failed to set last processed block: %w", err)
	}

	return nil
}

// SaveTask stores a task
func (s *BadgerAggregatorStore) SaveTask(ctx context.Context, task *types.Task) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	if task == nil {
		return errors.New("task is nil")
	}

	// Create task record with status
	record := &storage.TaskRecord{
		Task:   task,
		Status: storage.TaskStatusPending,
	}

	value, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		// Check if task already exists
		taskKey := fmt.Sprintf(prefixTask, task.TaskId)
		_, err := txn.Get([]byte(taskKey))
		if err == nil {
			return storage.ErrAlreadyExists
		}
		if !errors.Is(err, badgerv3.ErrKeyNotFound) {
			return err
		}

		// Save task
		if err := txn.Set([]byte(taskKey), value); err != nil {
			return err
		}

		// Add to status index
		statusKey := fmt.Sprintf(prefixTaskByStatus, storage.TaskStatusPending, task.TaskId)
		return txn.Set([]byte(statusKey), []byte{})
	})

	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	return nil
}

// GetTask retrieves a task by ID
func (s *BadgerAggregatorStore) GetTask(ctx context.Context, taskId string) (*types.Task, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var record storage.TaskRecord
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
			return json.Unmarshal(val, &record)
		})
	})

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return record.Task, nil
}

// ListPendingTasks returns all tasks with pending status
func (s *BadgerAggregatorStore) ListPendingTasks(ctx context.Context) ([]*types.Task, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var tasks []*types.Task
	prefix := fmt.Sprintf(prefixTaskByStatus, storage.TaskStatusPending, "")

	err := s.db.View(func(txn *badgerv3.Txn) error {
		opts := badgerv3.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			// Extract task ID from key
			key := string(it.Item().Key())
			taskId := key[len(prefix):]

			// Get the actual task
			taskKey := fmt.Sprintf(prefixTask, taskId)
			item, err := txn.Get([]byte(taskKey))
			if err != nil {
				continue // Skip if task not found
			}

			var record storage.TaskRecord
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &record)
			})
			if err != nil {
				continue // Skip on unmarshal error
			}

			tasks = append(tasks, record.Task)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list pending tasks: %w", err)
	}

	return tasks, nil
}

// UpdateTaskStatus updates the status of a task
func (s *BadgerAggregatorStore) UpdateTaskStatus(ctx context.Context, taskId string, status storage.TaskStatus) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	return s.db.Update(func(txn *badgerv3.Txn) error {
		// Get current task
		taskKey := fmt.Sprintf(prefixTask, taskId)
		item, err := txn.Get([]byte(taskKey))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		var record storage.TaskRecord
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &record)
		})
		if err != nil {
			return err
		}

		// Validate status transition
		if !isValidStatusTransition(record.Status, status) {
			return fmt.Errorf("invalid status transition from %s to %s", record.Status, status)
		}

		// Remove old status index
		oldStatusKey := fmt.Sprintf(prefixTaskByStatus, record.Status, taskId)
		if err := txn.Delete([]byte(oldStatusKey)); err != nil {
			return err
		}

		// Update status
		record.Status = status
		value, err := json.Marshal(record)
		if err != nil {
			return err
		}

		// Save updated task
		if err := txn.Set([]byte(taskKey), value); err != nil {
			return err
		}

		// Add new status index
		newStatusKey := fmt.Sprintf(prefixTaskByStatus, status, taskId)
		return txn.Set([]byte(newStatusKey), []byte{})
	})
}

// DeleteTask removes a task
func (s *BadgerAggregatorStore) DeleteTask(ctx context.Context, taskId string) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	return s.db.Update(func(txn *badgerv3.Txn) error {
		// Get current task to know its status
		taskKey := fmt.Sprintf(prefixTask, taskId)
		item, err := txn.Get([]byte(taskKey))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		var record storage.TaskRecord
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &record)
		})
		if err != nil {
			return err
		}

		// Delete task
		if err := txn.Delete([]byte(taskKey)); err != nil {
			return err
		}

		// Delete status index
		statusKey := fmt.Sprintf(prefixTaskByStatus, record.Status, taskId)
		return txn.Delete([]byte(statusKey))
	})
}

// SaveOperatorSetConfig stores an operator set configuration
func (s *BadgerAggregatorStore) SaveOperatorSetConfig(ctx context.Context, avsAddress string, operatorSetId uint32, config *storage.OperatorSetTaskConfig) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	if config == nil {
		return errors.New("config is nil")
	}

	key := fmt.Sprintf(prefixOperatorConfig, avsAddress, operatorSetId)
	value, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		return txn.Set([]byte(key), value)
	})

	if err != nil {
		return fmt.Errorf("failed to save operator set config: %w", err)
	}

	return nil
}

// GetOperatorSetConfig retrieves an operator set configuration
func (s *BadgerAggregatorStore) GetOperatorSetConfig(ctx context.Context, avsAddress string, operatorSetId uint32) (*storage.OperatorSetTaskConfig, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var config storage.OperatorSetTaskConfig
	key := fmt.Sprintf(prefixOperatorConfig, avsAddress, operatorSetId)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get operator set config: %w", err)
	}

	return &config, nil
}

// SaveAVSConfig stores an AVS configuration
func (s *BadgerAggregatorStore) SaveAVSConfig(ctx context.Context, avsAddress string, config *storage.AvsConfig) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	if config == nil {
		return errors.New("config is nil")
	}

	key := fmt.Sprintf(prefixAVSConfig, avsAddress)
	value, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		return txn.Set([]byte(key), value)
	})

	if err != nil {
		return fmt.Errorf("failed to save avs config: %w", err)
	}

	return nil
}

// GetAVSConfig retrieves an AVS configuration
func (s *BadgerAggregatorStore) GetAVSConfig(ctx context.Context, avsAddress string) (*storage.AvsConfig, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var config storage.AvsConfig
	key := fmt.Sprintf(prefixAVSConfig, avsAddress)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get avs config: %w", err)
	}

	return &config, nil
}

// Close shuts down the store
func (s *BadgerAggregatorStore) Close() error {
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

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(from, to storage.TaskStatus) bool {
	switch from {
	case storage.TaskStatusPending:
		return to == storage.TaskStatusProcessing || to == storage.TaskStatusFailed
	case storage.TaskStatusProcessing:
		return to == storage.TaskStatusCompleted || to == storage.TaskStatusFailed
	case storage.TaskStatusCompleted, storage.TaskStatusFailed:
		return false // Terminal states
	default:
		return false
	}
}
