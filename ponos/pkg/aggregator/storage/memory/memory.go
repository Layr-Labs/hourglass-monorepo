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
	blocks              map[string]*storage.BlockEntity
}

// NewInMemoryAggregatorStore creates a new in-memory aggregator store
func NewInMemoryAggregatorStore() *InMemoryAggregatorStore {
	return &InMemoryAggregatorStore{
		lastProcessedBlocks: make(map[string]uint64),
		tasks:               make(map[string]*storage.TaskRecord),
		operatorSetConfigs:  make(map[string]*storage.OperatorSetTaskConfig),
		avsConfigs:          make(map[string]*storage.AvsConfig),
		blocks:              make(map[string]*storage.BlockEntity),
	}
}

// makeBlockKey creates a composite key for last processed block storage
func makeBlockKey(avsAddress string, chainId config.ChainId) string {
	return fmt.Sprintf("%s:%d", avsAddress, chainId)
}

// makeBlockInfoKey creates a composite key for block info storage
func makeBlockInfoKey(avsAddress string, chainId config.ChainId, blockNumber uint64) string {
	return fmt.Sprintf("block:%s:%d:%d", avsAddress, chainId, blockNumber)
}

// GetLastProcessedBlock returns the last processed block for a chain for a specific AVS
func (s *InMemoryAggregatorStore) GetLastProcessedBlock(ctx context.Context, avsAddress string, chainId config.ChainId) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return 0, storage.ErrStoreClosed
	}

	key := makeBlockKey(avsAddress, chainId)
	blockNum, exists := s.lastProcessedBlocks[key]
	if !exists {
		return 0, storage.ErrNotFound
	}

	return blockNum, nil
}

// SetLastProcessedBlock sets the last processed block for a chain for a specific AVS
func (s *InMemoryAggregatorStore) SetLastProcessedBlock(ctx context.Context, avsAddress string, chainId config.ChainId, blockNum uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	key := makeBlockKey(avsAddress, chainId)
	s.lastProcessedBlocks[key] = blockNum
	return nil
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

// SaveOperatorSetConfig saves operator set configuration
func (s *InMemoryAggregatorStore) SaveOperatorSetConfig(ctx context.Context, avsAddress string, operatorSetId uint32, config *storage.OperatorSetTaskConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	key := makeOperatorSetKey(avsAddress, operatorSetId)
	s.operatorSetConfigs[key] = config
	return nil
}

// GetOperatorSetConfig retrieves operator set configuration
func (s *InMemoryAggregatorStore) GetOperatorSetConfig(ctx context.Context, avsAddress string, operatorSetId uint32) (*storage.OperatorSetTaskConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	key := makeOperatorSetKey(avsAddress, operatorSetId)
	config, exists := s.operatorSetConfigs[key]
	if !exists {
		return nil, storage.ErrNotFound
	}

	return config, nil
}

// SaveAVSConfig saves AVS configuration
func (s *InMemoryAggregatorStore) SaveAVSConfig(ctx context.Context, avsAddress string, config *storage.AvsConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	s.avsConfigs[avsAddress] = config
	return nil
}

// GetAVSConfig retrieves AVS configuration
func (s *InMemoryAggregatorStore) GetAVSConfig(ctx context.Context, avsAddress string) (*storage.AvsConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	config, exists := s.avsConfigs[avsAddress]
	if !exists {
		return nil, storage.ErrNotFound
	}

	return config, nil
}

// SaveBlock saves block information for reorg detection
func (s *InMemoryAggregatorStore) SaveBlock(ctx context.Context, avsAddress string, block *storage.BlockEntity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return storage.ErrStoreClosed
	}

	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	key := makeBlockInfoKey(avsAddress, block.ChainId, block.Number)
	s.blocks[key] = block
	return nil
}

// GetBlock retrieves block information by block number
func (s *InMemoryAggregatorStore) GetBlock(ctx context.Context, avsAddress string, chainId config.ChainId, blockNumber uint64) (*storage.BlockEntity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, storage.ErrStoreClosed
	}

	key := makeBlockInfoKey(avsAddress, chainId, blockNumber)
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

	key := makeBlockInfoKey(avsAddress, chainId, blockNumber)
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

// Helper functions

func makeOperatorSetKey(avsAddress string, operatorSetId uint32) string {
	return fmt.Sprintf("%s:%d", avsAddress, operatorSetId)
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
