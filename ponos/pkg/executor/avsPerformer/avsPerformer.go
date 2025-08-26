package avsPerformer

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
)

type AvsProcessType string

const (
	AvsProcessTypeServer AvsProcessType = "server"
	AvsProcessTypeOneOff AvsProcessType = "one-off"
)

type PerformerImage struct {
	Repository         string
	Tag                string
	Digest             string
	Envs               []config.AVSPerformerEnv
	ServiceAccountName string // Optional service account name for Kubernetes deployments
}

// PerformerStatus represents the health status of a performer container
type PerformerStatus int

const (
	PerformerHealthUnknown PerformerStatus = iota
	PerformerHealthy
	PerformerUnhealthy
)

// PerformerStatusEvent contains status information sent to deployment watchers
type PerformerStatusEvent struct {
	Status      PerformerStatus
	PerformerID string
	Message     string
	Timestamp   time.Time
}

// PerformerHealth tracks the health state of a container
type PerformerHealth struct {
	ContainerIsHealthy                   bool
	ApplicationIsHealthy                 bool
	ConsecutiveApplicationHealthFailures int
	LastHealthCheck                      time.Time
}

// PerformerResourceStatus represents the deployment status of a performer container
type PerformerResourceStatus string

const (
	PerformerResourceStatusInService PerformerResourceStatus = "InService"
	PerformerResourceStatusStaged    PerformerResourceStatus = "Staged"
)

// PerformerMetadata holds information about a performer container
type PerformerMetadata struct {
	PerformerID        string
	AvsAddress         string
	ResourceID         string
	Status             PerformerResourceStatus
	ArtifactRegistry   string
	ArtifactTag        string
	ArtifactDigest     string
	ContainerHealthy   bool
	ApplicationHealthy bool
	LastHealthCheck    time.Time
}

type AvsPerformerConfig struct {
	AvsAddress                     string
	ProcessType                    AvsProcessType
	Image                          PerformerImage
	PerformerNetworkName           string
	EndpointOverride               string        // Optional: Override the auto-detected endpoint (for testing when executor is outside cluster)
	ApplicationHealthCheckInterval time.Duration // Interval for health checks on the application running in the performer container
	ServiceAccountName             string
}

// DeploymentStatus represents the current state of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "pending"
	DeploymentStatusInProgress DeploymentStatus = "in_progress"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusCompleted  DeploymentStatus = "completed"
	DeploymentStatusCancelled  DeploymentStatus = "cancelled"
)

// DeploymentResult contains the result of a deployment operation
type DeploymentResult struct {
	ID          string
	PerformerID string
	Status      DeploymentStatus
	Image       PerformerImage
	StartTime   time.Time
	EndTime     time.Time
	Message     string
	Error       error
}

// PerformerCreationResult contains information about a deployed performer
type PerformerCreationResult struct {
	PerformerID string
	StatusChan  <-chan PerformerStatusEvent
}

type IAvsPerformer interface {
	Initialize(ctx context.Context) error
	RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error)
	Deploy(ctx context.Context, image PerformerImage) (*DeploymentResult, error)
	CreatePerformer(ctx context.Context, image PerformerImage) (*PerformerCreationResult, error)
	PromotePerformer(ctx context.Context, performerID string) error
	RemovePerformer(ctx context.Context, performerID string) error
	ListPerformers() []PerformerMetadata
	Shutdown() error
}
