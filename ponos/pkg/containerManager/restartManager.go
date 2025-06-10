package containerManager

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RestartManager handles container restart logic and policies
type RestartManager struct {
	logger       *zap.Logger
	eventManager *EventManager
}

// NewRestartManager creates a new restart manager
func NewRestartManager(logger *zap.Logger) *RestartManager {
	return &RestartManager{
		logger:       logger,
		eventManager: NewEventManager(logger),
	}
}

// RestartDecision represents a restart decision
type RestartDecision struct {
	ShouldRestart bool
	Reason        string
	Delay         time.Duration
}

// ShouldRestart determines if a container should be restarted based on policy and current state
func (rm *RestartManager) ShouldRestart(monitor *containerMonitor, reason string) RestartDecision {
	decision := RestartDecision{
		ShouldRestart: false,
		Reason:        reason,
	}

	if !monitor.restartPolicy.Enabled {
		decision.Reason = "restart policy disabled"
		return decision
	}

	// Check restart count limit
	if monitor.restartPolicy.MaxRestarts > 0 && monitor.restartCount >= monitor.restartPolicy.MaxRestarts {
		decision.Reason = fmt.Sprintf("restart limit reached (%d/%d)", monitor.restartCount, monitor.restartPolicy.MaxRestarts)
		return decision
	}

	decision.ShouldRestart = true
	decision.Delay = rm.calculateRestartDelay(monitor)
	return decision
}

// AttemptRestart attempts to restart a container with proper error handling and events
func (rm *RestartManager) AttemptRestart(ctx context.Context, monitor *containerMonitor, restartFunc func(context.Context, string, time.Duration) error, reason string) error {
	decision := rm.ShouldRestart(monitor, reason)

	if !decision.ShouldRestart {
		rm.logger.Warn("Container restart denied",
			zap.String("containerID", monitor.containerID),
			zap.String("reason", decision.Reason),
		)

		// Send restart denied event
		event := ContainerEvent{
			ContainerID: monitor.containerID,
			Type:        EventRestartFailed,
			Timestamp:   time.Now(),
			Message:     fmt.Sprintf("Restart denied: %s", decision.Reason),
		}
		rm.eventManager.SendEvent(ctx, monitor.eventChan, event)

		return fmt.Errorf("restart denied: %s", decision.Reason)
	}

	// Apply restart delay
	if decision.Delay > 0 {
		rm.logger.Info("Applying restart delay",
			zap.String("containerID", monitor.containerID),
			zap.Duration("delay", decision.Delay),
			zap.Int("restartCount", monitor.restartCount),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(decision.Delay):
		}
	}

	// Send restarting event
	restartingEvent := ContainerEvent{
		ContainerID: monitor.containerID,
		Type:        EventRestarting,
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Restarting container (attempt %d/%d): %s", monitor.restartCount+1, monitor.restartPolicy.MaxRestarts, reason),
	}
	rm.eventManager.SendEvent(ctx, monitor.eventChan, restartingEvent)

	// Create restart context with timeout
	restartCtx := ctx
	if monitor.restartPolicy.RestartTimeout > 0 {
		var cancel context.CancelFunc
		restartCtx, cancel = context.WithTimeout(ctx, monitor.restartPolicy.RestartTimeout)
		defer cancel()
	}

	// Perform the restart
	if err := restartFunc(restartCtx, monitor.containerID, DefaultStopTimeout); err != nil {
		return rm.handleRestartError(ctx, monitor, err, reason)
	}

	// Send success event
	successEvent := ContainerEvent{
		ContainerID: monitor.containerID,
		Type:        EventRestarted,
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Container restarted successfully (attempt %d)", monitor.restartCount),
	}
	rm.eventManager.SendEvent(ctx, monitor.eventChan, successEvent)

	rm.logger.Info("Container restarted successfully",
		zap.String("containerID", monitor.containerID),
		zap.String("reason", reason),
		zap.Int("restartCount", monitor.restartCount),
	)

	return nil
}

// handleRestartError handles different types of restart errors
func (rm *RestartManager) handleRestartError(ctx context.Context, monitor *containerMonitor, err error, reason string) error {
	if strings.Contains(err.Error(), "No such container") {
		rm.logger.Info("Container doesn't exist, recreation needed",
			zap.String("containerID", monitor.containerID),
			zap.String("reason", reason),
		)

		// Send recreation needed event
		recreationEvent := ContainerEvent{
			ContainerID: monitor.containerID,
			Type:        EventRestartFailed,
			Timestamp:   time.Now(),
			Message:     fmt.Sprintf("Container recreation needed (attempt %d): %s", monitor.restartCount+1, err.Error()),
		}
		rm.eventManager.SendEvent(ctx, monitor.eventChan, recreationEvent)

		// Don't return error for backward compatibility
		rm.logger.Debug("Container restart completed with recreation needed signal",
			zap.String("containerID", monitor.containerID),
		)
		return nil
	}

	// Send generic restart failed event
	failEvent := ContainerEvent{
		ContainerID: monitor.containerID,
		Type:        EventRestartFailed,
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Restart failed (attempt %d): %s", monitor.restartCount+1, err.Error()),
	}
	rm.eventManager.SendEvent(ctx, monitor.eventChan, failEvent)

	return err
}

// calculateRestartDelay calculates the delay before attempting a restart based on backoff policy
func (rm *RestartManager) calculateRestartDelay(monitor *containerMonitor) time.Duration {
	if monitor.restartCount == 0 {
		return monitor.restartPolicy.RestartDelay
	}

	// Apply exponential backoff
	delay := monitor.restartPolicy.RestartDelay
	for i := 0; i < monitor.restartCount && delay < monitor.restartPolicy.MaxBackoffDelay; i++ {
		delay = time.Duration(float64(delay) * monitor.restartPolicy.BackoffMultiplier)
	}

	// Cap at max backoff delay
	if delay > monitor.restartPolicy.MaxBackoffDelay {
		delay = monitor.restartPolicy.MaxBackoffDelay
	}

	return delay
}
