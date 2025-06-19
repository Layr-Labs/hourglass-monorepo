package deployment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
)

// Standard deployment errors
var (
	// ErrDeploymentInProgress indicates a deployment is already in progress for the AVS
	ErrDeploymentInProgress = errors.New("deployment already in progress")

	// ErrDeploymentTimeout indicates the deployment exceeded the timeout
	ErrDeploymentTimeout = errors.New("deployment timeout")

	// ErrDeploymentFailed indicates the deployment failed
	ErrDeploymentFailed = errors.New("deployment failed")

	// ErrDeploymentCancelled indicates the deployment was cancelled
	ErrDeploymentCancelled = errors.New("deployment cancelled")

	// ErrPerformerUnhealthy indicates the performer failed health checks
	ErrPerformerUnhealthy = errors.New("performer unhealthy")

	// ErrPromotionFailed indicates performer promotion failed
	ErrPromotionFailed = errors.New("performer promotion failed")
)

// DeploymentStatus represents the current state of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "pending"
	DeploymentStatusInProgress DeploymentStatus = "in_progress"
	DeploymentStatusHealthy    DeploymentStatus = "healthy"
	DeploymentStatusUnhealthy  DeploymentStatus = "unhealthy"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusCompleted  DeploymentStatus = "completed"
	DeploymentStatusCancelled  DeploymentStatus = "cancelled"
)

// DeploymentConfig contains configuration for a deployment
type DeploymentConfig struct {
	AvsAddress string
	Image      avsPerformer.PerformerImage
	Timeout    time.Duration
}

// DeploymentResult contains the result of a deployment operation
type DeploymentResult struct {
	DeploymentID string
	Status       DeploymentStatus
	PerformerID  string
	StartTime    time.Time
	EndTime      time.Time
	Message      string
	Error        error
}

// Deployment represents an active deployment
type Deployment struct {
	ID          string
	AvsAddress  string
	PerformerID string
	Image       avsPerformer.PerformerImage
	Status      DeploymentStatus
	StartTime   time.Time
	EndTime     time.Time
	StatusChan  <-chan avsPerformer.PerformerStatusEvent
	CancelFunc  context.CancelFunc
}

// DeploymentError wraps deployment-specific errors with additional context
type DeploymentError struct {
	DeploymentID string
	AvsAddress   string
	Err          error
	Message      string
}

func (e *DeploymentError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("deployment %s for AVS %s: %s: %v", e.DeploymentID, e.AvsAddress, e.Message, e.Err)
	}
	return fmt.Sprintf("deployment %s for AVS %s: %v", e.DeploymentID, e.AvsAddress, e.Err)
}

func (e *DeploymentError) Unwrap() error {
	return e.Err
}

// NewDeploymentError creates a new deployment error
func NewDeploymentError(deploymentID, avsAddress string, err error, message string) error {
	return &DeploymentError{
		DeploymentID: deploymentID,
		AvsAddress:   avsAddress,
		Err:          err,
		Message:      message,
	}
}
