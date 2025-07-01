package testUtils

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
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

	coreContracts, err := config.GetCoreContractsForChainId(chainId)
	if err != nil {
		return fmt.Errorf("failed to get core contracts for chain ID %d: %v", chainId, err)
	}

	avsCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AVSAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
		KeyRegistrarAddress: coreContracts.KeyRegistrar,
	}, ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create AVS contract caller: %v", err)
	}

	aggregatorCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.OperatorAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
		KeyRegistrarAddress: coreContracts.KeyRegistrar,
	}, ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create aggregator contract caller: %v", err)
	}

	l.Sugar().Infow("------------------- Registering aggregator -------------------")
	// register the aggregator
	result, err := operator.RegisterOperatorToOperatorSets(
		ctx,
		avsCc,
		aggregatorCc,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		[]uint32{0},
		aggregator,
		&operator.RegistrationConfig{
			Socket:          "",
			MetadataUri:     "https://some-metadata-uri.com",
			AllocationDelay: 7200,
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

	executorCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.ExecOperatorAccountPk,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
		KeyRegistrarAddress: coreContracts.KeyRegistrar,
	}, ethClient, l)
	if err != nil {
		return fmt.Errorf("failed to create executor contract caller: %v", err)
	}

	l.Sugar().Infow("------------------- Registering executor -------------------")
	// register the executor
	result, err = operator.RegisterOperatorToOperatorSets(
		ctx,
		avsCc,
		executorCc,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		[]uint32{1},
		executor,
		&operator.RegistrationConfig{
			Socket:          socket,
			MetadataUri:     "https://some-metadata-uri.com",
			AllocationDelay: 7200,
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
