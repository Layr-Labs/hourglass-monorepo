package testUtils

import (
	"context"
	"fmt"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

const (
	Strategy_WETH  = "0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc"
	Strategy_STETH = "0x8b29d91e67b013e855EaFe0ad704aC4Ab086a574"
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
	err := DelegateStakeToOperator(
		ctx,
		aggregatorConfig.StakerPrivateKey,
		aggregatorConfig.StakerAddress,
		aggregatorConfig.OperatorPrivateKey,
		aggregatorConfig.OperatorAddress,
		avsAddress,
		aggregatorConfig.OperatorSetId,
		aggregatorConfig.StrategyAddress,
		ethClient,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to delegate stake to aggregator operator: %v", err)
	}

	t.Logf("------------------------ Delegating Executor ------------------------")
	err = DelegateStakeToOperator(
		ctx,
		executorConfig.StakerPrivateKey,
		executorConfig.StakerAddress,
		executorConfig.OperatorPrivateKey,
		executorConfig.OperatorAddress,
		avsAddress,
		executorConfig.OperatorSetId,
		executorConfig.StrategyAddress,
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
		err := DelegateStakeToOperator(
			ctx,
			config.StakerPrivateKey,
			config.StakerAddress,
			config.OperatorPrivateKey,
			config.OperatorAddress,
			avsAddress,
			config.OperatorSetId,
			config.StrategyAddress,
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
	ethclient *ethclient.Client,
	l *zap.Logger,
) error {
	l.Sugar().Infow("Delegating stake to operator",
		zap.String("stakerPrivateKey", stakerPrivateKey),
		zap.String("operatorPrivateKey", operatorPrivateKey),
		zap.String("operatorAddress", operatorAddress),
		zap.String("avsAddress", avsAddress),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("strategyAddress", strategyAddress),
	)
	stakerPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(stakerPrivateKey, ethclient, l)
	if err != nil {
		return fmt.Errorf("failed to create staker private key signer: %v", err)
	}

	stakerCc, err := caller.NewContractCaller(ethclient, stakerPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create staker contract caller: %v", err)
	}

	if _, err := stakerCc.DelegateToOperator(
		ctx,
		common.HexToAddress(operatorAddress),
	); err != nil {
		return fmt.Errorf("failed to delegate stake to operator %s: %v", operatorAddress, err)
	}

	opPrivateKeySigner, err := transactionSigner.NewPrivateKeySigner(operatorPrivateKey, ethclient, l)
	if err != nil {
		return fmt.Errorf("failed to create operator private key signer: %v", err)
	}

	opCc, err := caller.NewContractCaller(ethclient, opPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create operator contract caller: %v", err)
	}

	_, err = opCc.ModifyAllocations(
		ctx,
		common.HexToAddress(operatorAddress),
		common.HexToAddress(avsAddress),
		operatorSetId,
		common.HexToAddress(strategyAddress),
	)
	if err != nil {
		return fmt.Errorf("failed to modify allocations for operator %s: %v", operatorAddress, err)
	}
	return nil
}
