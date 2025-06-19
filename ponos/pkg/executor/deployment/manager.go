package deployment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// DefaultDeploymentTimeout is the default timeout for deployments
	DefaultDeploymentTimeout = 1 * time.Minute
	DefaultCleanupTimeout    = 5 * time.Second
)

// Manager manages container deployments for AVS performers
type Manager struct {
	logger            *zap.Logger
	deploymentMu      sync.Mutex
	activeDeployments map[string]*Deployment
	deploymentHistory map[string]*DeploymentResult
}

// NewManager creates a new deployment manager
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		logger:            logger,
		activeDeployments: make(map[string]*Deployment),
		deploymentHistory: make(map[string]*DeploymentResult),
	}
}

// Deploy initiates a deployment for an AVS performer
func (m *Manager) Deploy(ctx context.Context, config DeploymentConfig, performer avsPerformer.IAvsPerformer) (*DeploymentResult, error) {
	avsAddress := strings.ToLower(config.AvsAddress)
	deploymentID := m.generateDeploymentID(avsAddress)

	m.logger.Info("Starting deployment",
		zap.String("deploymentID", deploymentID),
		zap.String("avsAddress", avsAddress),
		zap.String("repository", config.Image.Repository),
		zap.String("tag", config.Image.Tag),
	)

	if config.Timeout == 0 {
		config.Timeout = DefaultDeploymentTimeout
	}

	deploymentCtx, cancel := context.WithTimeout(ctx, config.Timeout)

	// Create deployment record
	deployment := &Deployment{
		ID:         deploymentID,
		AvsAddress: config.AvsAddress,
		Image:      config.Image,
		Status:     DeploymentStatusPending,
		StartTime:  time.Now(),
		CancelFunc: cancel,
	}

	// Check for existing deployment and store new one
	m.deploymentMu.Lock()
	if existingDeployment, exists := m.activeDeployments[avsAddress]; exists {
		m.logger.Error("Found existing deployment in activeDeployments",
			zap.String("deploymentID", existingDeployment.ID),
			zap.String("status", string(existingDeployment.Status)),
		)
		m.deploymentMu.Unlock()
		cancel()
		return nil, NewDeploymentError(existingDeployment.ID, avsAddress, ErrDeploymentInProgress, "deployment already in progress")
	}
	// Store active deployment while still holding the lock
	m.activeDeployments[avsAddress] = deployment
	m.deploymentMu.Unlock()

	// Execute deployment
	result, err := m.executeDeployment(deploymentCtx, deployment, performer)

	// Move deployment from active to history
	m.deploymentMu.Lock()
	defer m.deploymentMu.Unlock()
	if currentDeployment, exists := m.activeDeployments[avsAddress]; exists && currentDeployment.ID == deploymentID {
		delete(m.activeDeployments, avsAddress)
	}
	if result != nil {
		m.deploymentHistory[deploymentID] = result
	}

	return result, err
}

// executeDeployment executes the deployment with the performer
func (m *Manager) executeDeployment(
	ctx context.Context,
	deployment *Deployment,
	performer avsPerformer.IAvsPerformer,
) (*DeploymentResult, error) {
	m.logger.Info("Starting executeDeployment",
		zap.String("deploymentID", deployment.ID),
		zap.String("avsAddress", deployment.AvsAddress),
		zap.String("image", fmt.Sprintf("%s:%s", deployment.Image.Repository, deployment.Image.Tag)),
	)

	deploymentInfo, err := performer.CreatePerformer(ctx, deployment.Image)

	if err != nil {
		m.logger.Error("Failed to deploy performer",
			zap.String("deploymentID", deployment.ID),
			zap.Error(err),
		)
		return &DeploymentResult{
			DeploymentID: deployment.ID,
			Status:       DeploymentStatusFailed,
			StartTime:    deployment.StartTime,
			EndTime:      time.Now(),
			Error:        err,
			Message:      fmt.Sprintf("Failed to deploy performer: %v", err),
		}, NewDeploymentError(deployment.ID, deployment.AvsAddress, err, "performer deployment failed")
	}

	// Store the performer ID in the deployment
	deployment.PerformerID = deploymentInfo.PerformerID
	deployment.StatusChan = deploymentInfo.StatusChan

	result, err := m.monitorDeployment(ctx, deployment, deploymentInfo.StatusChan)

	if err != nil {
		m.logger.Error("Deployment monitoring failed, canceling deployment",
			zap.String("deploymentID", deployment.ID),
			zap.Error(err),
		)

		m.logger.Info("Calling performer.RemovePerformer",
			zap.String("deploymentID", deployment.ID),
		)

		// Use a fresh context for cleanup since the original context might be cancelled
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), DefaultDeploymentTimeout)
		defer cleanupCancel()

		if cancelErr := performer.RemovePerformer(cleanupCtx, deployment.PerformerID); cancelErr != nil {
			m.logger.Error("Failed to cancel deployment after error",
				zap.String("deploymentID", deployment.ID),
				zap.Error(cancelErr),
			)
		} else {
			m.logger.Info("Successfully canceled deployment",
				zap.String("deploymentID", deployment.ID),
			)
		}
		return result, err
	}

	promoteErr := performer.PromotePerformer(ctx, deploymentInfo.PerformerID)

	if promoteErr != nil {
		m.logger.Error("Promotion failed",
			zap.String("deploymentID", deployment.ID),
			zap.String("performerID", deploymentInfo.PerformerID),
			zap.Error(promoteErr),
		)

		// Promotion failed
		result.Status = DeploymentStatusFailed
		result.EndTime = time.Now()
		result.Error = ErrPromotionFailed
		result.Message = fmt.Sprintf("Failed to promote performer: %v", promoteErr)

		m.logger.Info("Updating deployment status to failed",
			zap.String("deploymentID", deployment.ID),
		)

		m.updateDeploymentStatus(deployment.ID, DeploymentStatusFailed)

		// Cancel deployment after promotion failure
		cancelCtx, cancelFunc := context.WithTimeout(context.Background(), DefaultCleanupTimeout)
		defer cancelFunc()

		m.logger.Info("Calling performer.RemovePerformer after promotion failure",
			zap.String("deploymentID", deployment.ID),
		)

		if cancelErr := performer.RemovePerformer(cancelCtx, deploymentInfo.PerformerID); cancelErr != nil {
			m.logger.Error("Failed to cancel deployment after promotion failure",
				zap.String("deploymentID", deployment.ID),
				zap.Error(cancelErr),
			)
		} else {
			m.logger.Info("Successfully canceled deployment after promotion failure",
				zap.String("deploymentID", deployment.ID),
			)
		}

		return result, NewDeploymentError(deployment.ID, deployment.AvsAddress, ErrPromotionFailed, "performer promotion failed")
	}

	// Deployment completed successfully
	result.Status = DeploymentStatusCompleted
	result.EndTime = time.Now()
	result.Message = "Deployment completed successfully"

	m.updateDeploymentStatus(deployment.ID, DeploymentStatusCompleted)

	m.logger.Info("Deployment completed successfully",
		zap.String("deploymentID", deployment.ID),
		zap.String("performerID", deploymentInfo.PerformerID),
		zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
	)

	return result, nil
}

// monitorDeployment monitors a deployment until it's healthy or fails
func (m *Manager) monitorDeployment(
	ctx context.Context,
	deployment *Deployment,
	statusChan <-chan avsPerformer.PerformerStatusEvent,
) (*DeploymentResult, error) {
	startTime := time.Now()

	result := &DeploymentResult{
		DeploymentID: deployment.ID,
		StartTime:    startTime,
		Status:       DeploymentStatusInProgress,
		PerformerID:  deployment.PerformerID,
	}

	m.updateDeploymentStatus(deployment.ID, DeploymentStatusInProgress)

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Context done in monitorDeployment",
				zap.String("deploymentID", deployment.ID),
				zap.Error(ctx.Err()),
			)

			// Check if this is a deadline exceeded (timeout) or cancellation
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				result.Status = DeploymentStatusFailed
				result.EndTime = time.Now()
				result.Error = ErrDeploymentTimeout
				result.Message = fmt.Sprintf("Deployment timed out")
				m.updateDeploymentStatus(deployment.ID, DeploymentStatusFailed)
				return result, ErrDeploymentTimeout
			}

			// Otherwise it's a cancellation
			result.Status = DeploymentStatusCancelled
			result.EndTime = time.Now()
			result.Error = ctx.Err()
			result.Message = "Deployment cancelled"
			m.updateDeploymentStatus(deployment.ID, DeploymentStatusCancelled)
			return result, ErrDeploymentCancelled

		case status, ok := <-statusChan:
			m.logger.Info("Received status event",
				zap.String("deploymentID", deployment.ID),
				zap.Bool("channelOpen", ok),
			)

			if !ok {
				// Channel closed unexpectedly
				m.logger.Error("Status channel closed unexpectedly",
					zap.String("deploymentID", deployment.ID),
				)
				result.Status = DeploymentStatusFailed
				result.EndTime = time.Now()
				result.Error = ErrDeploymentFailed
				result.Message = "Deployment failed unexpectedly. Please check the logs for more details."
				m.updateDeploymentStatus(deployment.ID, DeploymentStatusFailed)
				return result, ErrDeploymentFailed
			}

			// Process status event
			switch status.Status {
			case avsPerformer.PerformerHealthy:
				m.logger.Info("Performer is healthy, deployment successful",
					zap.String("deploymentID", deployment.ID),
					zap.String("performerID", status.PerformerID),
					zap.Duration("totalTime", time.Since(startTime)),
				)
				result.Status = DeploymentStatusHealthy
				result.PerformerID = status.PerformerID
				result.Message = "Performer is healthy and ready"
				m.updateDeploymentStatus(deployment.ID, DeploymentStatusHealthy)
				result.EndTime = time.Now()
				return result, nil

			case avsPerformer.PerformerUnhealthy:
				m.logger.Warn("Performer is unhealthy, continuing to monitor",
					zap.String("deploymentID", deployment.ID),
					zap.String("performerID", status.PerformerID),
					zap.String("message", status.Message),
					zap.Duration("elapsed", time.Since(startTime)),
				)
				result.Status = DeploymentStatusUnhealthy
				result.PerformerID = status.PerformerID
				result.Message = status.Message
				m.updateDeploymentStatus(deployment.ID, DeploymentStatusUnhealthy)

			default:
				m.logger.Warn("Unknown performer status",
					zap.String("deploymentID", deployment.ID),
					zap.String("status", fmt.Sprintf("%v", status.Status)),
				)
			}
		}
	}
}

// CancelDeployment cancels an active deployment
func (m *Manager) CancelDeployment(avsAddress string) error {
	avsAddress = strings.ToLower(avsAddress)

	m.deploymentMu.Lock()
	deployment, exists := m.activeDeployments[avsAddress]
	m.deploymentMu.Unlock()

	if !exists {
		return fmt.Errorf("no active deployment found for AVS %s", avsAddress)
	}

	m.logger.Info("Cancelling deployment",
		zap.String("deploymentID", deployment.ID),
		zap.String("avsAddress", avsAddress),
	)

	deployment.CancelFunc()
	return nil
}

// GetActiveDeployment returns the active deployment for an AVS
func (m *Manager) GetActiveDeployment(avsAddress string) (*Deployment, bool) {
	avsAddress = strings.ToLower(avsAddress)

	m.deploymentMu.Lock()
	deployment, exists := m.activeDeployments[avsAddress]
	m.deploymentMu.Unlock()

	return deployment, exists
}

// GetDeploymentResult returns the result of a deployment by ID
func (m *Manager) GetDeploymentResult(deploymentID string) (*DeploymentResult, bool) {
	m.deploymentMu.Lock()
	result, exists := m.deploymentHistory[deploymentID]
	m.deploymentMu.Unlock()

	return result, exists
}

// generateDeploymentID generates a unique deployment ID
func (m *Manager) generateDeploymentID(avsAddress string) string {
	return fmt.Sprintf("deployment-%s-%s", avsAddress, uuid.New().String())
}

// updateDeploymentStatus updates the status of a deployment
func (m *Manager) updateDeploymentStatus(deploymentID string, status DeploymentStatus) {
	m.deploymentMu.Lock()
	defer m.deploymentMu.Unlock()

	// Check active deployments
	for _, deployment := range m.activeDeployments {
		if deployment.ID == deploymentID {
			deployment.Status = status
			return
		}
	}
}
