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

	// Additional executor operators
	ExecOperator2AccountAddress   string `json:"execOperator2AccountAddress"`
	ExecOperator2AccountPk        string `json:"execOperator2AccountPk"`
	ExecOperator2AccountPublicKey string `json:"execOperator2AccountPublicKey"`
	ExecOperator2KeystorePath     string `json:"execOperator2KeystorePath"`
	ExecOperator2KeystorePassword string `json:"execOperator2KeystorePassword"`

	ExecOperator3AccountAddress   string `json:"execOperator3AccountAddress"`
	ExecOperator3AccountPk        string `json:"execOperator3AccountPk"`
	ExecOperator3AccountPublicKey string `json:"execOperator3AccountPublicKey"`
	ExecOperator3KeystorePath     string `json:"execOperator3KeystorePath"`
	ExecOperator3KeystorePassword string `json:"execOperator3KeystorePassword"`

	ExecOperator4AccountAddress   string `json:"execOperator4AccountAddress"`
	ExecOperator4AccountPk        string `json:"execOperator4AccountPk"`
	ExecOperator4AccountPublicKey string `json:"execOperator4AccountPublicKey"`
	ExecOperator4KeystorePath     string `json:"execOperator4KeystorePath"`
	ExecOperator4KeystorePassword string `json:"execOperator4KeystorePassword"`

	// Unregistered operator (for testing registration flow)
	UnregisteredOperatorAccountAddress   string `json:"unregisteredOperatorAccountAddress"`
	UnregisteredOperatorAccountPk        string `json:"unregisteredOperatorAccountPk"`
	UnregisteredOperatorAccountPublicKey string `json:"unregisteredOperatorAccountPublicKey"`
	UnregisteredOperatorKeystorePath     string `json:"unregisteredOperatorKeystorePath"`
	UnregisteredOperatorKeystorePassword string `json:"unregisteredOperatorKeystorePassword"`

	// Staker accounts
	AggStakerAccountAddress   string `json:"aggStakerAccountAddress"`
	AggStakerAccountPk        string `json:"aggStakerAccountPk"`
	AggStakerAccountPublicKey string `json:"aggStakerAccountPublicKey"`

	ExecStakerAccountAddress   string `json:"execStakerAccountAddress"`
	ExecStakerAccountPk        string `json:"execStakerAccountPk"`
	ExecStakerAccountPublicKey string `json:"execStakerAccountPublicKey"`

	ExecStaker2AccountAddress   string `json:"execStaker2AccountAddress"`
	ExecStaker2AccountPk        string `json:"execStaker2AccountPk"`
	ExecStaker2AccountPublicKey string `json:"execStaker2AccountPublicKey"`

	ExecStaker3AccountAddress   string `json:"execStaker3AccountAddress"`
	ExecStaker3AccountPk        string `json:"execStaker3AccountPk"`
	ExecStaker3AccountPublicKey string `json:"execStaker3AccountPublicKey"`

	ExecStaker4AccountAddress   string `json:"execStaker4AccountAddress"`
	ExecStaker4AccountPk        string `json:"execStaker4AccountPk"`
	ExecStaker4AccountPublicKey string `json:"execStaker4AccountPublicKey"`

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
