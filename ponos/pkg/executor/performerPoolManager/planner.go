package performerPoolManager

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"go.uber.org/zap"
)

// PerformerCapacityPlan represents the desired state for performers of an AVS
type PerformerCapacityPlan struct {
	// The desired number of performers
	TargetCount int

	// The digest/version this capacity plan applies to (optional)
	Digest string
}

// PerformerCapacityPlanner determines the desired capacity for each AVS
type PerformerCapacityPlanner struct {
	logger     *zap.Logger
	avsConfigs map[string]*executorConfig.AvsPerformerConfig
}

// NewPerformerCapacityPlanner creates a new capacity planner
func NewPerformerCapacityPlanner(
	logger *zap.Logger,
) *PerformerCapacityPlanner {
	return &PerformerCapacityPlanner{
		logger:     logger,
		avsConfigs: make(map[string]*executorConfig.AvsPerformerConfig),
	}
}

// RegisterAVS registers an AVS with the planner
func (p *PerformerCapacityPlanner) RegisterAVS(
	avsAddress string,
	config *executorConfig.AvsPerformerConfig,
) {
	p.avsConfigs[avsAddress] = config
}

// GetCapacityPlan returns a capacity plan for the given AVS
func (p *PerformerCapacityPlanner) GetCapacityPlan(
	avsAddress string,
) *PerformerCapacityPlan {
	// Get configuration for this AVS
	config, ok := p.avsConfigs[avsAddress]
	if !ok {
		// If no configuration, default to 1 performer
		p.logger.Sugar().Warnw("No configuration found for AVS, using default",
			zap.String("avsAddress", avsAddress))

		return &PerformerCapacityPlan{
			TargetCount: 1,
			Digest:      "default",
		}
	}

	// Use workerCount from config as target
	targetCount := config.WorkerCount
	if targetCount <= 0 {
		p.logger.Sugar().Warnw("No worker count found for AVS, using default",
			zap.String("avsAddress", avsAddress))
		targetCount = 1
	}

	return &PerformerCapacityPlan{
		TargetCount: targetCount,
		// TODO: Use image digest
		Digest: p.avsConfigs[avsAddress].Image.Repository,
	}
}
