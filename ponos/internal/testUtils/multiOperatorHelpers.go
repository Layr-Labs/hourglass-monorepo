package testUtils

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/tableTransporter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// MultiOperatorTestConfig contains configuration for multi-operator tests
type MultiOperatorTestConfig struct {
	NumOperators  int
	OperatorSetId uint32
	CurveType     config.CurveType
	AVSAddress    string
	AVSPrivateKey string
	ChainConfig   *ChainConfig
	L1EthClient   *ethclient.Client
	L2EthClient   *ethclient.Client // Optional: L2 client for L2 transport
	L2RpcUrl      string            // Optional: L2 RPC URL
	L2ChainId     uint64            // Optional: L2 chain ID
	Logger        *zap.Logger
	EqualWeights  bool      // If true, all operators get equal weights
	CustomWeights []float64 // Custom weights per operator (converted to proper values)
}

// OperatorTestData contains all the data for a test operator
type OperatorTestData struct {
	Operator         *operator.Operator
	KeyPair          *WrappedKeyPair
	PrivateKey       string
	Address          string
	StakerPrivateKey string
	StakerAddress    string
	Socket           string
	Weight           *big.Int
	BLSInfo          tableTransporter.OperatorBLSInfo
}

// MultiOperatorTestEnvironment contains the complete test environment
type MultiOperatorTestEnvironment struct {
	Config            *MultiOperatorTestConfig
	Operators         []*OperatorTestData
	ContractCaller    contractCaller.IContractCaller
	AVSContractCaller contractCaller.IContractCaller
}

// SetupMultiOperatorEnvironment sets up a complete multi-operator test environment
func SetupMultiOperatorEnvironment(t *testing.T, testConfig *MultiOperatorTestConfig) (*MultiOperatorTestEnvironment, error) {
	ctx := context.Background()

	// Create contract callers
	appPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(testConfig.ChainConfig.AppAccountPrivateKey, testConfig.L1EthClient, testConfig.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create app private key signer: %w", err)
	}

	contractCaller, err := caller.NewContractCaller(testConfig.L1EthClient, appPrivateKeySigner, testConfig.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract caller: %w", err)
	}

	avsPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(testConfig.AVSPrivateKey, testConfig.L1EthClient, testConfig.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVS private key signer: %w", err)
	}

	avsContractCaller, err := caller.NewContractCaller(testConfig.L1EthClient, avsPrivateKeySigner, testConfig.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVS contract caller: %w", err)
	}

	// Configure operator set
	t.Logf("Configuring operator set %d with curve type %v", testConfig.OperatorSetId, testConfig.CurveType)
	_, err = avsContractCaller.ConfigureAVSOperatorSet(ctx,
		common.HexToAddress(testConfig.AVSAddress),
		testConfig.OperatorSetId,
		testConfig.CurveType)
	if err != nil {
		return nil, fmt.Errorf("failed to configure operator set: %w", err)
	}

	// Generate operators
	operators := make([]*OperatorTestData, testConfig.NumOperators)

	// Determine weights
	weights := calculateWeights(testConfig)

	// Get operator accounts from chain config
	operatorAccounts := getOperatorAccounts(testConfig.ChainConfig, testConfig.NumOperators)
	stakerAccounts := getStakerAccounts(testConfig.ChainConfig, testConfig.NumOperators)

	for i := 0; i < testConfig.NumOperators; i++ {
		// Generate keys for curve type
		_, keyPair, _, err := GetKeysForCurveType(t, testConfig.CurveType, testConfig.ChainConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get keys for operator %d: %w", i, err)
		}

		// Create operator
		op := &operator.Operator{
			TransactionPrivateKey: operatorAccounts[i].PrivateKey,
			SigningPrivateKey:     keyPair.PrivateKey,
			Curve:                 testConfig.CurveType,
			OperatorSetIds:        []uint32{testConfig.OperatorSetId},
		}

		// Create BLS info for transport
		var blsInfo tableTransporter.OperatorBLSInfo
		if testConfig.CurveType == config.CurveTypeBN254 {
			blsPrivKey := keyPair.PrivateKey.(*bn254.PrivateKey)
			blsInfo = tableTransporter.OperatorBLSInfo{
				PrivateKeyHex:   fmt.Sprintf("0x%x", blsPrivKey.Bytes()),
				Weights:         []*big.Int{weights[i]},
				OperatorAddress: common.HexToAddress(operatorAccounts[i].Address),
			}
		}

		operators[i] = &OperatorTestData{
			Operator:         op,
			KeyPair:          keyPair,
			PrivateKey:       operatorAccounts[i].PrivateKey,
			Address:          operatorAccounts[i].Address,
			StakerPrivateKey: stakerAccounts[i].PrivateKey,
			StakerAddress:    stakerAccounts[i].Address,
			Socket:           fmt.Sprintf("localhost:%d", 9000+i),
			Weight:           weights[i],
			BLSInfo:          blsInfo,
		}

		t.Logf("Created operator %d: Address=%s, Weight=%s", i, operators[i].Address, operators[i].Weight.String())
	}

	return &MultiOperatorTestEnvironment{
		Config:            testConfig,
		Operators:         operators,
		ContractCaller:    contractCaller,
		AVSContractCaller: avsContractCaller,
	}, nil
}

// RegisterAndDelegateOperators registers operators and delegates stake to them
func RegisterAndDelegateOperators(t *testing.T, env *MultiOperatorTestEnvironment) error {
	ctx := context.Background()

	// Create operator configs
	operatorConfigs := make([]*OperatorConfig, len(env.Operators))
	for i, op := range env.Operators {
		operatorConfigs[i] = &OperatorConfig{
			Operator:        op.Operator,
			Socket:          op.Socket,
			MetadataUri:     "https://test-metadata.com",
			AllocationDelay: 1,
		}
	}

	// Register operators
	t.Logf("Registering %d operators", len(operatorConfigs))
	err := RegisterMultipleOperators(
		ctx,
		env.Config.L1EthClient,
		env.Config.AVSAddress,
		env.Config.AVSPrivateKey,
		operatorConfigs,
		env.Config.Logger,
	)
	if err != nil {
		return fmt.Errorf("failed to register operators: %w", err)
	}

	time.Sleep(time.Second * 6)

	// Create stake configs
	stakeConfigs := make([]*StakerDelegationConfig, len(env.Operators))
	for i, op := range env.Operators {
		stakeConfigs[i] = &StakerDelegationConfig{
			StakerPrivateKey:   op.StakerPrivateKey,
			StakerAddress:      op.StakerAddress,
			OperatorPrivateKey: op.PrivateKey,
			OperatorAddress:    op.Address,
			OperatorSetId:      env.Config.OperatorSetId,
			StrategyAddress:    Strategy_STETH,
		}
	}

	// Delegate stake
	t.Logf("Delegating stake to %d operators", len(stakeConfigs))
	err = DelegateStakeToMultipleOperators(
		t,
		ctx,
		stakeConfigs,
		env.Config.AVSAddress,
		env.Config.L1EthClient,
		env.Config.Logger,
	)
	if err != nil {
		return fmt.Errorf("failed to delegate stake: %w", err)
	}

	return nil
}

// TransportTablesForEnvironment transports tables for the test environment
// Important: The generator in OperatorTableUpdater is GLOBAL, not per-operator-set.
// When we update it with custom operator info for testing, it affects ALL operator sets.
// To avoid conflicts, we only transport our specific operator set.
func TransportTablesForEnvironment(t *testing.T, env *MultiOperatorTestEnvironment) error {
	// Create generation reservation for our operator set
	ctx := context.Background()
	avsAddr := common.HexToAddress(env.Config.AVSAddress)
	maxStalenessPeriod := uint32(604800) // 1 week

	bn254CalculatorAddr := env.AVSContractCaller.GetTableCalculatorAddress(env.Config.CurveType)
	t.Logf("Creating generation reservation with table calculator %s for operator set %d",
		bn254CalculatorAddr.Hex(), env.Config.OperatorSetId)

	_, err := env.AVSContractCaller.CreateGenerationReservation(
		ctx,
		avsAddr,
		env.Config.OperatorSetId,
		bn254CalculatorAddr,
		avsAddr, // AVS is the owner
		maxStalenessPeriod,
	)
	if err != nil {
		t.Logf("Warning: Failed to create generation reservation: %v", err)
	}

	time.Sleep(time.Second * 3)

	// Prepare BLS infos for our test operators
	blsInfos := make([]tableTransporter.OperatorBLSInfo, len(env.Operators))
	for i, op := range env.Operators {
		blsInfos[i] = op.BLSInfo
	}

	// Transport tables with our custom operator info
	// WARNING: This updates the GLOBAL generator in OperatorTableUpdater
	// We limit transport to only our operator set to avoid breaking others
	t.Logf("Transporting tables for %d operators (limiting to operator set %d)",
		len(blsInfos), env.Config.OperatorSetId)

	// Determine which chains to ignore based on L2 config
	chainIdsToIgnore := []*big.Int{
		new(big.Int).SetUint64(11155111), // eth sepolia
		new(big.Int).SetUint64(17000),    // holesky
		new(big.Int).SetUint64(84532),    // base sepolia
	}

	// If L2 is not configured, also ignore L2 anvil
	if env.Config.L2EthClient == nil || env.Config.L2ChainId == 0 {
		chainIdsToIgnore = append(chainIdsToIgnore, new(big.Int).SetUint64(31338)) // L2 anvil
	}

	err = TransportStakeTablesWithMultipleOperatorsConfig(
		env.Config.Logger,
		blsInfos,
		env.Config.AVSPrivateKey,
		env.Config.OperatorSetId,
		env.Config.AVSAddress,
		env.Config.L2RpcUrl,
		env.Config.L2ChainId,
		chainIdsToIgnore,
	)
	if err != nil {
		return fmt.Errorf("failed to transport tables: %w", err)
	}

	time.Sleep(time.Second * 6)
	return nil
}

// Helper functions

func calculateWeights(testConfig *MultiOperatorTestConfig) []*big.Int {
	weights := make([]*big.Int, testConfig.NumOperators)

	if testConfig.EqualWeights {
		// Equal weights for all operators
		for i := 0; i < testConfig.NumOperators; i++ {
			weights[i] = big.NewInt(1000000000000000000) // 1e18
		}
	} else if len(testConfig.CustomWeights) == testConfig.NumOperators {
		// Use custom weights
		for i, w := range testConfig.CustomWeights {
			// Convert float to wei (multiply by 1e18)
			weiValue := new(big.Float).Mul(big.NewFloat(w), big.NewFloat(1e18))
			weights[i], _ = weiValue.Int(nil)
		}
	} else {
		// Default descending weights (2, 1.5, 1, 0.5, ...)
		base := big.NewInt(2000000000000000000) // 2e18
		for i := 0; i < testConfig.NumOperators; i++ {
			weights[i] = new(big.Int).Div(base, big.NewInt(int64(i+1)))
		}
	}

	return weights
}

type AccountInfo struct {
	PrivateKey string
	Address    string
}

func getOperatorAccounts(chainConfig *ChainConfig, numOperators int) []AccountInfo {
	accounts := []AccountInfo{
		{PrivateKey: chainConfig.ExecOperatorAccountPk, Address: chainConfig.ExecOperatorAccountAddress},
		{PrivateKey: chainConfig.ExecOperator2AccountPk, Address: chainConfig.ExecOperator2AccountAddress},
		{PrivateKey: chainConfig.ExecOperator3AccountPk, Address: chainConfig.ExecOperator3AccountAddress},
		{PrivateKey: chainConfig.ExecOperator4AccountPk, Address: chainConfig.ExecOperator4AccountAddress},
	}

	if numOperators > len(accounts) {
		panic(fmt.Sprintf("requested %d operators but only %d available in chain config", numOperators, len(accounts)))
	}

	return accounts[:numOperators]
}

func getStakerAccounts(chainConfig *ChainConfig, numStakers int) []AccountInfo {
	accounts := []AccountInfo{
		{PrivateKey: chainConfig.ExecStakerAccountPrivateKey, Address: chainConfig.ExecStakerAccountAddress},
		{PrivateKey: chainConfig.ExecStaker2AccountPrivateKey, Address: chainConfig.ExecStaker2AccountAddress},
		{PrivateKey: chainConfig.ExecStaker3AccountPrivateKey, Address: chainConfig.ExecStaker3AccountAddress},
		{PrivateKey: chainConfig.ExecStaker4AccountPrivateKey, Address: chainConfig.ExecStaker4AccountAddress},
	}

	if numStakers > len(accounts) {
		panic(fmt.Sprintf("requested %d stakers but only %d available in chain config", numStakers, len(accounts)))
	}

	return accounts[:numStakers]
}
