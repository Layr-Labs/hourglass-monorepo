package testUtils

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

const (
	Strategy_WETH  = "0x0Fe4F44beE93503346A3Ac9EE5A26b130a5796d6"
	Strategy_STETH = "0x93c4b944D05dfe6df7645A86cd2206016c51564D"
)

type ExecutorWithSocket struct {
	Executor *operator.Operator
	Socket   string
}

type StakerDelegationConfig struct {
	StakerPrivateKey   string
	StakerAddress      string
	OperatorPrivateKey string
	OperatorAddress    string
	OperatorSetId      uint32
	StrategyAddress    string
	Magnitude          uint64
}

// OperatorConfig holds all configuration for an operator including
// registration details and runtime configuration
type OperatorConfig struct {
	Operator        *operator.Operator
	Socket          string         // Socket endpoint for this operator (e.g., "localhost:9000")
	MetadataUri     string         // Metadata URI for the operator
	AllocationDelay uint32         // Allocation delay for the operator
	Address         common.Address // Derived operator address
}

// RegisterMultipleOperators registers multiple operators with their own transaction keys
// This is a generic function that works for any curve type (BN254, ECDSA, etc.)
func RegisterMultipleOperators(
	ctx context.Context,
	ethClient *ethclient.Client,
	avsAddress string,
	avsPrivateKey string,
	operatorConfigs []*OperatorConfig,
	l *zap.Logger,
) error {
	// Create AVS signer for registering all operators
	avsSigner, err := transactionSigner.NewPrivateKeySigner(avsPrivateKey, ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create AVS signer: %v", err)
	}

	avsCaller, err := caller.NewContractCaller(ethClient, avsSigner, l)
	if err != nil {
		return fmt.Errorf("failed to create AVS caller: %v", err)
	}

	for i, opConfig := range operatorConfigs {
		op := opConfig.Operator

		// Create transaction signer for this specific operator using its own private key
		operatorSigner, err := transactionSigner.NewPrivateKeySigner(op.TransactionPrivateKey, ethClient, l)
		if err != nil {
			return fmt.Errorf("failed to create operator %d private key signer: %v", i+1, err)
		}

		operatorCaller, err := caller.NewContractCaller(ethClient, operatorSigner, l)
		if err != nil {
			return fmt.Errorf("failed to create operator %d contract caller: %v", i+1, err)
		}

		// Derive the operator address and store it in the config
		operatorAddress, err := op.DeriveAddress()
		if err != nil {
			return fmt.Errorf("failed to derive address for operator %d: %v", i+1, err)
		}
		opConfig.Address = operatorAddress

		l.Sugar().Infow("Registering operator",
			zap.Int("operatorNumber", i+1),
			zap.String("operatorAddress", operatorAddress.Hex()),
			zap.Uint32s("operatorSetIds", op.OperatorSetIds),
			zap.String("socket", opConfig.Socket),
			zap.String("curveType", string(op.Curve)),
		)

		// Register the operator to operator sets
		_, err = operator.RegisterOperatorToOperatorSets(
			ctx,
			avsCaller,
			operatorCaller,
			common.HexToAddress(avsAddress),
			op.OperatorSetIds,
			op,
			&operator.RegistrationConfig{
				Socket:          opConfig.Socket,
				MetadataUri:     opConfig.MetadataUri,
				AllocationDelay: opConfig.AllocationDelay,
			},
			l,
		)
		if err != nil {
			return fmt.Errorf("failed to register operator %d: %v", i+1, err)
		}
	}

	return nil
}

func SetupOperatorPeeringWithMultipleExecutors(
	ctx context.Context,
	chainConfig *ChainConfig,
	chainId config.ChainId,
	ethClient *ethclient.Client,
	aggregator *operator.Operator,
	executors []ExecutorWithSocket,
	l *zap.Logger,
) error {
	aggOperatorAddress, err := aggregator.DeriveAddress()
	if err != nil {
		return fmt.Errorf("failed to convert aggregator operator private key: %v", err)
	}

	avsPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create AVS private key signer: %v", err)
	}

	avsCc, err := caller.NewContractCaller(ethClient, avsPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create AVS contract caller: %v", err)
	}

	aggregatorPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.OperatorAccountPrivateKey, ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create aggregator private key signer: %v", err)
	}

	aggregatorCc, err := caller.NewContractCaller(ethClient, aggregatorPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create aggregator contract caller: %v", err)
	}

	l.Sugar().Infow("------------------- Registering aggregator -------------------")
	if len(aggregator.OperatorSetIds) == 0 {
		l.Sugar().Infow("No operator sets defined for aggregator")
		return fmt.Errorf("aggregator operator sets are empty, cannot register")
	}
	// register the aggregator
	result, err := operator.RegisterOperatorToOperatorSets(
		ctx,
		avsCc,
		aggregatorCc,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		aggregator.OperatorSetIds,
		aggregator,
		&operator.RegistrationConfig{
			Socket:          "",
			MetadataUri:     "https://some-metadata-uri.com",
			AllocationDelay: 1,
		},
		l,
	)
	if err != nil {
		return fmt.Errorf("failed to register aggregator operator: %v", err)
	}
	l.Sugar().Infow("Aggregator operator registered successfully",
		zap.String("operatorAddress", aggOperatorAddress.String()),
		zap.String("transactionHash", result.TxHash.String()),
	)

	executorPrivateKeys := []string{
		chainConfig.ExecOperatorAccountPk,
		chainConfig.ExecOperator2AccountPk,
		chainConfig.ExecOperator3AccountPk,
		chainConfig.ExecOperator4AccountPk,
	}

	for i, execWithSocket := range executors {
		if i >= len(executorPrivateKeys) {
			return fmt.Errorf("executor index %d exceeds available private keys (%d)", i, len(executorPrivateKeys))
		}

		exec := execWithSocket.Executor
		execOperatorAddress, err := exec.DeriveAddress()
		if err != nil {
			return fmt.Errorf("failed to convert executor %d operator private key: %v", i, err)
		}

		executorPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(executorPrivateKeys[i], ethClient, l)
		if err != nil {
			return fmt.Errorf("failed to create executor %d private key signer: %v", i, err)
		}

		executorCc, err := caller.NewContractCaller(ethClient, executorPrivateKeySigner, l)
		if err != nil {
			return fmt.Errorf("failed to create executor %d contract caller: %v", i, err)
		}

		l.Sugar().Infow("------------------- Registering executor -------------------", zap.Int("executorIndex", i))
		if len(exec.OperatorSetIds) == 0 {
			l.Sugar().Infow("No operator sets defined for executor", zap.Int("executorIndex", i))
			return fmt.Errorf("executor %d operator sets are empty, cannot register", i)
		}

		// register the executor
		result, err = operator.RegisterOperatorToOperatorSets(
			ctx,
			avsCc,
			executorCc,
			common.HexToAddress(chainConfig.AVSAccountAddress),
			exec.OperatorSetIds,
			exec,
			&operator.RegistrationConfig{
				Socket:          execWithSocket.Socket,
				MetadataUri:     "https://some-metadata-uri.com",
				AllocationDelay: 1,
			},
			l,
		)
		if err != nil {
			return fmt.Errorf("failed to register executor %d operator: %v", i, err)
		}
		l.Sugar().Infow("Executor operator registered successfully",
			zap.Int("executorIndex", i),
			zap.String("operatorAddress", execOperatorAddress.String()),
			zap.String("transactionHash", result.TxHash.String()),
		)
	}

	return nil
}

func SetupOperatorPeering(
	ctx context.Context,
	chainConfig *ChainConfig,
	chainId config.ChainId,
	ethClient *ethclient.Client,
	aggregator *operator.Operator,
	executor *operator.Operator,
	socket string,
	l *zap.Logger,
) error {
	return SetupOperatorPeeringWithMultipleExecutors(
		ctx,
		chainConfig,
		chainId,
		ethClient,
		aggregator,
		[]ExecutorWithSocket{{Executor: executor, Socket: socket}},
		l,
	)
}

func DelegateStakeToOperators(
	t *testing.T,
	ctx context.Context,
	aggregatorConfig *StakerDelegationConfig,
	executorConfig *StakerDelegationConfig,
	avsAddress string,
	ethClient *ethclient.Client,
	l *zap.Logger,
) error {
	t.Logf("------------------------ Delegating aggregator ------------------------")

	// Use configured magnitude or default to 1e18
	aggMagnitude := aggregatorConfig.Magnitude
	if aggMagnitude == 0 {
		aggMagnitude = 1e18
	}

	err := DelegateStakeToOperator(
		ctx,
		aggregatorConfig.StakerPrivateKey,
		aggregatorConfig.StakerAddress,
		aggregatorConfig.OperatorPrivateKey,
		aggregatorConfig.OperatorAddress,
		avsAddress,
		aggregatorConfig.OperatorSetId,
		aggregatorConfig.StrategyAddress,
		aggMagnitude,
		ethClient,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to delegate stake to aggregator operator: %v", err)
	}

	t.Logf("------------------------ Delegating Executor ------------------------")

	// Use configured magnitude or default to 1e18
	execMagnitude := executorConfig.Magnitude
	if execMagnitude == 0 {
		execMagnitude = 1e18
	}

	err = DelegateStakeToOperator(
		ctx,
		executorConfig.StakerPrivateKey,
		executorConfig.StakerAddress,
		executorConfig.OperatorPrivateKey,
		executorConfig.OperatorAddress,
		avsAddress,
		executorConfig.OperatorSetId,
		executorConfig.StrategyAddress,
		execMagnitude,
		ethClient,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to delegate stake to aggregator operator: %v", err)
	}
	return nil
}

// DelegateStakeToMultipleOperators delegates stake to multiple operators in a single call
func DelegateStakeToMultipleOperators(
	t *testing.T,
	ctx context.Context,
	configs []*StakerDelegationConfig,
	avsAddress string,
	ethClient *ethclient.Client,
	l *zap.Logger,
) error {
	for i, config := range configs {
		t.Logf("------------------------ Delegating Operator %d ------------------------", i+1)

		// Use configured magnitude or default to 1e18 if not specified
		magnitude := config.Magnitude
		if magnitude == 0 {
			magnitude = 1e18
			l.Sugar().Warnw("No magnitude specified, using default 1e18",
				zap.String("operatorAddress", config.OperatorAddress),
			)
		}

		err := DelegateStakeToOperator(
			ctx,
			config.StakerPrivateKey,
			config.StakerAddress,
			config.OperatorPrivateKey,
			config.OperatorAddress,
			avsAddress,
			config.OperatorSetId,
			config.StrategyAddress,
			magnitude,
			ethClient,
			l,
		)
		if err != nil {
			return fmt.Errorf("failed to delegate stake to operator %d (%s): %v", i+1, config.OperatorAddress, err)
		}
	}
	return nil
}

func DelegateStakeToOperator(
	ctx context.Context,
	stakerPrivateKey string,
	stakerAddress string,
	operatorPrivateKey string,
	operatorAddress string,
	avsAddress string,
	operatorSetId uint32,
	strategyAddress string,
	magnitude uint64,
	ec *ethclient.Client,
	l *zap.Logger,
) error {
	l.Sugar().Infow("Delegating stake to operator",
		zap.String("stakerPrivateKey", stakerPrivateKey),
		zap.String("operatorPrivateKey", operatorPrivateKey),
		zap.String("operatorAddress", operatorAddress),
		zap.String("avsAddress", avsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("strategyAddress", strategyAddress),
		zap.Uint64("magnitude", magnitude),
	)
	stakerPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(stakerPrivateKey, ec, l)
	if err != nil {
		return fmt.Errorf("failed to create staker private key signer: %v", err)
	}

	stakerCc, err := caller.NewContractCaller(ec, stakerPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create staker contract caller: %v", err)
	}

	operatorAddr := common.HexToAddress(operatorAddress)
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return fmt.Errorf("failed to connect to operator: %v", err)
	}
	defer client.Close()

	// Calculate storage slot for *allocationDelayInfo mapping
	// For mapping(address => struct), storage slot = keccak256(abi.encode(key, slot))
	slotBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(slotBytes[24:], 155)
	keyBytes := common.LeftPadBytes(operatorAddr.Bytes(), 32)

	encoded := append(keyBytes, slotBytes...)
	storageKey := common.BytesToHash(crypto.Keccak256(encoded))

	// Define struct fields
	var (
		delay        uint32 = 0    // rightmost 4 bytes
		isSet        byte   = 0x00 // 1 byte before delay
		pendingDelay uint32 = 0    // 4 bytes before isSet
		effectBlock  uint32 = 1    // 4 bytes before pendingDelay - set to early block (far in the past)
	)

	// Create a 32-byte array (filled with zeros)
	structValue := make([]byte, 32)

	// Offset starts from the right
	offset := 32

	// Set delay (4 bytes)
	offset -= 4
	binary.BigEndian.PutUint32(structValue[offset:], delay)

	// Set isSet (1 byte)
	offset -= 1
	structValue[offset] = isSet

	// Set pendingDelay (4 bytes)
	offset -= 4
	binary.BigEndian.PutUint32(structValue[offset:], pendingDelay)

	// Set effectBlock (4 bytes)
	offset -= 4
	binary.BigEndian.PutUint32(structValue[offset:], effectBlock)

	// Mainnet AllocationManager address
	allocationManagerAddr := common.HexToAddress("0x948a420b8CC1d6BFd0B6087C2E7c344a2CD0bc39")

	var setStorageResult interface{}
	err = client.Client().Call(&setStorageResult, "anvil_setStorageAt",
		allocationManagerAddr.Hex(),
		storageKey.Hex(),
		"0x"+hex.EncodeToString(structValue))
	if err != nil {
		l.Warn("Failed to manipulate AllocationDelayInfo storage for operator")
	} else {
		l.Info("Manipulated AllocationDelayInfo storage for operator at effectBlock")
	}

	if _, err := stakerCc.DelegateToOperator(
		ctx,
		common.HexToAddress(operatorAddress),
	); err != nil {
		return fmt.Errorf("failed to delegate stake to operator %s: %v", operatorAddress, err)
	}

	opPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(operatorPrivateKey, ec, l)
	if err != nil {
		return fmt.Errorf("failed to create operator private key signer: %v", err)
	}

	opCc, err := caller.NewContractCaller(ec, opPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create operator contract caller: %v", err)
	}

	time.Sleep(time.Duration(6) * time.Second)

	_, err = opCc.ModifyAllocations(
		ctx,
		common.HexToAddress(operatorAddress),
		common.HexToAddress(avsAddress),
		operatorSetId,
		common.HexToAddress(strategyAddress),
		magnitude,
	)
	if err != nil {
		return fmt.Errorf("failed to modify allocations for operator %s: %v", operatorAddress, err)
	}

	return nil
}
