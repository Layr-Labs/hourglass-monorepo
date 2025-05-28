package performerCapacityPlanner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// PerformerCapacityPlanner determines the desired capacity for each AVS
type PerformerCapacityPlanner struct {
	logger          *zap.Logger
	planMutex       sync.RWMutex
	chainEventsChan chan *chainPoller.LogWithBlock

	// Map of AVS address to capacity plan
	capacityPlans map[string]*PerformerCapacityPlan

	// Contract store for accessing contract addresses
	contractStore contractStore.IContractStore

	// Chain contract caller for interacting with contracts
	chainContractCaller contractCaller.IContractCaller

	// Operator address to determine which artifacts are relevant
	operatorAddress string

	// AVS performer configurations from startup
	avsPerformersConfig map[string]*avsPerformer.AvsPerformerConfig
}

// NewPerformerCapacityPlanner creates a new capacity planner
func NewPerformerCapacityPlanner(
	logger *zap.Logger,
	operatorAddress string,
	contractStore contractStore.IContractStore,
	chainContractCaller contractCaller.IContractCaller,
	eventsChan chan *chainPoller.LogWithBlock,
	avsPerformersConfig []*avsPerformer.AvsPerformerConfig,
) *PerformerCapacityPlanner {
	configMap := make(map[string]*avsPerformer.AvsPerformerConfig)
	for _, config := range avsPerformersConfig {
		configMap[config.AvsAddress] = config
	}

	return &PerformerCapacityPlanner{
		logger:              logger,
		planMutex:           sync.RWMutex{},
		capacityPlans:       make(map[string]*PerformerCapacityPlan),
		contractStore:       contractStore,
		chainContractCaller: chainContractCaller,
		operatorAddress:     operatorAddress,
		chainEventsChan:     eventsChan,
		avsPerformersConfig: configMap,
	}
}

// GetCapacityPlan returns a capacity plan for the given AVS
func (p *PerformerCapacityPlanner) GetCapacityPlan(avsAddress string) (*PerformerCapacityPlan, error) {
	p.planMutex.RLock()
	defer p.planMutex.RUnlock()

	// Check if we have a capacity plan for this AVS
	if plan, exists := p.capacityPlans[avsAddress]; exists {
		return plan, nil
	}

	// No plan exists, return error
	// TODO: if no plan for the AVS, return a targetCount 0 plan with otherwise empty fields.
	p.logger.Sugar().Warnw("No capacity plan found for AVS",
		zap.String("avsAddress", avsAddress))
	return nil, fmt.Errorf("no capacity plan exists for AVS %s", avsAddress)
}

// Start begins processing chain events in the background and periodically discovers operator sets
func (p *PerformerCapacityPlanner) Start(ctx context.Context) {
	p.logger.Sugar().Infow("Starting capacity planner event processor")

	// Start event processor
	go p.processEvents(ctx)

	// Start operator set discovery routine
	go p.discoverOperatorSets(ctx)
}

// processEvents continuously processes events from the chain events channel
func (p *PerformerCapacityPlanner) processEvents(ctx context.Context) {
	if p.chainEventsChan == nil {
		p.logger.Sugar().Warnw("Chain events channel not set, skipping event processing")
		return
	}

	for {
		select {
		case <-ctx.Done():
			p.logger.Sugar().Infow("Context cancelled, stopping capacity planner event processor")
			return
		case event, ok := <-p.chainEventsChan:
			if !ok {
				p.logger.Sugar().Warnw("Chain events channel closed, stopping capacity planner event processor")
				return
			}

			if err := p.processEvent(event); err != nil {
				p.logger.Sugar().Errorw("Error processing event", "error", err)
			}
		}
	}
}

// processEvent handles a single chain event
func (p *PerformerCapacityPlanner) processEvent(event *chainPoller.LogWithBlock) error {
	if event == nil || event.Log == nil {
		return nil
	}

	logEvent := event.Log
	p.logger.Sugar().Debugw("Processing AVS Artifact Registry event",
		zap.String("eventName", logEvent.EventName),
		zap.String("contractAddress", logEvent.Address),
	)

	// Process based on contract address
	switch logEvent.EventName {
	case "PublishedArtifact":
		return p.handlePublishedArtifact(event)
	default:
		// Ignore logs for other events
		return nil
	}
}

// handlePublishedArtifact processes the PublishedArtifact event
func (p *PerformerCapacityPlanner) handlePublishedArtifact(event *chainPoller.LogWithBlock) error {
	logEvent := event.Log

	// Extract relevant fields from the event
	var avsAddress, operatorSetId, digest, registryUrl string

	// Parse event arguments
	for _, arg := range logEvent.Arguments {
		switch arg.Name {
		case "avs":
			if val, ok := arg.Value.(string); ok {
				avsAddress = val
			}
		case "operatorSetId":
			if val, ok := arg.Value.(string); ok {
				operatorSetId = val
			}
		}
	}

	// Extract artifact digest and registryUrl from newArtifact
	for _, arg := range logEvent.Arguments {
		if arg.Name == "newArtifact" {
			if artifactMap, ok := arg.Value.(map[string]interface{}); ok {
				if digestVal, ok := artifactMap["digest"]; ok {
					if digestStr, ok := digestVal.(string); ok {
						digest = digestStr
					}
				}
				if registryUrlVal, ok := artifactMap["registryUrl"]; ok {
					if registryUrlStr, ok := registryUrlVal.(string); ok {
						registryUrl = registryUrlStr
					}
				}
			}
		}
	}

	if avsAddress == "" || operatorSetId == "" || digest == "" {
		p.logger.Sugar().Warnw("Invalid PublishedArtifact event, missing required fields",
			zap.String("avsAddress", avsAddress),
			zap.String("operatorSetId", operatorSetId),
			zap.String("digest", digest),
		)
		return fmt.Errorf("invalid PublishedArtifact event")
	}

	// Check if this operator is part of the operator set
	isRelevant, err := p.isArtifactRelevantToOperator(avsAddress, operatorSetId)
	if err != nil {
		p.logger.Sugar().Errorw("Failed to check if artifact is relevant",
			zap.String("avsAddress", avsAddress),
			zap.String("operatorSetId", operatorSetId),
			zap.Error(err),
		)
		return err
	}

	if !isRelevant {
		p.logger.Sugar().Infow("Ignoring artifact for operator set we're not part of",
			zap.String("avsAddress", avsAddress),
			zap.String("operatorSetId", operatorSetId),
		)
		return nil
	}

	// Check if we have configuration for this AVS
	performerConfig, hasConfig := p.avsPerformersConfig[avsAddress]
	if !hasConfig {
		p.logger.Sugar().Infow("Ignoring artifact for AVS without configuration",
			zap.String("avsAddress", avsAddress),
			zap.String("operatorSetId", operatorSetId),
			zap.String("digest", digest),
		)
		return nil
	}

	// Create artifact version
	artifactVersion := &ArtifactVersion{
		AvsAddress:    avsAddress,
		OperatorSetId: operatorSetId,
		Digest:        digest,
		RegistryUrl:   registryUrl,
		PublishedAt:   event.Block.Number.Value(),
	}

	// Update capacity plan with the new artifact and config
	p.updateCapacityPlanWithArtifact(avsAddress, artifactVersion, performerConfig)

	p.logger.Sugar().Infow("Updated capacity plan with new artifact",
		"avsAddress", avsAddress,
		"operatorSetId", operatorSetId,
		"digest", digest,
		"registryUrl", registryUrl,
		"blockNumber", event.Block.Number.Value(),
	)

	return nil
}

// isArtifactRelevantToOperator checks if the artifact is for an operator set that includes this operator
func (p *PerformerCapacityPlanner) isArtifactRelevantToOperator(avsAddress, operatorSetId string) (bool, error) {
	// Convert operatorSetId from hex string to uint32 if necessary
	var operatorSetIdInt uint32

	// Try to parse as integer first
	if _, err := fmt.Sscanf(operatorSetId, "%d", &operatorSetIdInt); err != nil {
		// If that fails, assume it's a hex string and convert to bytes
		operatorSetIdBytes := common.FromHex(operatorSetId)
		// Only use the first 4 bytes for uint32
		if len(operatorSetIdBytes) >= 4 {
			operatorSetIdInt = uint32(operatorSetIdBytes[0])<<24 |
				uint32(operatorSetIdBytes[1])<<16 |
				uint32(operatorSetIdBytes[2])<<8 |
				uint32(operatorSetIdBytes[3])
		}
	}

	// Get the members of this operator set
	// TODO: is there are race condition here? If I add an operator and immediately release an artifact, will this catch?
	members, err := p.chainContractCaller.GetOperatorSetMembers(avsAddress, operatorSetIdInt)
	if err != nil {
		return false, fmt.Errorf("failed to get operator set members: %w", err)
	}

	// Check if our operator address is in the set
	for _, member := range members {
		if strings.ToLower(member) == strings.ToLower(p.operatorAddress) {
			return true, nil
		}
	}

	return false, nil
}

// discoverOperatorSets periodically discovers and updates the operator sets this operator is registered with
func (p *PerformerCapacityPlanner) discoverOperatorSets(ctx context.Context) {
	if err := p.updateOperatorSets(); err != nil {
		p.logger.Sugar().Errorw("Failed to discover operator sets", "error", err)
	}

	// Then run periodically
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Sugar().Infow("Context cancelled, stopping operator set discovery")
			return
		case <-ticker.C:
			if err := p.updateOperatorSets(); err != nil {
				p.logger.Sugar().Errorw("Failed to discover operator sets", "error", err)
			}
		}
	}
}

// updateOperatorSets discovers and updates the operator sets this operator is registered with
func (p *PerformerCapacityPlanner) updateOperatorSets() error {
	p.logger.Sugar().Infow("Discovering operator sets for operator",
		"operatorAddress", p.operatorAddress)

	// Get all operator sets from AllocationManager that this operator is registered to
	operatorSets, err := p.chainContractCaller.GetRegisteredSets(p.operatorAddress)
	if err != nil {
		return fmt.Errorf("failed to get registered operator sets for operator %s: %w", p.operatorAddress, err)
	}

	p.logger.Sugar().Infow("Found operator sets", "count", len(operatorSets))

	// Track current AVSs to detect removals
	currentAvs := make(map[string]bool)

	// Process each operator set to identify relevant AVSs
	for _, operatorSet := range operatorSets {
		avsAddress := operatorSet.Avs.String()
		operatorSetId := operatorSet.Id

		// Query the TaskMailbox to determine if this is a relevant AVS
		avsConfig, err := p.chainContractCaller.GetAVSConfig(avsAddress)
		if err != nil {
			p.logger.Sugar().Warnw("Failed to get AVS config, skipping",
				"avsAddress", avsAddress,
				"error", err)
			continue
		}

		// Check if this operator set ID is an executor
		isExecutorSet := false
		for _, execOpSetId := range avsConfig.ExecutorOperatorSetIds {
			if execOpSetId == operatorSetId {
				isExecutorSet = true
				break
			}
		}

		if !isExecutorSet {
			continue
		}

		// Mark this as a current AVS
		currentAvs[avsAddress] = true

		// Check if we have configuration for this AVS
		performerConfig, hasConfig := p.avsPerformersConfig[avsAddress]
		if !hasConfig {
			p.logger.Sugar().Infow("Ignoring AVS without configuration",
				"avsAddress", avsAddress,
				"operatorSetId", operatorSetId)
			continue
		}

		// Get the latest artifact for this AVS and operator set
		operatorSetIdStr := fmt.Sprintf("%d", operatorSetId)
		artifact, err := p.chainContractCaller.GetLatestArtifact(avsAddress, operatorSetIdStr)
		if err != nil {
			p.logger.Sugar().Warnw("No latest artifact in artifact registry, using startup config",
				"avsAddress", avsAddress,
				"operatorSetId", operatorSetId,
				"error", err,
			)

			// Create artifact version with config values
			artifactVersion := &ArtifactVersion{
				AvsAddress:    avsAddress,
				OperatorSetId: operatorSetIdStr,
				Digest:        performerConfig.Image.Digest,
				Tag:           performerConfig.Image.Tag,
				RegistryUrl:   performerConfig.Image.Registry,
			}

			// Update capacity plan with config-based artifact
			p.updateCapacityPlanWithArtifact(avsAddress, artifactVersion, performerConfig)
		} else {
			// Create artifact version
			artifactVersion := &ArtifactVersion{
				AvsAddress:    avsAddress,
				OperatorSetId: operatorSetIdStr,
				Digest:        string(artifact.Digest),
				RegistryUrl:   string(artifact.RegistryUrl),
			}

			// Update capacity plan for this AVS
			p.updateCapacityPlanWithArtifact(avsAddress, artifactVersion, performerConfig)
		}
	}

	// Remove capacity plans for AVSs that are no longer registered
	p.planMutex.Lock()
	for avsAddress := range p.capacityPlans {
		if !currentAvs[avsAddress] {
			p.logger.Sugar().Infow("Operator no longer registered for AVS, removing capacity plan",
				"avsAddress", avsAddress)
			delete(p.capacityPlans, avsAddress)
		}
	}
	p.planMutex.Unlock()

	p.logger.Sugar().Infow("Operator set discovery complete",
		"registeredAvsCount", len(currentAvs))
	return nil
}

// updateCapacityPlanWithArtifact updates a capacity plan with the given artifact
func (p *PerformerCapacityPlanner) updateCapacityPlanWithArtifact(
	avsAddress string,
	artifactVersion *ArtifactVersion,
	performerConfig *avsPerformer.AvsPerformerConfig,
) {
	p.planMutex.Lock()
	defer p.planMutex.Unlock()

	// Create or update the plan
	p.capacityPlans[avsAddress] = &PerformerCapacityPlan{
		TargetCount: 1,
		Artifact:    artifactVersion,
	}

	p.logger.Sugar().Infow("Updated capacity plan",
		"avsAddress", avsAddress,
		"targetCount", 1,
		"digest", artifactVersion.Digest,
		"tag", performerConfig.Image.Tag,
		"registryUrl", artifactVersion.RegistryUrl,
	)
}
