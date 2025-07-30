package storage

import (
	"context"
	"time"
)

// ExecutorStore defines the interface for executor state persistence
type ExecutorStore interface {
	// Performer state management
	SavePerformerState(ctx context.Context, performerId string, state *PerformerState) error
	GetPerformerState(ctx context.Context, performerId string) (*PerformerState, error)
	ListPerformerStates(ctx context.Context) ([]*PerformerState, error)
	DeletePerformerState(ctx context.Context, performerId string) error

	// Task tracking
	SaveInflightTask(ctx context.Context, taskId string, task *TaskInfo) error
	GetInflightTask(ctx context.Context, taskId string) (*TaskInfo, error)
	ListInflightTasks(ctx context.Context) ([]*TaskInfo, error)
	DeleteInflightTask(ctx context.Context, taskId string) error

	// Deployment tracking
	SaveDeployment(ctx context.Context, deploymentId string, deployment *DeploymentInfo) error
	GetDeployment(ctx context.Context, deploymentId string) (*DeploymentInfo, error)
	UpdateDeploymentStatus(ctx context.Context, deploymentId string, status DeploymentStatus) error

	// Lifecycle management
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
	DeploymentMode     string // "docker" or "kubernetes"
	CreatedAt          time.Time
	LastHealthCheck    time.Time
	ContainerHealthy   bool
	ApplicationHealthy bool
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

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusFailed    DeploymentStatus = "failed"
)
