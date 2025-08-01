package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ChainConfig represents the test chain configuration
type ChainConfig struct {
	// Account addresses and keys
	DeployAccountAddress    string `json:"deployAccountAddress"`
	DeployAccountPk         string `json:"deployAccountPk"`
	DeployAccountPublicKey  string `json:"deployAccountPublicKey"`
	
	AVSAccountAddress       string `json:"avsAccountAddress"`
	AVSAccountPk            string `json:"avsAccountPk"`
	AVSAccountPublicKey     string `json:"avsAccountPublicKey"`
	
	AppAccountAddress       string `json:"appAccountAddress"`
	AppAccountPk            string `json:"appAccountPk"`
	AppAccountPublicKey     string `json:"appAccountPublicKey"`
	
	OperatorAccountAddress  string `json:"operatorAccountAddress"`
	OperatorAccountPk       string `json:"operatorAccountPk"`
	OperatorAccountPublicKey string `json:"operatorAccountPublicKey"`
	OperatorKeystorePath    string `json:"operatorKeystorePath"`
	OperatorKeystorePassword string `json:"operatorKeystorePassword"`
	
	ExecOperatorAccountAddress  string `json:"execOperatorAccountAddress"`
	ExecOperatorAccountPk       string `json:"execOperatorAccountPk"`
	ExecOperatorAccountPublicKey string `json:"execOperatorAccountPublicKey"`
	ExecOperatorKeystorePath    string `json:"execOperatorKeystorePath"`
	ExecOperatorKeystorePassword string `json:"execOperatorKeystorePassword"`
	
	// Contract addresses
	AVSTaskRegistrarAddress string `json:"avsTaskRegistrarAddress"`
	AVSTaskHookAddressL1    string `json:"avsTaskHookAddressL1"`
	AVSTaskHookAddressL2    string `json:"avsTaskHookAddressL2"`
	KeyRegistrarAddress     string `json:"keyRegistrarAddress"`
	ReleaseManagerAddress   string `json:"releaseManagerAddress"`
	
	// Environment info
	DestinationEnv string `json:"destinationEnv"`
	ForkL1Block    int64  `json:"forkL1Block"`
	ForkL2Block    int64  `json:"forkL2Block"`
}

// LoadChainConfig loads the chain configuration from the JSON file
func LoadChainConfig(projectRoot string) (*ChainConfig, error) {
	configPath := filepath.Join(projectRoot, "hgctl-go", "internal", "testutils", "chainData", "chain-config.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chain config: %w", err)
	}
	
	var config ChainConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse chain config: %w", err)
	}
	
	return &config, nil
}

// GetTestChainCoreContracts returns core contract addresses for test chains
func GetTestChainCoreContracts(chainID uint64, config *ChainConfig) map[string]string {
	contracts := make(map[string]string)
	
	// For test chains, we use the addresses from our deployed contracts
	if chainID == 31337 { // L1 test chain
		contracts["KeyRegistrar"] = config.KeyRegistrarAddress
		contracts["ReleaseManager"] = config.ReleaseManagerAddress
		// Add other L1 contracts as needed
	} else if chainID == 31338 { // L2 test chain
		// Add L2 specific contracts
	}
	
	return contracts
}