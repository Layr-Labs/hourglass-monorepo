package testUtils

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	transactionsigner "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"testing"
)

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
	aggOperatorAddress, err := aggregator.DeriveAddress()
	if err != nil {
		return fmt.Errorf("failed to convert aggregator operator private key: %v", err)
	}

	execOperatorAddress, err := executor.DeriveAddress()
	if err != nil {
		return fmt.Errorf("failed to convert executor operator private key: %v", err)
	}

	avsSigningContext, err := transactionsigner.NewSigningContext(ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create AVS signing context: %v", err)
	}
	avsPrivateKeySigner, err := transactionsigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, avsSigningContext)
	if err != nil {
		return fmt.Errorf("failed to create AVS private key signer: %v", err)
	}

	avsCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
	}, ethClient, avsPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create AVS contract caller: %v", err)
	}

	aggregatorSigningContext, err := transactionsigner.NewSigningContext(ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create aggregator signing context: %v", err)
	}
	aggregatorPrivateKeySigner, err := transactionsigner.NewPrivateKeySigner(chainConfig.OperatorAccountPrivateKey, aggregatorSigningContext)
	if err != nil {
		return fmt.Errorf("failed to create aggregator private key signer: %v", err)
	}

	aggregatorCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
	}, ethClient, aggregatorPrivateKeySigner, l)
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

	executorSigningContext, err := transactionsigner.NewSigningContext(ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create executor signing context: %v", err)
	}
	executorPrivateKeySigner, err := transactionsigner.NewPrivateKeySigner(chainConfig.ExecOperatorAccountPk, executorSigningContext)
	if err != nil {
		return fmt.Errorf("failed to create executor private key signer: %v", err)
	}

	executorCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
	}, ethClient, executorPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create executor contract caller: %v", err)
	}

	l.Sugar().Infow("------------------- Registering executor -------------------")
	if len(executor.OperatorSetIds) == 0 {
		l.Sugar().Infow("No operator sets defined for executor")
		return fmt.Errorf("executor operator sets are empty, cannot register")
	}
	// register the executor
	result, err = operator.RegisterOperatorToOperatorSets(
		ctx,
		avsCc,
		executorCc,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		executor.OperatorSetIds,
		executor,
		&operator.RegistrationConfig{
			Socket:          socket,
			MetadataUri:     "https://some-metadata-uri.com",
			AllocationDelay: 1,
		},
		l,
	)
	if err != nil {
		return fmt.Errorf("failed to register executor operator: %v", err)
	}
	l.Sugar().Infow("Executor operator registered successfully",
		zap.String("operatorAddress", execOperatorAddress.String()),
		zap.String("transactionHash", result.TxHash.String()),
	)
	return nil
}

const (
	Strategy_WETH  = "0x424246eF71b01ee33aA33aC590fd9a0855F5eFbc"
	Strategy_STETH = "0x8b29d91e67b013e855EaFe0ad704aC4Ab086a574"
)

type StakerDelegationConfig struct {
	StakerPrivateKey   string
	StakerAddress      string
	OperatorPrivateKey string
	OperatorAddress    string
	OperatorSetId      uint32
	StrategyAddress    string
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
	stakerSigningContext, err := transactionsigner.NewSigningContext(ethclient, l)
	if err != nil {
		return fmt.Errorf("failed to create staker signing context: %v", err)
	}
	stakerPrivateKeySigner, err := transactionsigner.NewPrivateKeySigner(stakerPrivateKey, stakerSigningContext)
	if err != nil {
		return fmt.Errorf("failed to create staker private key signer: %v", err)
	}

	stakerCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{}, ethclient, stakerPrivateKeySigner, l)
	if err != nil {
		return fmt.Errorf("failed to create staker contract caller: %v", err)
	}

	if _, err := stakerCc.DelegateToOperator(
		ctx,
		common.HexToAddress(stakerAddress),
		common.HexToAddress(operatorAddress),
	); err != nil {
		return fmt.Errorf("failed to delegate stake to operator %s: %v", operatorAddress, err)
	}

	opSigningContext, err := transactionsigner.NewSigningContext(ethclient, l)
	if err != nil {
		return fmt.Errorf("failed to create operator signing context: %v", err)
	}
	opPrivateKeySigner, err := transactionsigner.NewPrivateKeySigner(operatorPrivateKey, opSigningContext)
	if err != nil {
		return fmt.Errorf("failed to create operator private key signer: %v", err)
	}

	opCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{}, ethclient, opPrivateKeySigner, l)
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
