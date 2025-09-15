package storage

import (
	"context"
	"time"
)

// EnvironmentVarRecord represents a single environment variable configuration
type EnvironmentVarRecord struct {
	Name         string `json:"name"`
	Value        string `json:"value,omitempty"`
	ValueFromEnv string `json:"valueFromEnv,omitempty"`
}

// ExecutorStore defines the interface for executor state persistence
type ExecutorStore interface {
	SavePerformerState(ctx context.Context, performerId string, state *PerformerState) error
	GetPerformerState(ctx context.Context, performerId string) (*PerformerState, error)
	ListPerformerStates(ctx context.Context) ([]*PerformerState, error)
	DeletePerformerState(ctx context.Context, performerId string) error

	MarkTaskProcessed(ctx context.Context, taskId string) error
	IsTaskProcessed(ctx context.Context, taskId string) (bool, error)

	Close() error
}

// PerformerState represents the persisted state of an AVS performer
type PerformerState struct {
	PerformerId        string
	AvsAddress         string
	ContainerId        string
	Status             string
	ArtifactRegistry   string
	ArtifactDigest     string
	ArtifactTag        string
	DeploymentMode     string
	CreatedAt          time.Time
	LastHealthCheck    time.Time
	ContainerHealthy   bool
	ApplicationHealthy bool
	NetworkName        string
	ContainerEndpoint  string
	ContainerHostname  string
	InternalPort       int
	EnvironmentVars    []EnvironmentVarRecord
}

// TaskInfo represents information about an inflight task
type TaskInfo struct {
	TaskId            string
	AvsAddress        string
	OperatorAddress   string
	ReceivedAt        time.Time
	Status            string
	AggregatorAddress string
	OperatorSetId     uint32
}

// DeploymentInfo tracks deployment information
type DeploymentInfo struct {
	DeploymentId     string
	AvsAddress       string
	ArtifactRegistry string
	ArtifactDigest   string
	Status           DeploymentStatus
	StartedAt        time.Time
	CompletedAt      *time.Time
	Error            string
}

// ProcessedTask represents a task that has been processed
type ProcessedTask struct {
	TaskId      string    `json:"taskId"`
	ProcessedAt time.Time `json:"processedAt"`
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusFailed    DeploymentStatus = "failed"
)
