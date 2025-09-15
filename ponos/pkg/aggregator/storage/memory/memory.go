package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

// InMemoryAggregatorStore implements AggregatorStore interface with in-memory storage
type InMemoryAggregatorStore struct {
	mu                  sync.RWMutex
	closed              bool
	lastProcessedBlocks map[string]uint64
	tasks               map[string]*storage.TaskRecord
	operatorSetConfigs  map[string]*storage.OperatorSetTaskConfig
	avsConfigs          map[string]*storage.AvsConfig
	blocks              map[string]*storage.BlockRecord
}

// NewInMemoryAggregatorStore creates a new in-memory aggregator store
func NewInMemoryAggregatorStore() *InMemoryAggregatorStore {
	return &InMemoryAggregatorStore{
		lastProcessedBlocks: make(map[string]uint64),
		tasks:               make(map[string]*storage.TaskRecord),
		operatorSetConfigs:  make(map[string]*storage.OperatorSetTaskConfig),
		avsConfigs:          make(map[string]*storage.AvsConfig),
		blocks:              make(map[string]*storage.BlockRecord),
	}
}

// makeBlockKey creates a composite key for last processed block storage
func makeBlockKey(avsAddress string, chainId config.ChainId) string {
	return fmt.Sprintf("%s:%d", avsAddress, chainId)
}

// makeBlockRecordKey creates a composite key for block info storage
func makeBlockRecordKey(avsAddress string, chainId config.ChainId, blockNumber uint64) string {
	return fmt.Sprintf("block:%s:%d:%d", avsAddress, chainId, blockNumber)
}

// GetLastProcessedBlock returns the last processed block for a chain for a specific AVS
func (s *InMemoryAggregatorStore) GetLastProcessedBlock(ctx context.Context, avsAddress string, chainId config.ChainId) (*storage.BlockRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	key := makeBlockKey(avsAddress, chainId)
	blockNum, exists := s.lastProcessedBlocks[key]
	if !exists {
		return nil, storage.ErrNotFound
	}

	blockKey := makeBlockRecordKey(avsAddress, chainId, blockNum)
	block, exists := s.blocks[blockKey]
	if !exists {
		return nil, storage.ErrNotFound
	}

	return block, nil
}

// SavePendingTask SaveTask saves a task to storage
func (s *InMemoryAggregatorStore) SavePendingTask(ctx context.Context, task *types.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	if task == nil || task.TaskId == "" {
		return fmt.Errorf("invalid task: task or taskId is empty")
	}

	// Create task record with pending status
	record := &storage.TaskRecord{
		Task:      task,
		Status:    storage.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.tasks[task.TaskId] = record
	return nil
}

// GetTask retrieves a task by ID
func (s *InMemoryAggregatorStore) GetTask(ctx context.Context, taskId string) (*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	record, exists := s.tasks[taskId]
	if !exists {
		return nil, storage.ErrNotFound
	}

	return record.Task, nil
}

// ListPendingTasks returns all tasks with pending status
func (s *InMemoryAggregatorStore) ListPendingTasks(ctx context.Context) ([]*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	var pendingTasks []*types.Task
	for _, record := range s.tasks {
		if record.Status == storage.TaskStatusPending {
			pendingTasks = append(pendingTasks, record.Task)
		}
	}

	return pendingTasks, nil
}

// ListPendingTasksForAVS returns all pending tasks for a specific AVS address
func (s *InMemoryAggregatorStore) ListPendingTasksForAVS(ctx context.Context, avsAddress string) ([]*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	var pendingTasks []*types.Task
	for _, record := range s.tasks {
		if record.Status == storage.TaskStatusPending &&
			strings.EqualFold(record.Task.AVSAddress, avsAddress) {
			pendingTasks = append(pendingTasks, record.Task)
		}
	}

	return pendingTasks, nil
}

// UpdateTaskStatus updates the status of a task
func (s *InMemoryAggregatorStore) UpdateTaskStatus(ctx context.Context, taskId string, status storage.TaskStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	record, exists := s.tasks[taskId]
	if !exists {
		return storage.ErrNotFound
	}

	// Validate status transition
	if err := validateTaskStatusTransition(record.Status, status); err != nil {
		return err
	}

	record.Status = status
	record.UpdatedAt = time.Now()
	return nil
}

// DeleteTask removes a task from storage
func (s *InMemoryAggregatorStore) DeleteTask(ctx context.Context, taskId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	if _, exists := s.tasks[taskId]; !exists {
		return storage.ErrNotFound
	}

	delete(s.tasks, taskId)
	return nil
}

// SaveBlock saves block information for reorg detection
func (s *InMemoryAggregatorStore) SaveBlock(ctx context.Context, avsAddress string, block *storage.BlockRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	key := makeBlockRecordKey(avsAddress, block.ChainId, block.Number)
	s.blocks[key] = block
	key = makeBlockKey(avsAddress, block.ChainId)
	s.lastProcessedBlocks[key] = block.Number

	return nil
}

// GetBlock retrieves block information by block number
func (s *InMemoryAggregatorStore) GetBlock(ctx context.Context, avsAddress string, chainId config.ChainId, blockNumber uint64) (*storage.BlockRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	key := makeBlockRecordKey(avsAddress, chainId, blockNumber)
	block, exists := s.blocks[key]
	if !exists {
		return nil, storage.ErrNotFound
	}

	return block, nil
}

// DeleteBlock removes block information from storage
func (s *InMemoryAggregatorStore) DeleteBlock(ctx context.Context, avsAddress string, chainId config.ChainId, blockNumber uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	key := makeBlockRecordKey(avsAddress, chainId, blockNumber)
	if _, exists := s.blocks[key]; !exists {
		return storage.ErrNotFound
	}

	delete(s.blocks, key)
	return nil
}

// Close closes the store
func (s *InMemoryAggregatorStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	s.closed = true

	// Clear all maps
	s.lastProcessedBlocks = nil
	s.tasks = nil
	s.operatorSetConfigs = nil
	s.avsConfigs = nil
	s.blocks = nil

	return nil
}

func validateTaskStatusTransition(from, to storage.TaskStatus) error {
	// Define valid transitions
	validTransitions := map[storage.TaskStatus][]storage.TaskStatus{
		storage.TaskStatusPending:    {storage.TaskStatusProcessing, storage.TaskStatusFailed},
		storage.TaskStatusProcessing: {storage.TaskStatusCompleted, storage.TaskStatusFailed},
		storage.TaskStatusCompleted:  {}, // Terminal state
		storage.TaskStatusFailed:     {}, // Terminal state
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("%w: unknown status %s", storage.ErrInvalidTaskStatus, from)
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return nil
		}
	}

	return fmt.Errorf("%w: cannot transition from %s to %s", storage.ErrInvalidTaskStatus, from, to)
}
