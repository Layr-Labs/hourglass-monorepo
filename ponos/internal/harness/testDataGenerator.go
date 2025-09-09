package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"go.uber.org/zap"
)

// TestDataGenerator manages the generation of test states
type TestDataGenerator struct {
	config     *GeneratorConfig
	logger     *zap.SugaredLogger
	generators map[string]StateGenerator
	mu         sync.RWMutex
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator(config *GeneratorConfig) (*TestDataGenerator, error) {
	logConfig := &logger.LoggerConfig{Debug: config.Verbose}
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	
	return &TestDataGenerator{
		config:     config,
		logger:     log.Sugar(),
		generators: make(map[string]StateGenerator),
	}, nil
}

// RegisterGenerator registers a state generator for a specific test suite
func (g *TestDataGenerator) RegisterGenerator(testSuite string, generator StateGenerator) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.generators[testSuite] = generator
	g.logger.Infow("Registered state generator",
		"testSuite", testSuite,
		"stateID", generator.GetStateID(),
	)
}

// GenerateState generates state for a specific test suite
func (g *TestDataGenerator) GenerateState(ctx context.Context, testSuite string) (*GeneratedState, error) {
	g.mu.RLock()
	generator, exists := g.generators[testSuite]
	g.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no generator registered for test suite: %s", testSuite)
	}
	
	g.logger.Infow("Starting state generation",
		"testSuite", testSuite,
		"stateID", generator.GetStateID(),
		"description", generator.GetDescription(),
	)
	
	// Create output directory structure
	outputDir := filepath.Join(g.config.OutputDir, "states", testSuite, generator.GetStateID())
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Check if state already exists and skip if not overwriting
	metadataPath := filepath.Join(outputDir, "metadata.json")
	if !g.config.Overwrite && fileExists(metadataPath) {
		g.logger.Infow("State already exists, skipping generation",
			"testSuite", testSuite,
			"stateID", generator.GetStateID(),
		)
		return g.loadExistingState(outputDir)
	}
	
	// Start anvil instances
	l1Anvil, l2Anvil, err := g.startAnvilInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start anvil instances: %w", err)
	}
	defer l1Anvil.Stop()
	defer l2Anvil.Stop()
	
	// Load chain configuration
	chainConfig, err := testUtils.ReadChainConfig(filepath.Dir(g.config.ChainConfigPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read chain config: %w", err)
	}
	
	// Create contract callers
	l1Caller, l2Caller, err := g.createContractCallers(ctx, l1Anvil, l2Anvil, chainConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract callers: %w", err)
	}
	
	// Run the state generator
	startTime := time.Now()
	if err := generator.GenerateState(ctx, l1Caller, l2Caller); err != nil {
		return nil, fmt.Errorf("state generation failed: %w", err)
	}
	duration := time.Since(startTime)
	
	g.logger.Infow("State generation completed",
		"duration", duration,
		"testSuite", testSuite,
	)
	
	// Dump states
	l1StatePath := filepath.Join(outputDir, "l1-state.json")
	l2StatePath := filepath.Join(outputDir, "l2-state.json")
	
	if err := l1Anvil.DumpState(ctx, l1StatePath); err != nil {
		return nil, fmt.Errorf("failed to dump L1 state: %w", err)
	}
	
	if err := l2Anvil.DumpState(ctx, l2StatePath); err != nil {
		return nil, fmt.Errorf("failed to dump L2 state: %w", err)
	}
	
	// Create metadata
	metadata := StateDefinition{
		ID:          generator.GetStateID(),
		Description: generator.GetDescription(),
		TestSuite:   testSuite,
		CreatedAt:   time.Now(),
		Config: map[string]string{
			"l1_chain_id": fmt.Sprintf("%d", l1Anvil.Config.ChainID),
			"l2_chain_id": fmt.Sprintf("%d", l2Anvil.Config.ChainID),
			"duration":    duration.String(),
		},
		L1State: "l1-state.json",
		L2State: "l2-state.json",
	}
	
	// Save metadata
	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}
	
	return &GeneratedState{
		Definition:   metadata,
		L1StatePath:  l1StatePath,
		L2StatePath:  l2StatePath,
		MetadataPath: metadataPath,
	}, nil
}

// GenerateAll generates states for all registered test suites
func (g *TestDataGenerator) GenerateAll(ctx context.Context) error {
	g.mu.RLock()
	suites := make([]string, 0, len(g.generators))
	for suite := range g.generators {
		suites = append(suites, suite)
	}
	g.mu.RUnlock()
	
	g.logger.Infow("Generating states for all test suites",
		"count", len(suites),
		"suites", suites,
	)
	
	for _, suite := range suites {
		if _, err := g.GenerateState(ctx, suite); err != nil {
			g.logger.Errorw("Failed to generate state",
				"testSuite", suite,
				"error", err,
			)
			// Continue with other suites even if one fails
		}
	}
	
	return nil
}

// startAnvilInstances starts L1 and L2 anvil instances
func (g *TestDataGenerator) startAnvilInstances(ctx context.Context) (*AnvilInstance, *AnvilInstance, error) {
	// Start L1 anvil
	l1Config := &AnvilConfig{
		ChainID:   31337,
		Port:      8545,
		StatePath: g.config.BaseL1StatePath,
	}
	
	l1Anvil, err := StartAnvil(ctx, l1Config, g.logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start L1 anvil: %w", err)
	}
	
	// Start L2 anvil
	l2Config := &AnvilConfig{
		ChainID:   31338,
		Port:      9545,
		StatePath: g.config.BaseL2StatePath,
	}
	
	l2Anvil, err := StartAnvil(ctx, l2Config, g.logger)
	if err != nil {
		l1Anvil.Stop()
		return nil, nil, fmt.Errorf("failed to start L2 anvil: %w", err)
	}
	
	return l1Anvil, l2Anvil, nil
}

// createContractCallers creates contract callers for L1 and L2
func (g *TestDataGenerator) createContractCallers(ctx context.Context, l1Anvil, l2Anvil *AnvilInstance, chainConfig *testUtils.ChainConfig) (*caller.ContractCaller, *caller.ContractCaller, error) {
	// Create L1 signer
	l1Signer, err := transactionSigner.NewPrivateKeySigner(chainConfig.OperatorAccountPrivateKey, l1Anvil.GetClient(), &logger.Logger{Logger: g.logger.Desugar()})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}
	
	// Create L1 contract caller
	l1Caller, err := caller.NewContractCaller(l1Anvil.GetClient(), l1Signer, &logger.Logger{Logger: g.logger.Desugar()})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create L1 contract caller: %w", err)
	}
	
	// Create L2 signer
	l2Signer, err := transactionSigner.NewPrivateKeySigner(chainConfig.OperatorAccountPrivateKey, l2Anvil.GetClient(), &logger.Logger{Logger: g.logger.Desugar()})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create L2 signer: %w", err)
	}
	
	// Create L2 contract caller
	l2Caller, err := caller.NewContractCaller(l2Anvil.GetClient(), l2Signer, &logger.Logger{Logger: g.logger.Desugar()})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create L2 contract caller: %w", err)
	}
	
	return l1Caller, l2Caller, nil
}

// loadExistingState loads an existing state from disk
func (g *TestDataGenerator) loadExistingState(outputDir string) (*GeneratedState, error) {
	metadataPath := filepath.Join(outputDir, "metadata.json")
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	
	var metadata StateDefinition
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	return &GeneratedState{
		Definition:   metadata,
		L1StatePath:  filepath.Join(outputDir, metadata.L1State),
		L2StatePath:  filepath.Join(outputDir, metadata.L2State),
		MetadataPath: metadataPath,
	}, nil
}