package performerPoolManager

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/serverPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/planner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// PerformerHealth tracks the health status of a performer
type PerformerHealth struct {
	ContainerId  string
	LastChecked  time.Time
	Healthy      bool
	FailureCount int
	LastError    error
}

// PerformerPool manages the performers for a specific AVS
type PerformerPool struct {
	logger      *zap.Logger
	networkName string
	// TODO: likely use config for avsAddress and signingCurve
	avsAddress     string
	signingCurve   string
	performers     map[string]avsPerformer.IAvsPerformer
	peeringFetcher peering.IPeeringDataFetcher
	dockerClient   *client.Client

	// Health tracking
	healthStatus map[string]*PerformerHealth
	statusMutex  sync.RWMutex
}

// NewPerformerPool creates a new performer pool for a specific AVS
func NewPerformerPool(
	networkName string,
	avsConfig *avsPerformer.AvsPerformerConfig,
	dockerClient *client.Client,
	logger *zap.Logger,
	peeringFetcher peering.IPeeringDataFetcher,
) *PerformerPool {
	return &PerformerPool{
		logger:         logger,
		networkName:    networkName,
		avsAddress:     strings.ToLower(avsConfig.AvsAddress),
		signingCurve:   avsConfig.SigningCurve,
		performers:     make(map[string]avsPerformer.IAvsPerformer),
		peeringFetcher: peeringFetcher,
		dockerClient:   dockerClient,
		healthStatus:   make(map[string]*PerformerHealth),
	}
}

// createPerformer creates a new performer instance
func (p *PerformerPool) createPerformer(ctx context.Context, performerId string, artifact *planner.ArtifactVersion) error {
	// Check if performer already exists
	if _, exists := p.performers[performerId]; exists {
		return fmt.Errorf("performer with ID %s already exists", performerId)
	}

	performer, err := serverPerformer.NewAvsPerformerServer(
		&avsPerformer.AvsPerformerConfig{
			AvsAddress:           p.avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Registry: artifact.RegistryUrl, Digest: artifact.Digest},
			PerformerNetworkName: p.networkName,
			SigningCurve:         p.signingCurve,
		},
		p.peeringFetcher,
		p.logger,
	)
	if err != nil {
		p.logger.Sugar().Errorw("Failed to create AVS performer server",
			zap.String("avsAddress", p.avsAddress),
			zap.String("performerId", performerId),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create AVS performer server: %v", err)
	}

	if err := performer.Initialize(ctx); err != nil {
		p.logger.Sugar().Errorw("Failed to initialize performer",
			zap.String("avsAddress", p.avsAddress),
			zap.String("performerId", performerId),
			zap.Error(err),
		)
		return err
	}

	p.performers[performerId] = performer

	// Store performer health status with container ID
	p.statusMutex.Lock()
	p.healthStatus[performerId] = &PerformerHealth{
		ContainerId: performer.GetContainerId(),
		Healthy:     true,
		LastChecked: time.Now(),
	}
	p.statusMutex.Unlock()

	return nil
}

// ExecutePlan executes a capacity plan for this AVS
func (p *PerformerPool) ExecutePlan(ctx context.Context, plan *planner.PerformerCapacityPlan) error {
	p.logger.Sugar().Infow("Executing capacity plan",
		zap.String("avsAddress", p.avsAddress),
		zap.Int("targetCount", plan.TargetCount),
		zap.String("digest", plan.Digest),
	)

	// First check health of all performers
	healthyCount, err := p.CheckHealth(ctx)
	if err != nil {
		p.logger.Sugar().Errorw("Error checking health of performers",
			zap.String("avsAddress", p.avsAddress),
			zap.Error(err),
		)
		// TODO: double click on this.
	}

	// Calculate the difference between healthy count and target
	diff := plan.TargetCount - healthyCount

	if diff > 0 {
		// Need to scale up
		return p.scaleUp(ctx, diff, plan.LatestArtifact)
	} else if diff < 0 {
		// Need to scale down
		return p.scaleDown(ctx, -diff)
	}

	// No changes needed
	return nil
}

// CheckHealth checks health of all performers in the pool and returns the count of healthy performers
func (p *PerformerPool) CheckHealth(ctx context.Context) (int, error) {
	p.logger.Sugar().Debugw("Checking health of performers",
		zap.String("avsAddress", p.avsAddress),
		zap.Int("performerCount", len(p.performers)),
	)

	// Track performers that need to be recreated
	var performersToRecreate []string
	healthyCount := 0

	// Check health of all existing performers
	for id, performer := range p.performers {
		healthy, err := p.checkPerformerHealth(ctx, id, performer)

		p.statusMutex.Lock()
		health := p.healthStatus[id]
		if health == nil {
			health = &PerformerHealth{}
			p.healthStatus[id] = health
		}

		health.LastChecked = time.Now()
		health.Healthy = healthy

		if healthy {
			health.FailureCount = 0
			health.LastError = nil
			healthyCount++
		} else {
			health.FailureCount++
			health.LastError = err
			p.logger.Sugar().Warnw("Performer unhealthy",
				zap.String("avsAddress", p.avsAddress),
				zap.String("performerId", id),
				zap.Int("failureCount", health.FailureCount),
				zap.Error(err),
			)

			// Mark for recreation if failed too many times
			// Hard limit of 3 failures for now, could be configurable in the future
			if health.FailureCount >= 3 {
				p.logger.Sugar().Warnw("Marking performer for recreation due to repeated failures",
					zap.String("avsAddress", p.avsAddress),
					zap.String("performerId", id),
					zap.Int("failureCount", health.FailureCount),
				)
				performersToRecreate = append(performersToRecreate, id)
			}
		}
		p.statusMutex.Unlock()
	}

	// Recreate unhealthy performers
	for _, id := range performersToRecreate {
		// Don't count the performers we're going to recreate in the healthy count,
		// since we'll remove them first before creating new ones
		if err := p.recreatePerformer(ctx, id); err != nil {
			p.logger.Sugar().Errorw("Failed to recreate unhealthy performer",
				zap.String("avsAddress", p.avsAddress),
				zap.String("performerId", id),
				zap.Error(err),
			)
		} else {
			// If recreation was successful, count it as healthy
			// It's a new performer, so we add it to the healthy count
			healthyCount++
		}
	}

	return healthyCount, nil
}

// checkPerformerHealth checks if a specific performer is healthy
func (p *PerformerPool) checkPerformerHealth(ctx context.Context, performerId string, performer avsPerformer.IAvsPerformer) (bool, error) {
	// First check if the Docker container is still running
	p.statusMutex.RLock()
	health, ok := p.healthStatus[performerId]
	p.statusMutex.RUnlock()

	if ok && health.ContainerId != "" {
		// Check container health via Docker API
		if !p.isContainerHealthy(ctx, health.ContainerId) {
			return false, fmt.Errorf("container is not running or healthy")
		}
	}

	// Then check via the performer's health check (RunTask)
	taskCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dummyTask := &performerTask.PerformerTask{
		TaskID:  "health-check",
		Avs:     p.avsAddress,
		Payload: []byte("health-check"),
	}

	_, err := performer.RunTask(taskCtx, dummyTask)
	if err != nil {
		return false, err
	}

	return true, nil
}

// isContainerHealthy checks if the container is running via Docker API
func (p *PerformerPool) isContainerHealthy(ctx context.Context, containerId string) bool {
	if containerId == "" || p.dockerClient == nil {
		return false
	}

	inspection, err := p.dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		p.logger.Sugar().Warnw("Failed to inspect container",
			zap.String("avsAddress", p.avsAddress),
			zap.String("containerId", containerId),
			zap.Error(err),
		)
		return false
	}

	if !inspection.State.Running {
		p.logger.Sugar().Warnw("Container is not running",
			zap.String("avsAddress", p.avsAddress),
			zap.String("containerId", containerId),
			zap.Any("state", inspection.State.Status),
		)
		return false
	}

	// Check health status if available
	if inspection.State.Health != nil {
		p.logger.Sugar().Debugw("Container health status",
			zap.String("avsAddress", p.avsAddress),
			zap.String("containerId", containerId),
			zap.String("health", inspection.State.Health.Status),
		)

		return inspection.State.Health.Status == "healthy"
	}

	// If no health check configured, just ensure it's running
	return inspection.State.Running
}

// recreatePerformer recreates a failed performer instance
func (p *PerformerPool) recreatePerformer(ctx context.Context, performerId string) error {
	p.logger.Sugar().Infow("Recreating performer",
		zap.String("avsAddress", p.avsAddress),
		zap.String("performerId", performerId),
	)

	// Get the performer to shut down
	performer, exists := p.performers[performerId]
	if !exists {
		return fmt.Errorf("performer %s does not exist", performerId)
	}

	// Shutdown the performer
	if err := performer.Shutdown(); err != nil {
		p.logger.Sugar().Errorw("Failed to shutdown unhealthy performer",
			zap.String("avsAddress", p.avsAddress),
			zap.String("performerId", performerId),
			zap.Error(err),
		)
		// Continue with recreation even if shutdown failed
	}

	// Remove from maps
	delete(p.performers, performerId)
	p.statusMutex.Lock()
	delete(p.healthStatus, performerId)
	p.statusMutex.Unlock()

	// Create a new performer with the same ID
	return p.createPerformer(ctx, performerId, nil)
}

// scaleUp creates new performers to reach the target count
func (p *PerformerPool) scaleUp(ctx context.Context, countToAdd int, artifact *planner.ArtifactVersion) error {
	p.logger.Sugar().Infow("Scaling up performers",
		zap.String("avsAddress", p.avsAddress),
		zap.Int("currentCount", len(p.performers)),
		zap.Int("countToAdd", countToAdd),
		zap.String("artifact", artifact.Digest),
	)

	// Create the required number of performers
	for i := 0; i < countToAdd; i++ {
		performerId := fmt.Sprintf("performer-%d", len(p.performers)+1)

		// TODO: make this async as pulling a container can block other callers.
		if err := p.createPerformer(ctx, performerId, artifact); err != nil {
			p.logger.Sugar().Errorw("Failed to create performer during scale up",
				zap.String("avsAddress", p.avsAddress),
				zap.String("performerId", performerId),
				zap.Error(err),
			)
			return err
		}

		p.logger.Sugar().Infow("Created new performer",
			zap.String("avsAddress", p.avsAddress),
			zap.String("performerId", performerId),
		)
	}

	return nil
}

// scaleDown removes performers to reach the target count
func (p *PerformerPool) scaleDown(ctx context.Context, countToRemove int) error {
	p.logger.Sugar().Infow("Scaling down performers",
		zap.String("avsAddress", p.avsAddress),
		zap.Int("currentCount", len(p.performers)),
		zap.Int("countToRemove", countToRemove),
	)

	// Get list of performers to remove, prioritizing unhealthy performers first
	performersToRemove := p.selectPerformersToRemove(countToRemove)

	// Remove the identified performers
	for _, id := range performersToRemove {
		if err := p.removePerformer(ctx, id); err != nil {
			p.logger.Sugar().Errorw("Failed to remove performer",
				zap.String("avsAddress", p.avsAddress),
				zap.String("performerId", id),
				zap.Error(err),
			)
			// Continue with other removals even if this one failed
		}
	}

	return nil
}

// selectPerformersToRemove identifies performers to remove, prioritizing unhealthy ones
func (p *PerformerPool) selectPerformersToRemove(count int) []string {
	var performersToRemove []string

	// First, identify unhealthy performers
	p.statusMutex.RLock()
	for id, health := range p.healthStatus {
		if !health.Healthy && len(performersToRemove) < count {
			performersToRemove = append(performersToRemove, id)
		}
	}
	p.statusMutex.RUnlock()

	// If we still need to remove more, add the remaining performers
	if len(performersToRemove) < count {
		for id := range p.performers {
			// Skip if already marked for removal
			if containsString(performersToRemove, id) {
				continue
			}

			performersToRemove = append(performersToRemove, id)

			if len(performersToRemove) >= count {
				break
			}
		}
	}

	return performersToRemove
}

// removePerformer shuts down and removes a performer
func (p *PerformerPool) removePerformer(ctx context.Context, performerId string) error {
	p.logger.Sugar().Infow("Removing performer",
		zap.String("avsAddress", p.avsAddress),
		zap.String("performerId", performerId),
	)

	performer, exists := p.performers[performerId]
	if !exists {
		return fmt.Errorf("performer %s does not exist", performerId)
	}

	if err := performer.Shutdown(); err != nil {
		p.logger.Sugar().Errorw("Failed to shutdown performer",
			zap.String("avsAddress", p.avsAddress),
			zap.String("performerId", performerId),
			zap.Error(err),
		)
		return err
	}

	delete(p.performers, performerId)
	p.statusMutex.Lock()
	delete(p.healthStatus, performerId)
	p.statusMutex.Unlock()

	return nil
}

// Shutdown shuts down all performers in the pool
func (p *PerformerPool) Shutdown() error {
	p.logger.Sugar().Infow("Shutting down performer pool",
		zap.String("avsAddress", p.avsAddress),
		zap.Int("performerCount", len(p.performers)),
	)

	var lastErr error
	for id, performer := range p.performers {
		if err := performer.Shutdown(); err != nil {
			p.logger.Sugar().Errorw("Failed to shutdown performer",
				zap.String("avsAddress", p.avsAddress),
				zap.String("performerId", id),
				zap.Error(err),
			)
			lastErr = err
		}
		delete(p.performers, id)
	}

	return lastErr
}

// GetHealthyPerformer returns a healthy performer from the pool if available
func (p *PerformerPool) GetHealthyPerformer() (avsPerformer.IAvsPerformer, bool) {
	p.statusMutex.RLock()
	defer p.statusMutex.RUnlock()

	// Simple strategy: return the first healthy performer
	for id, performer := range p.performers {
		health, ok := p.healthStatus[id]
		if ok && health.Healthy {
			return performer, true
		}
	}

	// If no healthy performers, return the first one (if any)
	if len(p.performers) > 0 {
		for _, performer := range p.performers {
			return performer, false
		}
	}

	return nil, false
}

// GetPerformerCount returns the number of performers in the pool
func (p *PerformerPool) GetPerformerCount() int {
	return len(p.performers)
}

// GetAvsAddress returns the AVS address for this pool
func (p *PerformerPool) GetAvsAddress() string {
	return p.avsAddress
}

// containsString checks if a string is in a slice
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
