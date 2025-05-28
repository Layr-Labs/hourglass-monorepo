package performerPoolManager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerCapacityPlanner"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// PerformerPoolManager is responsible for managing the lifecycle of performer containers
// for multiple AVSs
type PerformerPoolManager struct {
	logger         *zap.Logger
	config         *executorConfig.ExecutorConfig
	peeringFetcher peering.IPeeringDataFetcher
	dockerClient   *client.Client

	pools        map[string]*PerformerPool
	planner      *performerCapacityPlanner.PerformerCapacityPlanner
	containerMgr containerManager.IContainerManager

	poolsMutex sync.RWMutex
}

// NewPerformerPoolManager creates a new performer pool manager
func NewPerformerPoolManager(
	config *executorConfig.ExecutorConfig,
	logger *zap.Logger,
	peeringFetcher peering.IPeeringDataFetcher,
	planner *performerCapacityPlanner.PerformerCapacityPlanner,
	containerMgr containerManager.IContainerManager,
) *PerformerPoolManager {
	return &PerformerPoolManager{
		logger:         logger,
		config:         config,
		peeringFetcher: peeringFetcher,
		pools:          make(map[string]*PerformerPool),
		planner:        planner,
		containerMgr:   containerMgr,
	}
}

// Initialize initializes the pool manager and all AVS performer pools
func (p *PerformerPoolManager) Initialize() error {
	p.logger.Sugar().Infow("Initializing PerformerPoolManager")

	// Initialize Docker client for container management
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		p.logger.Sugar().Errorw("Failed to create Docker client", zap.Error(err))
		return fmt.Errorf("failed to create Docker client: %v", err)
	}

	p.dockerClient = dockerClient

	// Create performer pools for each AVS
	for _, avsConfig := range p.config.AvsPerformers {
		avsAddress := avsConfig.AvsAddress

		// Create pool
		pool := NewPerformerPool(
			p.config.PerformerNetworkName,
			avsConfig,
			p.dockerClient,
			p.logger,
			p.peeringFetcher,
			p.containerMgr,
		)

		p.pools[avsAddress] = pool
	}

	return nil
}

// Start starts all performer pools and begins lifecycle management
func (p *PerformerPoolManager) Start(ctx context.Context) error {
	p.logger.Sugar().Infow("Starting performer pool manager")

	// Start lifecycle management in background
	go p.startLifecycleManagement(ctx)

	return nil
}

// startLifecycleManagement runs a loop to maintain performers
func (p *PerformerPoolManager) startLifecycleManagement(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Initial check immediately
	p.performLifecycleManagement(ctx)

	for {
		select {
		case <-ticker.C:
			p.performLifecycleManagement(ctx)
		case <-ctx.Done():
			p.logger.Sugar().Info("Shutting down performer pool manager")

			// Shutdown all pools
			p.poolsMutex.RLock()
			defer p.poolsMutex.RUnlock()

			for avsAddress, pool := range p.pools {
				if err := pool.Shutdown(); err != nil {
					p.logger.Sugar().Errorw("Failed to shutdown performer pool",
						zap.String("avsAddress", avsAddress),
						zap.Error(err),
					)
				}
			}
			return
		}
	}
}

// performLifecycleManagement checks all performers and maintains desired state
func (p *PerformerPoolManager) performLifecycleManagement(ctx context.Context) {
	p.logger.Sugar().Debugw("Performing performer lifecycle check")

	// Check and update each pool
	p.poolsMutex.RLock()
	defer p.poolsMutex.RUnlock()

	for avsAddress, pool := range p.pools {
		// Get capacity plan for this AVS
		plan, err := p.planner.GetCapacityPlan(avsAddress)
		// TODO: if the plan is not found, we should tear down the pool by passing a targetCount 0 plan.
		if err != nil {
			p.logger.Sugar().Warnw("Failed to get capacity plan for AVS, skipping lifecycle check",
				zap.String("avsAddress", avsAddress),
				zap.Error(err),
			)
			continue
		}

		// Execute plan
		if err := pool.ExecutePlan(ctx, plan); err != nil {
			p.logger.Sugar().Errorw("Error executing capacity plan",
				zap.String("avsAddress", avsAddress),
				zap.Error(err),
			)
		}
	}
}

// GetPerformer returns a performer for the given AVS address
func (p *PerformerPoolManager) GetPerformer(avsAddress string) (avsPerformer.IAvsPerformer, bool) {
	p.poolsMutex.RLock()
	defer p.poolsMutex.RUnlock()

	pool, ok := p.pools[avsAddress]
	if !ok {
		return nil, false
	}

	return pool.GetHealthyPerformer()
}
