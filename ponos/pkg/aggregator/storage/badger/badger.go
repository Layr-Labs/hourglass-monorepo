package badger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	prefixTask         = "task:%s"
	prefixTaskByStatus = "taskstatus:%s:%s" // status:taskId
	prefixBlock        = "block:%s:%d:%d"   // avsAddress:chainId:blockNumber
	prefixLatestBlock  = "block:%s:%d"      // avsAddress:chainId
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
		opts.ValueLogFileSize = cfg.ValueLogFileSize
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

// GetLastProcessedBlock retrieves the last processed block record for a chain for a specific AVS
func (s *BadgerAggregatorStore) GetLastProcessedBlock(ctx context.Context, avsAddress string, chainId config.ChainId) (*storage.BlockRecord, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var blockRecord *storage.BlockRecord
	prefix := fmt.Sprintf(prefixLatestBlock, avsAddress, chainId)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		opts := badgerv3.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		opts.Reverse = true

		it := txn.NewIterator(opts)
		defer it.Close()

		it.Rewind()
		if it.Valid() {
			item := it.Item()

			var record storage.BlockRecord
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &record)
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal block record: %w", err)
			}

			blockRecord = &record
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get last processed block: %w", err)
	}

	if blockRecord == nil {
		return nil, storage.ErrNotFound
	}

	return blockRecord, nil
}

func (s *BadgerAggregatorStore) SavePendingTask(ctx context.Context, task *types.Task) error {
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

// ListPendingTasksForAVS returns all pending tasks for a specific AVS address
func (s *BadgerAggregatorStore) ListPendingTasksForAVS(ctx context.Context, avsAddress string) ([]*types.Task, error) {
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
				continue
			}

			// Filter by AVS address
			if strings.EqualFold(record.Task.AVSAddress, avsAddress) {
				tasks = append(tasks, record.Task)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list pending tasks for AVS: %w", err)
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

// SaveBlock saves block information for reorg detection
func (s *BadgerAggregatorStore) SaveBlock(ctx context.Context, avsAddress string, block *storage.BlockRecord) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	if block == nil {
		return errors.New("block is nil")
	}

	key := fmt.Sprintf(prefixBlock, avsAddress, block.ChainId, block.Number)
	value, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block info: %w", err)
	}

	err = s.db.Update(func(txn *badgerv3.Txn) error {
		return txn.Set([]byte(key), value)
	})

	if err != nil {
		return fmt.Errorf("failed to save block info: %w", err)
	}

	return nil
}

// GetBlock retrieves block information by block number
func (s *BadgerAggregatorStore) GetBlock(ctx context.Context, avsAddress string, chainId config.ChainId, blockNumber uint64) (*storage.BlockRecord, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	var block storage.BlockRecord
	key := fmt.Sprintf(prefixBlock, avsAddress, chainId, blockNumber)

	err := s.db.View(func(txn *badgerv3.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badgerv3.ErrKeyNotFound) {
				return storage.ErrNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &block)
		})
	})

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get block info: %w", err)
	}

	return &block, nil
}

// DeleteBlock removes block information from storage
func (s *BadgerAggregatorStore) DeleteBlock(ctx context.Context, avsAddress string, chainId config.ChainId, blockNumber uint64) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return storage.ErrStoreClosed
	}
	s.mu.RUnlock()

	key := fmt.Sprintf(prefixBlock, avsAddress, chainId, blockNumber)

	err := s.db.Update(func(txn *badgerv3.Txn) error {
		// Check if key exists first
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
		if errors.Is(err, storage.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete block info: %w", err)
	}

	return nil
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
