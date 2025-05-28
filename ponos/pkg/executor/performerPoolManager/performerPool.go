package performerPoolManager

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/serverPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerCapacityPlanner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/docker/docker/api/types/container"
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
	logger         *zap.Logger
	networkName    string
	performers     map[string]avsPerformer.IAvsPerformer
	peeringFetcher peering.IPeeringDataFetcher
	dockerClient   *client.Client
	containerMgr   containerManager.IContainerManager
	avsConfig      *avsPerformer.AvsPerformerConfig

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
	containerMgr containerManager.IContainerManager,
) *PerformerPool {
	return &PerformerPool{
		logger:         logger,
		networkName:    networkName,
		performers:     make(map[string]avsPerformer.IAvsPerformer),
		peeringFetcher: peeringFetcher,
		dockerClient:   dockerClient,
		containerMgr:   containerMgr,
		avsConfig:      avsConfig,
		healthStatus:   make(map[string]*PerformerHealth),
	}
}

// ExecutePlan executes a capacity plan for this AVS
func (p *PerformerPool) ExecutePlan(
	ctx context.Context,
	plan *performerCapacityPlanner.PerformerCapacityPlan,
) error {
	p.logger.Sugar().Infow("Executing capacity plan",
		zap.String("avsAddress", p.avsConfig.AvsAddress),
		zap.Int("targetCount", plan.TargetCount),
	)

	healthyCount, err := p.CheckHealth(ctx)
	if err != nil {
		p.logger.Sugar().Errorw("Error checking health of performers",
			zap.String("avsAddress", p.avsConfig.AvsAddress),
			zap.Error(err),
		)
	}

	// Calculate the difference between healthy count and target
	diff := plan.TargetCount - healthyCount

	if diff > 0 {
		// Need to scale up
		return p.scaleUp(ctx, diff, plan.Artifact)
	} else if diff < 0 {
		// Need to scale down
		return p.scaleDown(ctx, -diff)
	}

	// No changes needed
	return nil
}

// createPerformer creates a new performer instance
func (p *PerformerPool) createPerformer(
	ctx context.Context,
	performerId string,
	artifact *performerCapacityPlanner.ArtifactVersion,
) error {
	// Check if performer already exists
	if _, exists := p.performers[performerId]; exists {
		return fmt.Errorf("performer with ID %s already exists", performerId)
	}

	// If we have a container manager and a valid artifact, pull the container first
	imageRegistry := p.avsConfig.Image.Registry
	imageTag := p.avsConfig.Image.Tag
	imageDigest := p.avsConfig.Image.Digest

	if artifact != nil && artifact.RegistryUrl != "" && (artifact.Digest != "" || artifact.Tag != "") {
		p.logger.Sugar().Infow("Pulling container for performer",
			"performerId", performerId,
			"registryUrl", artifact.RegistryUrl,
			"digest", artifact.Digest,
			"tag", artifact.Tag,
		)

		pullResult, err := p.containerMgr.PullContainer(ctx, artifact)
		if err != nil {
			p.logger.Sugar().Errorw("Failed to pull container for performer",
				"performerId", performerId,
				"registryUrl", artifact.RegistryUrl,
				"digest", artifact.Digest,
				"tag", artifact.Tag,
				"error", err,
			)
			return fmt.Errorf("failed to pull container: %w", err)
		}

		// If artifact has a tag, use it; otherwise fall back to config
		if artifact.Tag != "" {
			imageTag = artifact.Tag
		}

		p.logger.Sugar().Infow("Successfully pulled container for performer",
			"performerId", performerId,
			"imageId", pullResult.ImageID,
		)
	}

	performer, err := serverPerformer.NewAvsPerformerServer(
		&avsPerformer.AvsPerformerConfig{
			AvsAddress:           p.avsConfig.AvsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Registry: imageRegistry, Tag: imageTag, Digest: imageDigest},
			PerformerNetworkName: p.networkName,
			SigningCurve:         p.avsConfig.SigningCurve,
		},
		p.peeringFetcher,
		p.logger,
	)
	if err != nil {
		p.logger.Sugar().Errorw("Failed to create AVS performer server",
			zap.String("avsAddress", p.avsConfig.AvsAddress),
			zap.String("performerId", performerId),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create AVS performer server: %v", err)
	}

	if err := performer.Initialize(ctx); err != nil {
		p.logger.Sugar().Errorw("Failed to initialize performer",
			zap.String("avsAddress", p.avsConfig.AvsAddress),
			zap.String("performerId", performerId),
			zap.Error(err),
		)
		return err
	}

	p.performers[performerId] = performer

	// Store performer health status with performer Id
	p.statusMutex.Lock()
	p.healthStatus[performerId] = &PerformerHealth{
		ContainerId: performer.GetContainerId(),
		Healthy:     true,
		LastChecked: time.Now(),
	}
	p.statusMutex.Unlock()

	return nil
}

// CheckHealth checks health of all performers in the pool and returns the count of healthy performers
func (p *PerformerPool) CheckHealth(ctx context.Context) (int, error) {
	p.logger.Sugar().Debugw("Checking health of performers",
		zap.String("avsAddress", p.avsConfig.AvsAddress),
		zap.Int("performerCount", len(p.performers)),
	)

	// Track performers that need to be removed
	var performersToRemove []string
	healthyCount := 0

	// Check health of all existing performers
	for id, performer := range p.performers {
		// First check Docker container health
		dockerHealthy, dockerFailingStreak := p.getDockerHealthStatus(ctx, id)

		performerHealthy, err := p.checkPerformerHealth(ctx, id, performer)

		p.statusMutex.Lock()
		health := p.healthStatus[id]
		if health == nil {
			health = &PerformerHealth{}
			p.healthStatus[id] = health
		}

		health.LastChecked = time.Now()

		// Consider both Docker health and performer health
		healthy := dockerHealthy && performerHealthy
		health.Healthy = healthy

		if healthy {
			health.FailureCount = 0
			health.LastError = nil
			healthyCount++
		} else {
			health.FailureCount++
			health.LastError = err
			p.logger.Sugar().Warnw("Performer unhealthy",
				zap.String("avsAddress", p.avsConfig.AvsAddress),
				zap.String("performerId", id),
				zap.Int("failureCount", health.FailureCount),
				zap.Int("dockerFailingStreak", dockerFailingStreak),
				zap.Bool("dockerHealthy", dockerHealthy),
				zap.Bool("performerHealthy", performerHealthy),
				zap.Error(err),
			)

			// Mark for removal if failed too many times
			// Consider both our failure count and Docker's failing streak
			// Hard limit of 3 failures for now, could be configurable in the future
			if health.FailureCount >= 3 || dockerFailingStreak >= 3 {
				p.logger.Sugar().Warnw("Marking performer for removal due to repeated failures",
					zap.String("avsAddress", p.avsConfig.AvsAddress),
					zap.String("performerId", id),
					zap.Int("failureCount", health.FailureCount),
					zap.Int("dockerFailingStreak", dockerFailingStreak),
				)
				performersToRemove = append(performersToRemove, id)
			}
		}
		p.statusMutex.Unlock()
	}

	// Remove unhealthy performers
	for _, id := range performersToRemove {
		if err := p.removePerformer(id); err != nil {
			p.logger.Sugar().Errorw("Failed to remove unhealthy performer",
				zap.String("avsAddress", p.avsConfig.AvsAddress),
				zap.String("performerId", id),
				zap.Error(err),
			)
			// Continue with other removals even if this one failed
		}
	}

	return healthyCount, nil
}

// getDockerHealthStatus checks Docker container health and returns health status and failing streak
func (p *PerformerPool) getDockerHealthStatus(ctx context.Context, performerId string) (bool, int) {
	p.statusMutex.RLock()
	health, ok := p.healthStatus[performerId]
	p.statusMutex.RUnlock()

	if !ok || health.ContainerId == "" {
		return false, 0
	}

	if p.dockerClient == nil {
		return false, 0
	}

	inspection, err := p.dockerClient.ContainerInspect(ctx, health.ContainerId)
	if err != nil {
		p.logger.Sugar().Warnw("Failed to inspect container for health status",
			zap.String("avsAddress", p.avsConfig.AvsAddress),
			zap.String("performerId", performerId),
			zap.String("containerId", health.ContainerId),
			zap.Error(err),
		)
		return false, 0
	}

	if !inspection.State.Running {
		return false, 0
	}

	if inspection.State.Health != nil {
		healthy := inspection.State.Health.Status == container.Healthy ||
			inspection.State.Health.Status == container.Starting
		return healthy, inspection.State.Health.FailingStreak
	}

	// No health check configured, consider it healthy if running
	return true, 0
}

// checkPerformerHealth checks if a specific performer is healthy via gRPC
func (p *PerformerPool) checkPerformerHealth(ctx context.Context, performerId string, performer avsPerformer.IAvsPerformer) (bool, error) {
	// Check via the performer's health check (RunTask)
	taskCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dummyTask := &performerTask.PerformerTask{
		TaskID:  "health-check",
		Avs:     p.avsConfig.AvsAddress,
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
			zap.String("avsAddress", p.avsConfig.AvsAddress),
			zap.String("containerId", containerId),
			zap.Error(err),
		)
		return false
	}

	// First check if container is running
	if !inspection.State.Running {
		p.logger.Sugar().Warnw("Container is not running",
			zap.String("avsAddress", p.avsConfig.AvsAddress),
			zap.String("containerId", containerId),
			zap.String("state", inspection.State.Status),
		)
		return false
	}

	// Check health status if available
	if inspection.State.Health != nil {
		p.logger.Sugar().Debugw("Container health status",
			zap.String("avsAddress", p.avsConfig.AvsAddress),
			zap.String("containerId", containerId),
			zap.String("healthStatus", inspection.State.Health.Status),
			zap.Int("failingStreak", inspection.State.Health.FailingStreak),
		)

		switch inspection.State.Health.Status {
		case container.Healthy:
			return true
		case container.Starting:
			// Container is still starting, consider it healthy for now
			p.logger.Sugar().Debugw("Container is still starting",
				zap.String("avsAddress", p.avsConfig.AvsAddress),
				zap.String("containerId", containerId),
			)
			return true
		case container.Unhealthy:
			p.logger.Sugar().Warnw("Container is unhealthy",
				zap.String("avsAddress", p.avsConfig.AvsAddress),
				zap.String("containerId", containerId),
				zap.Int("failingStreak", inspection.State.Health.FailingStreak),
			)
			return false
		default:
			// Unknown health status, log and consider unhealthy
			p.logger.Sugar().Warnw("Unknown container health status",
				zap.String("avsAddress", p.avsConfig.AvsAddress),
				zap.String("containerId", containerId),
				zap.String("healthStatus", inspection.State.Health.Status),
			)
			return false
		}
	}

	return true
}

// scaleUp creates new performers to reach the target count
func (p *PerformerPool) scaleUp(ctx context.Context, countToAdd int, artifact *performerCapacityPlanner.ArtifactVersion) error {
	p.logger.Sugar().Infow("Scaling up performers",
		zap.String("avsAddress", p.avsConfig.AvsAddress),
		zap.Int("currentCount", len(p.performers)),
		zap.Int("countToAdd", countToAdd),
	)

	// Create the required number of performers
	for i := 0; i < countToAdd; i++ {
		performerId := fmt.Sprintf("performer-%d", len(p.performers)+1)

		// TODO: make this async as pulling a container can block other callers.
		if err := p.createPerformer(ctx, performerId, artifact); err != nil {
			p.logger.Sugar().Errorw("Failed to create performer during scale up",
				zap.String("avsAddress", p.avsConfig.AvsAddress),
				zap.String("performerId", performerId),
				zap.Error(err),
			)
			return err
		}

		p.logger.Sugar().Infow("Created new performer",
			zap.String("avsAddress", p.avsConfig.AvsAddress),
			zap.String("performerId", performerId),
		)
	}

	return nil
}

// scaleDown removes performers to reach the target count
func (p *PerformerPool) scaleDown(ctx context.Context, countToRemove int) error {
	p.logger.Sugar().Infow("Scaling down performers",
		zap.String("avsAddress", p.avsConfig.AvsAddress),
		zap.Int("currentCount", len(p.performers)),
		zap.Int("countToRemove", countToRemove),
	)

	// Get list of performers to remove, prioritizing unhealthy performers first
	performersToRemove := p.selectPerformersToRemove(countToRemove)

	// Remove the identified performers
	for _, id := range performersToRemove {
		if err := p.removePerformer(id); err != nil {
			p.logger.Sugar().Errorw("Failed to remove performer",
				zap.String("avsAddress", p.avsConfig.AvsAddress),
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
			if slices.Contains(performersToRemove, id) {
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
func (p *PerformerPool) removePerformer(performerId string) error {
	p.logger.Sugar().Infow("Removing performer",
		zap.String("avsAddress", p.avsConfig.AvsAddress),
		zap.String("performerId", performerId),
	)

	performer, exists := p.performers[performerId]
	if !exists {
		return fmt.Errorf("performer %s does not exist", performerId)
	}

	if err := performer.Shutdown(); err != nil {
		p.logger.Sugar().Errorw("Failed to shutdown performer",
			zap.String("avsAddress", p.avsConfig.AvsAddress),
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
		zap.String("avsAddress", p.avsConfig.AvsAddress),
		zap.Int("performerCount", len(p.performers)),
	)

	var lastErr error
	for id, performer := range p.performers {
		if err := performer.Shutdown(); err != nil {
			p.logger.Sugar().Errorw("Failed to shutdown performer",
				zap.String("avsAddress", p.avsConfig.AvsAddress),
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

	for id, performer := range p.performers {
		health, ok := p.healthStatus[id]
		if ok && health.Healthy {
			return performer, true
		}
	}

	// TODO: return an error if no healthy performers are available
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
	return p.avsConfig.AvsAddress
}
