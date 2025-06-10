package containerManager

import (
	"context"

	"go.uber.org/zap"
)

// EventManager handles safe event channel operations
type EventManager struct {
	logger *zap.Logger
}

// NewEventManager creates a new event manager
func NewEventManager(logger *zap.Logger) *EventManager {
	return &EventManager{
		logger: logger,
	}
}

// SendEvent safely sends an event to a channel, handling closed channels and context cancellation
func (em *EventManager) SendEvent(ctx context.Context, eventChan chan<- ContainerEvent, event ContainerEvent) {
	select {
	case eventChan <- event:
		// Event sent successfully
	case <-ctx.Done():
		em.logger.Debug("Context cancelled while sending event",
			zap.String("containerID", event.ContainerID),
			zap.String("eventType", string(event.Type)),
		)
	default:
		// Channel might be closed or full, log and continue
		em.logger.Debug("Could not send event, channel closed or full",
			zap.String("containerID", event.ContainerID),
			zap.String("eventType", string(event.Type)),
		)
	}
}

// SendEventNonBlocking sends an event without blocking, returns true if sent successfully
func (em *EventManager) SendEventNonBlocking(eventChan chan<- ContainerEvent, event ContainerEvent) bool {
	select {
	case eventChan <- event:
		return true
	default:
		em.logger.Debug("Could not send event, channel closed or full",
			zap.String("containerID", event.ContainerID),
			zap.String("eventType", string(event.Type)),
		)
		return false
	}
}
