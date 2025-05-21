package performerPoolManager

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

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

	// Components
	pools   map[string]*PerformerPool
	planner *PerformerCapacityPlanner

	// Lifecycle management
	poolsMutex sync.RWMutex
}

// NewPerformerPoolManager creates a new performer pool manager
func NewPerformerPoolManager(
	config *executorConfig.ExecutorConfig,
	logger *zap.Logger,
	peeringFetcher peering.IPeeringDataFetcher,
) *PerformerPoolManager {
	return &PerformerPoolManager{
		logger:         logger,
		config:         config,
		peeringFetcher: peeringFetcher,
		pools:          make(map[string]*PerformerPool),
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

	// Initialize capacity planner
	p.planner = NewPerformerCapacityPlanner(p.logger)

	// Create performer pools for each AVS
	for _, avsConfig := range p.config.AvsPerformers {
		avsAddress := strings.ToLower(avsConfig.AvsAddress)

		// Skip if unsupported type
		if avsConfig.ProcessType != string(avsPerformer.AvsProcessTypeServer) {
			p.logger.Sugar().Warnw("Unsupported AVS performer process type, skipping",
				zap.String("avsAddress", avsAddress),
				zap.String("processType", avsConfig.ProcessType),
			)
			continue
		}

		// Register with capacity planner
		p.planner.RegisterAVS(avsAddress, avsConfig)

		// Create pool
		pool := NewPerformerPool(
			avsConfig,
			p.config.PerformerNetworkName,
			p.dockerClient,
			p.logger,
			p.peeringFetcher,
		)

		p.pools[avsAddress] = pool
	}

	return nil
}

// BootPerformers starts all performer pools and begins lifecycle management
func (p *PerformerPoolManager) BootPerformers(ctx context.Context) error {
	p.logger.Sugar().Infow("Booting AVS performers")

	// Initialize pools with primary performer
	for avsAddress, pool := range p.pools {
		if err := pool.createPerformer(ctx, "primary"); err != nil {
			p.logger.Sugar().Errorw("Failed to initialize performer pool",
				zap.String("avsAddress", avsAddress),
				zap.Error(err),
			)
			return fmt.Errorf("failed to initialize performer pool for %s: %v", avsAddress, err)
		}
	}

	// Start lifecycle management in background
	go p.startLifecycleManagement(ctx)

	// Set up cleanup on context done
	go func() {
		<-ctx.Done()
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
	}()

	return nil
}

// startLifecycleManagement runs a loop to maintain performers
func (p *PerformerPoolManager) startLifecycleManagement(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Initial check immediately
	p.performLifecycleCheck(ctx)

	for {
		select {
		case <-ticker.C:
			p.performLifecycleCheck(ctx)
		case <-ctx.Done():
			p.logger.Sugar().Info("Context done, stopping performer lifecycle management")
			return
		}
	}
}

// performLifecycleCheck checks all performers and maintains desired state
func (p *PerformerPoolManager) performLifecycleCheck(ctx context.Context) {
	p.logger.Sugar().Debugw("Performing performer lifecycle check")

	// Check and update each pool
	p.poolsMutex.RLock()
	defer p.poolsMutex.RUnlock()

	for avsAddress, pool := range p.pools {
		// Get capacity plan for this AVS
		plan := p.planner.GetCapacityPlan(avsAddress)

		// Execute plan (this will handle health checks internally)
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
