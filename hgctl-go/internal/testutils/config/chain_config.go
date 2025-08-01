package config

// ChainConfig represents the test chain configuration
type ChainConfig struct {
	// Account addresses and keys
	DeployAccountAddress   string `json:"deployAccountAddress"`
	DeployAccountPk        string `json:"deployAccountPk"`
	DeployAccountPublicKey string `json:"deployAccountPublicKey"`

	AVSAccountAddress   string `json:"avsAccountAddress"`
	AVSAccountPk        string `json:"avsAccountPk"`
	AVSAccountPublicKey string `json:"avsAccountPublicKey"`

	AppAccountAddress   string `json:"appAccountAddress"`
	AppAccountPk        string `json:"appAccountPk"`
	AppAccountPublicKey string `json:"appAccountPublicKey"`

	// Operator accounts
	OperatorAccountAddress   string `json:"operatorAccountAddress"`
	OperatorAccountPk        string `json:"operatorAccountPk"`
	OperatorAccountPublicKey string `json:"operatorAccountPublicKey"`
	OperatorKeystorePath     string `json:"operatorKeystorePath"`
	OperatorKeystorePassword string `json:"operatorKeystorePassword"`

	ExecOperatorAccountAddress   string `json:"execOperatorAccountAddress"`
	ExecOperatorAccountPk        string `json:"execOperatorAccountPk"`
	ExecOperatorAccountPublicKey string `json:"execOperatorAccountPublicKey"`
	ExecOperatorKeystorePath     string `json:"execOperatorKeystorePath"`
	ExecOperatorKeystorePassword string `json:"execOperatorKeystorePassword"`

	// Contract addresses
	AVSTaskRegistrarAddress  string `json:"avsTaskRegistrarAddress"`
	AVSTaskHookAddressL1     string `json:"avsTaskHookAddressL1"`
	AVSTaskHookAddressL2     string `json:"avsTaskHookAddressL2"`
	KeyRegistrarAddress      string `json:"keyRegistrarAddress"`
	ReleaseManagerAddress    string `json:"releaseManagerAddress"`
	DelegationManagerAddress string `json:"delegationManagerAddress"`
	AllocationManagerAddress string `json:"allocationManagerAddress"`
	StrategyManagerAddress   string `json:"strategyManagerAddress"`

	// Chain configuration
	L1ChainID int    `json:"l1ChainId,omitempty"`
	L2ChainID int    `json:"l2ChainId,omitempty"`
	L1RPC     string `json:"l1RPC,omitempty"`
	L2RPC     string `json:"l2RPC,omitempty"`

	// Environment info
	DestinationEnv string `json:"destinationEnv"`
	ForkL1Block    int64  `json:"forkL1Block"`
	ForkL2Block    int64  `json:"forkL2Block"`
}

//// LoadChainConfig loads the chain configuration from the JSON file
//func LoadChainConfig(projectRoot string) (*ChainConfig, error) {
//	configPath := filepath.Join(projectRoot, "hgctl-go", "internal", "testutils", "chainData", "chain-config.json")
//
//	data, err := os.ReadFile(configPath)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read chain config: %w", err)
//	}
//
//	var config ChainConfig
//	if err := json.Unmarshal(data, &config); err != nil {
//		return nil, fmt.Errorf("failed to parse chain config: %w", err)
//	}
//
//	// Set default chain IDs and RPCs if not in config
//	if config.L1ChainID == 0 {
//		config.L1ChainID = 31337
//	}
//	if config.L2ChainID == 0 {
//		config.L2ChainID = 31338
//	}
//	if config.L1RPC == "" {
//		config.L1RPC = "http://localhost:8545"
//	}
//	if config.L2RPC == "" {
//		config.L2RPC = "http://localhost:9545"
//	}
//
//	return &config, nil
//}
