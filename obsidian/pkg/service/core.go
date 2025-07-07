package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ComputeRequest represents a computation request
type ComputeRequest struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Input     interface{}            `json:"input"`
	Nonce     string                 `json:"nonce"`
	Metadata  map[string]string      `json:"metadata"`
}

// ComputeResult represents computation output
type ComputeResult struct {
	ID            string              `json:"id"`
	RequestID     string              `json:"request_id"`
	Output        interface{}         `json:"output"`
	ComputedAt    time.Time          `json:"computed_at"`
	ComputeTime   time.Duration      `json:"compute_time"`
	NonDeterministic map[string]interface{} `json:"non_deterministic"`
}

// Core handles service business logic
type Core struct {
	outputs    map[string]*ComputeResult
	outputsMu  sync.RWMutex
}

// NewCore creates a new service core
func NewCore() *Core {
	return &Core{
		outputs: make(map[string]*ComputeResult),
	}
}

// ProcessCompute handles computation requests
func (c *Core) ProcessCompute(req ComputeRequest) (*ComputeResult, error) {
	startTime := time.Now()

	// Generate result ID
	resultID := generateID()

	// Perform computation based on type
	var output interface{}
	nonDeterministic := make(map[string]interface{})

	switch req.Type {
	case "random":
		// Example: Generate random data
		randomData := make([]byte, 32)
		rand.Read(randomData)
		output = hex.EncodeToString(randomData)
		nonDeterministic["seed"] = time.Now().UnixNano()

	case "timestamp":
		// Example: Current timestamp computation
		output = map[string]interface{}{
			"unix":      time.Now().Unix(),
			"formatted": time.Now().Format(time.RFC3339),
		}
		nonDeterministic["system_time"] = time.Now().String()

	case "analysis":
		// Example: Analyze input data
		output = map[string]interface{}{
			"processed": true,
			"analysis": fmt.Sprintf("Analyzed input with nonce %s", req.Nonce),
			"score": rand.Float64() * 100,
		}
		nonDeterministic["random_factor"] = rand.Float64()

	default:
		return nil, fmt.Errorf("unknown compute type: %s", req.Type)
	}

	result := &ComputeResult{
		ID:               resultID,
		RequestID:        req.ID,
		Output:           output,
		ComputedAt:       time.Now(),
		ComputeTime:      time.Since(startTime),
		NonDeterministic: nonDeterministic,
	}

	// Store result for verification
	c.outputsMu.Lock()
	c.outputs[resultID] = result
	c.outputsMu.Unlock()

	return result, nil
}

// VerifyOutput checks if an output exists and is valid
func (c *Core) VerifyOutput(outputID string) (bool, error) {
	c.outputsMu.RLock()
	defer c.outputsMu.RUnlock()

	result, exists := c.outputs[outputID]
	if !exists {
		return false, fmt.Errorf("output not found")
	}

	// Check if output is still valid (24 hour window)
	if time.Since(result.ComputedAt) > 24*time.Hour {
		return false, fmt.Errorf("output expired")
	}

	return true, nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}