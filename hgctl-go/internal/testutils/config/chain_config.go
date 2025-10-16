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
	OperatorAccountAddress         string `json:"operatorAccountAddress"`
	OperatorAccountPk              string `json:"operatorAccountPk"`
	OperatorAccountPublicKey       string `json:"operatorAccountPublicKey"`
	OperatorKeystorePath           string `json:"operatorKeystorePath"`
	OperatorKeystorePassword       string `json:"operatorKeystorePassword"`
	OperatorSystemAddress          string `json:"operatorSystemAddress"`
	OperatorSystemPk               string `json:"operatorSystemPk"`
	OperatorSystemKeystorePath     string `json:"operatorSystemKeystorePath"`
	OperatorSystemKeystorePassword string `json:"operatorSystemKeystorePassword"`

	ExecOperatorAccountAddress         string `json:"execOperatorAccountAddress"`
	ExecOperatorAccountPk              string `json:"execOperatorAccountPk"`
	ExecOperatorAccountPublicKey       string `json:"execOperatorAccountPublicKey"`
	ExecOperatorKeystorePath           string `json:"execOperatorKeystorePath"`
	ExecOperatorKeystorePassword       string `json:"execOperatorKeystorePassword"`
	ExecOperatorSystemAddress          string `json:"execOperatorSystemAddress"`
	ExecOperatorSystemPk               string `json:"execOperatorSystemPk"`
	ExecOperatorSystemKeystorePath     string `json:"execOperatorSystemKeystorePath"`
	ExecOperatorSystemKeystorePassword string `json:"execOperatorSystemKeystorePassword"`

	// Additional executor operators
	ExecOperator2AccountAddress         string `json:"execOperator2AccountAddress"`
	ExecOperator2AccountPk              string `json:"execOperator2AccountPk"`
	ExecOperator2AccountPublicKey       string `json:"execOperator2AccountPublicKey"`
	ExecOperator2KeystorePath           string `json:"execOperator2KeystorePath"`
	ExecOperator2KeystorePassword       string `json:"execOperator2KeystorePassword"`
	ExecOperator2SystemAddress          string `json:"execOperator2SystemAddress"`
	ExecOperator2SystemPk               string `json:"execOperator2SystemPk"`
	ExecOperator2SystemKeystorePath     string `json:"execOperator2SystemKeystorePath"`
	ExecOperator2SystemKeystorePassword string `json:"execOperator2SystemKeystorePassword"`

	ExecOperator3AccountAddress         string `json:"execOperator3AccountAddress"`
	ExecOperator3AccountPk              string `json:"execOperator3AccountPk"`
	ExecOperator3AccountPublicKey       string `json:"execOperator3AccountPublicKey"`
	ExecOperator3KeystorePath           string `json:"execOperator3KeystorePath"`
	ExecOperator3KeystorePassword       string `json:"execOperator3KeystorePassword"`
	ExecOperator3SystemAddress          string `json:"execOperator3SystemAddress"`
	ExecOperator3SystemPk               string `json:"execOperator3SystemPk"`
	ExecOperator3SystemKeystorePath     string `json:"execOperator3SystemKeystorePath"`
	ExecOperator3SystemKeystorePassword string `json:"execOperator3SystemKeystorePassword"`

	ExecOperator4AccountAddress         string `json:"execOperator4AccountAddress"`
	ExecOperator4AccountPk              string `json:"execOperator4AccountPk"`
	ExecOperator4AccountPublicKey       string `json:"execOperator4AccountPublicKey"`
	ExecOperator4KeystorePath           string `json:"execOperator4KeystorePath"`
	ExecOperator4KeystorePassword       string `json:"execOperator4KeystorePassword"`
	ExecOperator4SystemAddress          string `json:"execOperator4SystemAddress"`
	ExecOperator4SystemPk               string `json:"execOperator4SystemPk"`
	ExecOperator4SystemKeystorePath     string `json:"execOperator4SystemKeystorePath"`
	ExecOperator4SystemKeystorePassword string `json:"execOperator4SystemKeystorePassword"`

	// Unregistered operators (for testing registration flow)
	UnregisteredOperator1AccountAddress              string `json:"unregisteredOperator1AccountAddress"`
	UnregisteredOperator1AccountPk                   string `json:"unregisteredOperator1AccountPk"`
	UnregisteredOperator1AccountPublicKey            string `json:"unregisteredOperator1AccountPublicKey"`
	UnregisteredOperator1KeystorePath                string `json:"unregisteredOperator1KeystorePath"`
	UnregisteredOperator1KeystorePassword            string `json:"unregisteredOperator1KeystorePassword"`
	UnregisteredOperator1SystemBN254Pk               string `json:"unregisteredOperator1SystemBN254Pk"`
	UnregisteredOperator1SystemBN254KeystorePath     string `json:"unregisteredOperator1SystemBN254KeystorePath"`
	UnregisteredOperator1SystemBN254KeystorePassword string `json:"unregisteredOperator1SystemBN254KeystorePassword"`
	UnregisteredOperator1SystemECDSAPk               string `json:"unregisteredOperator1SystemECDSAPk"`
	UnregisteredOperator1SystemECDSAAddress          string `json:"unregisteredOperator1SystemECDSAAddress"`
	UnregisteredOperator1SystemECDSAKeystorePath     string `json:"unregisteredOperator1SystemECDSAKeystorePath"`
	UnregisteredOperator1SystemECDSAKeystorePassword string `json:"unregisteredOperator1SystemECDSAKeystorePassword"`

	UnregisteredOperator2AccountAddress              string `json:"unregisteredOperator2AccountAddress"`
	UnregisteredOperator2AccountPk                   string `json:"unregisteredOperator2AccountPk"`
	UnregisteredOperator2AccountPublicKey            string `json:"unregisteredOperator2AccountPublicKey"`
	UnregisteredOperator2KeystorePath                string `json:"unregisteredOperator2KeystorePath"`
	UnregisteredOperator2KeystorePassword            string `json:"unregisteredOperator2KeystorePassword"`
	UnregisteredOperator2SystemBN254Pk               string `json:"unregisteredOperator2SystemBN254Pk"`
	UnregisteredOperator2SystemBN254KeystorePath     string `json:"unregisteredOperator2SystemBN254KeystorePath"`
	UnregisteredOperator2SystemBN254KeystorePassword string `json:"unregisteredOperator2SystemBN254KeystorePassword"`
	UnregisteredOperator2SystemECDSAPk               string `json:"unregisteredOperator2SystemECDSAPk"`
	UnregisteredOperator2SystemECDSAAddress          string `json:"unregisteredOperator2SystemECDSAAddress"`
	UnregisteredOperator2SystemECDSAKeystorePath     string `json:"unregisteredOperator2SystemECDSAKeystorePath"`
	UnregisteredOperator2SystemECDSAKeystorePassword string `json:"unregisteredOperator2SystemECDSAKeystorePassword"`

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
	AVSTaskRegistrarAddress     string `json:"avsTaskRegistrarAddress"`
	AVSTaskHookAddressL1        string `json:"avsTaskHookAddressL1"`
	AVSTaskHookAddressL2        string `json:"avsTaskHookAddressL2"`
	KeyRegistrarAddress         string `json:"keyRegistrarAddress"`
	ReleaseManagerAddress       string `json:"releaseManagerAddress"`
	DelegationManagerAddress    string `json:"delegationManagerAddress"`
	AllocationManagerAddress    string `json:"allocationManagerAddress"`
	StrategyManagerAddress      string `json:"strategyManagerAddress"`
	PermissionControllerAddress string `json:"permissionControllerAddress"`
	TaskMailboxAddress          string `json:"taskMailboxAddress"`

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
