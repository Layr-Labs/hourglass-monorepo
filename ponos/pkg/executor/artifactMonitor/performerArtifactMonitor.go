package artifactMonitor

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"go.uber.org/zap"
	"sync"
)

// PerformerArtifactMonitor determines the desired capacity for each AVS
type PerformerArtifactMonitor struct {
	logger          *zap.Logger
	artifactsMutex  sync.RWMutex
	chainEventsChan chan *chainPoller.LogWithBlock

	// Map of AVS address to performer artifacts
	performerArtifacts map[string][]*PerformerArtifact

	// Contract store for accessing contract addresses
	contractStore contractStore.IContractStore

	// AVS performer configurations from startup
	avsPerformersConfig map[string]*executorConfig.AvsPerformerConfig
}

// NewPerformerArtifactMonitor creates a new capacity planner
func NewPerformerArtifactMonitor(
	logger *zap.Logger,
	contractStore contractStore.IContractStore,
	eventsChan chan *chainPoller.LogWithBlock,
	avsPerformersConfig []*executorConfig.AvsPerformerConfig,
) *PerformerArtifactMonitor {
	configMap := make(map[string]*executorConfig.AvsPerformerConfig)
	for _, config := range avsPerformersConfig {
		configMap[config.AvsAddress] = config
	}

	return &PerformerArtifactMonitor{
		logger:              logger,
		artifactsMutex:      sync.RWMutex{},
		performerArtifacts:  make(map[string][]*PerformerArtifact),
		contractStore:       contractStore,
		chainEventsChan:     eventsChan,
		avsPerformersConfig: configMap,
	}
}

// GetArtifacts returns artifacts expected to be run for the given AVS
func (p *PerformerArtifactMonitor) GetArtifacts(avsAddress string) ([]*PerformerArtifact, error) {
	p.artifactsMutex.RLock()
	defer p.artifactsMutex.RUnlock()

	// Check if we have artifacts for this AVS
	if artifacts, exists := p.performerArtifacts[avsAddress]; exists {
		return artifacts, nil
	}

	p.logger.Sugar().Warnw("No capacity artifacts found for AVS",
		zap.String("avsAddress", avsAddress))
	return nil, fmt.Errorf("no capacity artifacts exists for AVS %s", avsAddress)
}

// Start begins processing chain events in the background and periodically discovers operator sets
func (p *PerformerArtifactMonitor) Start(ctx context.Context) {
	p.logger.Sugar().Infow("Starting capacity planner event processor")

	// Start event processor
	go p.processEvents(ctx)
}

// processEvents continuously processes events from the chain events channel
func (p *PerformerArtifactMonitor) processEvents(ctx context.Context) {
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
func (p *PerformerArtifactMonitor) processEvent(event *chainPoller.LogWithBlock) error {
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
func (p *PerformerArtifactMonitor) handlePublishedArtifact(event *chainPoller.LogWithBlock) error {
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

	// Check if we have configuration for this AVS
	_, hasConfig := p.avsPerformersConfig[avsAddress]
	if !hasConfig {
		p.logger.Sugar().Infow("Ignoring artifact for AVS without configuration",
			zap.String("avsAddress", avsAddress),
			zap.String("operatorSetId", operatorSetId),
			zap.String("digest", digest),
		)
		return nil
	}

	// Create artifact version
	artifactVersion := &PerformerArtifact{
		AvsAddress:    avsAddress,
		OperatorSetId: operatorSetId,
		Digest:        digest,
		RegistryUrl:   registryUrl,
		PublishedAt:   event.Block.Number.Value(),
	}

	// Update capacity plan with the new artifact and config
	p.updateArtifacts(avsAddress, artifactVersion)

	p.logger.Sugar().Infow("Updated capacity plan with new artifact",
		"avsAddress", avsAddress,
		"operatorSetId", operatorSetId,
		"digest", digest,
		"registryUrl", registryUrl,
		"blockNumber", event.Block.Number.Value(),
	)

	return nil
}

func (p *PerformerArtifactMonitor) updateArtifacts(
	avsAddress string,
	artifactVersion *PerformerArtifact,
) {
	p.artifactsMutex.Lock()
	defer p.artifactsMutex.Unlock()

	// Retrieve the current list or initialize a new one
	artifacts, exists := p.performerArtifacts[avsAddress]
	if !exists {
		artifacts = []*PerformerArtifact{}
	}

	// Check for duplicates by comparing Digest and OperatorSetId
	for _, existing := range artifacts {
		if existing.Digest == artifactVersion.Digest && existing.AvsAddress == artifactVersion.AvsAddress {
			// Already exists, skip update
			p.logger.Sugar().Infow("Duplicate artifact, skipping",
				"avsAddress", avsAddress,
				"operatorSetId", artifactVersion.OperatorSetId,
				"digest", artifactVersion.Digest)
			return
		}
	}

	// Append the new artifact
	p.performerArtifacts[avsAddress] = append(artifacts, artifactVersion)

	p.logger.Sugar().Infow("Updated capacity plan",
		"avsAddress", avsAddress,
		"targetCount", len(p.performerArtifacts[avsAddress]),
		"digest", artifactVersion.Digest,
		"registryUrl", artifactVersion.RegistryUrl)
}
