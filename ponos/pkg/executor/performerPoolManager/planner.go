package performerPoolManager

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// ArtifactVersion represents a published artifact version for an AVS
type ArtifactVersion struct {
	AvsAddress    string
	OperatorSetId string
	Digest        string
	PublishedAt   uint64 // Block number
}

// PerformerCapacityPlan represents the desired state for performers of an AVS
type PerformerCapacityPlan struct {
	// The desired number of performers
	TargetCount int

	// The digest/version this capacity plan applies to (optional)
	Digest string

	// The latest artifact version for this AVS
	LatestArtifact *ArtifactVersion
}

// ChainPollerConfig represents the configuration for a chain poller
type ChainPollerConfig struct {
	ChainId             uint
	PollIntervalSeconds int
}

// PerformerCapacityPlanner determines the desired capacity for each AVS
type PerformerCapacityPlanner struct {
	logger          *zap.Logger
	avsConfigs      map[string]*executorConfig.AvsPerformerConfig
	mutex           sync.RWMutex
	chainEventsChan chan *chainPoller.LogWithBlock

	// Tracks which AVSs the executor is registered for
	registeredAvs map[string]bool

	// Tracks operator sets for each AVS
	operatorSets map[string][]string

	// Tracks latest artifact versions per AVS and operator set
	artifactVersions map[string]map[string]*ArtifactVersion

	// Chain poller for indexing events
	chainPoller chainPoller.IChainPoller

	// Contract store for accessing contract addresses
	contractStore contractStore.IContractStore

	// Chain contract caller for interacting with contracts
	chainContractCaller contractCaller.IContractCaller

	// Operator address to determine which artifacts are relevant
	operatorAddress string
}

// NewPerformerCapacityPlanner creates a new capacity planner
func NewPerformerCapacityPlanner(
	logger *zap.Logger,
	operatorAddress string,
	contractStore contractStore.IContractStore,
	chainContractCaller contractCaller.IContractCaller,
) *PerformerCapacityPlanner {
	return &PerformerCapacityPlanner{
		logger:              logger,
		avsConfigs:          make(map[string]*executorConfig.AvsPerformerConfig),
		mutex:               sync.RWMutex{},
		registeredAvs:       make(map[string]bool),
		operatorSets:        make(map[string][]string),
		artifactVersions:    make(map[string]map[string]*ArtifactVersion),
		contractStore:       contractStore,
		chainContractCaller: chainContractCaller,
		operatorAddress:     strings.ToLower(operatorAddress),
		chainEventsChan:     make(chan *chainPoller.LogWithBlock, 10000),
	}
}

// GetCapacityPlan returns a capacity plan for the given AVS
func (p *PerformerCapacityPlanner) GetCapacityPlan(avsAddress string) *PerformerCapacityPlan {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	avsAddress = strings.ToLower(avsAddress)

	// Check if we're registered for this AVS
	if registered, exists := p.registeredAvs[avsAddress]; exists && !registered {
		// If explicitly deregistered, return zero target count
		p.logger.Sugar().Infow("Executor is deregistered for AVS, scaling to zero",
			zap.String("avsAddress", avsAddress))

		return &PerformerCapacityPlan{
			TargetCount: 0,
			Digest:      "deregistered",
		}
	}

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

	// Check if we have operator set information
	operatorSetSize := 0
	if operators, exists := p.operatorSets[avsAddress]; exists {
		operatorSetSize = len(operators)

		// Log operator set information
		p.logger.Sugar().Debugw("Found operator set for AVS",
			zap.String("avsAddress", avsAddress),
			zap.Int("operatorCount", operatorSetSize),
		)
	}

	// Find the latest artifact for this AVS
	var latestArtifact *ArtifactVersion
	if artifactsBySet, exists := p.artifactVersions[avsAddress]; exists {
		for _, artifact := range artifactsBySet {
			if latestArtifact == nil || artifact.PublishedAt > latestArtifact.PublishedAt {
				latestArtifact = artifact
			}
		}
	}

	// Create digest information including artifact version if available
	digest := fmt.Sprintf("%s-ops%d", config.Image.Repository, operatorSetSize)
	if latestArtifact != nil {
		digest = fmt.Sprintf("%s-artifact-%s", digest, latestArtifact.Digest)
	}

	return &PerformerCapacityPlan{
		TargetCount:    targetCount,
		Digest:         digest,
		LatestArtifact: latestArtifact,
	}
}

// RegisterAVS registers an AVS with the planner
func (p *PerformerCapacityPlanner) RegisterAVS(
	avsAddress string,
	config *executorConfig.AvsPerformerConfig,
) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	avsAddress = strings.ToLower(avsAddress)
	p.avsConfigs[avsAddress] = config
	p.registeredAvs[avsAddress] = true

	// Initialize artifact versions map for this AVS
	if _, exists := p.artifactVersions[avsAddress]; !exists {
		p.artifactVersions[avsAddress] = make(map[string]*ArtifactVersion)
	}

	p.logger.Sugar().Infow("Registered AVS with planner",
		"avsAddress", avsAddress,
		"config", config,
	)
}

// SetChainEventsChan sets the channel for receiving chain events
func (p *PerformerCapacityPlanner) SetChainEventsChan(eventsChan chan *chainPoller.LogWithBlock) {
	p.chainEventsChan = eventsChan
}

// Start begins processing chain events in the background
func (p *PerformerCapacityPlanner) Start(ctx context.Context) {
	p.logger.Sugar().Infow("Starting capacity planner event processor")

	go p.processEvents(ctx)
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
	case "OperatorSetCreated":
		return p.handleOperatorSetCreated(logEvent)
	case "OperatorAddedToOperatorSet":
		return p.handleOperatorAddedToSet(logEvent)
	case "OperatorRemovedFromOperatorSet":
		return p.handleOperatorRemovedFromSet(logEvent)
	case "ExecutorRegistered":
		return p.handleExecutorRegistered(logEvent)
	case "ExecutorDeregistered":
		return p.handleExecutorDeregistered(logEvent)
	default:
		// Ignore logs from other contracts
		return nil
	}
}

// handlePublishedArtifact processes the PublishedArtifact event
func (p *PerformerCapacityPlanner) handlePublishedArtifact(event *chainPoller.LogWithBlock) error {
	logEvent := event.Log

	// Extract relevant fields from the event
	var avsAddress, operatorSetId, digest string

	// Parse event arguments
	for _, arg := range logEvent.Arguments {
		switch arg.Name {
		case "avs":
			if val, ok := arg.Value.(string); ok {
				avsAddress = strings.ToLower(val)
			}
		case "operatorSetId":
			if val, ok := arg.Value.(string); ok {
				operatorSetId = val
			}
		}
	}

	// Extract artifact digest from newArtifact
	for _, arg := range logEvent.Arguments {
		if arg.Name == "newArtifact" {
			if artifactMap, ok := arg.Value.(map[string]interface{}); ok {
				if digestVal, ok := artifactMap["digest"]; ok {
					if digestStr, ok := digestVal.(string); ok {
						digest = digestStr
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

	// Create and store the artifact version
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Initialize artifact versions map for this AVS if needed
	if _, exists := p.artifactVersions[avsAddress]; !exists {
		p.artifactVersions[avsAddress] = make(map[string]*ArtifactVersion)
	}

	// Store the new artifact version
	p.artifactVersions[avsAddress][operatorSetId] = &ArtifactVersion{
		AvsAddress:    avsAddress,
		OperatorSetId: operatorSetId,
		Digest:        digest,
		PublishedAt:   event.Block.Number.Value(),
	}

	p.logger.Sugar().Infow("Stored new artifact version",
		zap.String("avsAddress", avsAddress),
		zap.String("operatorSetId", operatorSetId),
		zap.String("digest", digest),
		zap.Uint64("blockNumber", event.Block.Number.Value()),
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
	members, err := p.chainContractCaller.GetOperatorSetMembers(avsAddress, operatorSetIdInt)
	if err != nil {
		return false, fmt.Errorf("failed to get operator set members: %w", err)
	}

	// Check if our operator address is in the set
	for _, member := range members {
		if strings.ToLower(member) == p.operatorAddress {
			return true, nil
		}
	}

	return false, nil
}

// handleOperatorSetCreated handles the OperatorSetCreated event
func (p *PerformerCapacityPlanner) handleOperatorSetCreated(logEvent *log.DecodedLog) error {
	// Extract AvsAddress and initialize empty operator set
	var avsAddress string
	for _, arg := range logEvent.Arguments {
		if arg.Name == "avsAddress" {
			if strValue, ok := arg.Value.(string); ok {
				avsAddress = strings.ToLower(strValue)
			}
		}
	}

	if avsAddress == "" {
		p.logger.Sugar().Warnw("Invalid OperatorSetCreated event, missing avsAddress")
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Initialize an empty operator set for this AVS
	p.operatorSets[avsAddress] = []string{}

	p.logger.Sugar().Infow("Created new operator set",
		zap.String("avsAddress", avsAddress),
	)

	return nil
}

// handleOperatorAddedToSet handles the OperatorAddedToOperatorSet event
func (p *PerformerCapacityPlanner) handleOperatorAddedToSet(logEvent *log.DecodedLog) error {
	var avsAddress, operatorAddress string
	for _, arg := range logEvent.Arguments {
		if arg.Name == "avsAddress" {
			if strValue, ok := arg.Value.(string); ok {
				avsAddress = strings.ToLower(strValue)
			}
		} else if arg.Name == "operatorAddress" {
			if strValue, ok := arg.Value.(string); ok {
				operatorAddress = strings.ToLower(strValue)
			}
		}
	}

	if avsAddress == "" || operatorAddress == "" {
		p.logger.Sugar().Warnw("Invalid OperatorAddedToOperatorSet event, missing required fields",
			zap.String("avsAddress", avsAddress),
			zap.String("operatorAddress", operatorAddress),
		)
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Initialize the operator set if it doesn't exist
	if _, exists := p.operatorSets[avsAddress]; !exists {
		p.operatorSets[avsAddress] = []string{}
	}

	// Check if operator is already in the set
	for _, op := range p.operatorSets[avsAddress] {
		if op == operatorAddress {
			// Already in the set, nothing to do
			return nil
		}
	}

	// Add the operator to the set
	p.operatorSets[avsAddress] = append(p.operatorSets[avsAddress], operatorAddress)

	p.logger.Sugar().Infow("Added operator to set",
		zap.String("avsAddress", avsAddress),
		zap.String("operatorAddress", operatorAddress),
		zap.Int("operatorCount", len(p.operatorSets[avsAddress])),
	)

	return nil
}

// handleOperatorRemovedFromSet handles the OperatorRemovedFromOperatorSet event
func (p *PerformerCapacityPlanner) handleOperatorRemovedFromSet(logEvent *log.DecodedLog) error {
	var avsAddress, operatorAddress string
	for _, arg := range logEvent.Arguments {
		if arg.Name == "avsAddress" {
			if strValue, ok := arg.Value.(string); ok {
				avsAddress = strings.ToLower(strValue)
			}
		} else if arg.Name == "operatorAddress" {
			if strValue, ok := arg.Value.(string); ok {
				operatorAddress = strings.ToLower(strValue)
			}
		}
	}

	if avsAddress == "" || operatorAddress == "" {
		p.logger.Sugar().Warnw("Invalid OperatorRemovedFromOperatorSet event, missing required fields",
			zap.String("avsAddress", avsAddress),
			zap.String("operatorAddress", operatorAddress),
		)
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if the operator set exists
	operators, exists := p.operatorSets[avsAddress]
	if !exists {
		p.logger.Sugar().Warnw("Operator set does not exist for AVS",
			zap.String("avsAddress", avsAddress),
		)
		return nil
	}

	// Remove the operator from the set
	var updatedOperators []string
	for _, op := range operators {
		if op != operatorAddress {
			updatedOperators = append(updatedOperators, op)
		}
	}

	p.operatorSets[avsAddress] = updatedOperators

	p.logger.Sugar().Infow("Removed operator from set",
		zap.String("avsAddress", avsAddress),
		zap.String("operatorAddress", operatorAddress),
		zap.Int("operatorCount", len(p.operatorSets[avsAddress])),
	)

	return nil
}

// handleExecutorRegistered handles the ExecutorRegistered event
func (p *PerformerCapacityPlanner) handleExecutorRegistered(logEvent *log.DecodedLog) error {
	var avsAddress string
	for _, arg := range logEvent.Arguments {
		if arg.Name == "avsAddress" {
			if strValue, ok := arg.Value.(string); ok {
				avsAddress = strings.ToLower(strValue)
			}
		}
	}

	if avsAddress == "" {
		p.logger.Sugar().Warnw("Invalid ExecutorRegistered event, missing avsAddress")
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.registeredAvs[avsAddress] = true

	p.logger.Sugar().Infow("Executor registered for AVS",
		zap.String("avsAddress", avsAddress),
	)

	return nil
}

// handleExecutorDeregistered handles the ExecutorDeregistered event
func (p *PerformerCapacityPlanner) handleExecutorDeregistered(logEvent *log.DecodedLog) error {
	var avsAddress string
	for _, arg := range logEvent.Arguments {
		if arg.Name == "avsAddress" {
			if strValue, ok := arg.Value.(string); ok {
				avsAddress = strings.ToLower(strValue)
			}
		}
	}

	if avsAddress == "" {
		p.logger.Sugar().Warnw("Invalid ExecutorDeregistered event, missing avsAddress")
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.registeredAvs[avsAddress] = false

	p.logger.Sugar().Infow("Executor deregistered for AVS",
		zap.String("avsAddress", avsAddress),
	)

	return nil
}
