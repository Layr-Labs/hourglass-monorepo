package harness

import (
	"context"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
)

// StateGenerator defines the interface for test-specific state generators
type StateGenerator interface {
	// GenerateState performs all necessary setup operations on the blockchain
	GenerateState(ctx context.Context, l1Caller, l2Caller contractCaller.IContractCaller) error
	
	// GetStateID returns a unique identifier for this state configuration
	GetStateID() string
	
	// GetDescription returns a human-readable description of the state
	GetDescription() string
}

// StateDefinition contains metadata about a generated state
type StateDefinition struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	TestSuite   string            `json:"test_suite"`
	CreatedAt   time.Time         `json:"created_at"`
	Config      map[string]string `json:"config"`
	L1State     string            `json:"l1_state_file"`
	L2State     string            `json:"l2_state_file"`
}

// GeneratedState represents the result of a state generation
type GeneratedState struct {
	Definition  StateDefinition
	L1StatePath string
	L2StatePath string
	MetadataPath string
}

// GeneratorConfig contains configuration for the test data generator
type GeneratorConfig struct {
	// Base directory for storing generated states
	OutputDir string
	
	// Whether to overwrite existing states
	Overwrite bool
	
	// L1 RPC URL for anvil
	L1RPCURL string
	
	// L2 RPC URL for anvil
	L2RPCURL string
	
	// L1 WebSocket URL
	L1WSURL string
	
	// L2 WebSocket URL
	L2WSURL string
	
	// Path to base L1 state file
	BaseL1StatePath string
	
	// Path to base L2 state file
	BaseL2StatePath string
	
	// Chain configuration path
	ChainConfigPath string
	
	// Verbose logging
	Verbose bool
}

// AnvilConfig contains configuration for an anvil instance
type AnvilConfig struct {
	ChainID    uint64
	Port       uint16
	StatePath  string
	ConfigPath string
	ForkURL    string
	ForkBlock  uint64
}