package storage

import (
	"context"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

// AggregatorStore defines the interface for aggregator state persistence
type AggregatorStore interface {
	GetLastProcessedBlock(ctx context.Context, avsAddress string, chainId config.ChainId) (*BlockRecord, error)

	SaveBlock(ctx context.Context, avsAddress string, block *BlockRecord) error
	GetBlock(ctx context.Context, avsAddress string, chainId config.ChainId, blockNumber uint64) (*BlockRecord, error)
	DeleteBlock(ctx context.Context, avsAddress string, chainId config.ChainId, blockNumber uint64) error

	SavePendingTask(ctx context.Context, task *types.Task) error
	GetTask(ctx context.Context, taskId string) (*types.Task, error)
	ListPendingTasks(ctx context.Context) ([]*types.Task, error)
	ListPendingTasksForAVS(ctx context.Context, avsAddress string) ([]*types.Task, error)
	UpdateTaskStatus(ctx context.Context, taskId string, status TaskStatus) error
	DeleteTask(ctx context.Context, taskId string) error

	Close() error
}

// TaskStatus represents the status of a task in the system
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// TaskRecord wraps a task with additional metadata for storage
type TaskRecord struct {
	Task      *types.Task
	Status    TaskStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BlockRecord stores essential block information for reorg detection
type BlockRecord struct {
	Number     uint64
	Hash       string
	ParentHash string
	Timestamp  uint64
	ChainId    config.ChainId
}

// ConsensusType represents the type of consensus mechanism
type ConsensusType uint8

const (
	ConsensusTypeNone                     ConsensusType = 0
	ConsensusTypeStakeProportionThreshold ConsensusType = 1
)

// OperatorSetTaskConsensus defines consensus parameters for an operator set
type OperatorSetTaskConsensus struct {
	ConsensusType ConsensusType
	Threshold     uint16
}

// OperatorSetTaskConfig contains configuration for operator set tasks
type OperatorSetTaskConfig struct {
	TaskSLA      int64
	CurveType    config.CurveType
	TaskMetadata []byte
	Consensus    OperatorSetTaskConsensus
}

// AvsConfig contains AVS-specific configuration
type AvsConfig struct {
	AggregatorOperatorSetId uint32
	ExecutorOperatorSetIds  []uint32
	CurveType               config.CurveType
}
