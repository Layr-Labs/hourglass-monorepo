package testUtils

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"strings"
)

func SetupOperatorPeering(
	ctx context.Context,
	chainConfig *ChainConfig,
	chainId config.ChainId,
	ethClient *ethclient.Client,
	aggregatorPrivateBLSKey *bn254.PrivateKey,
	executorPrivateBLSKey *bn254.PrivateKey,
	socket string,
	l *zap.Logger,
) error {
	aggOperatorPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(chainConfig.OperatorAccountPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to convert aggregator operator private key: %v", err)
	}
	aggOperatorAddress := cryptoUtils.DeriveAddress(aggOperatorPrivateKey)
	if !strings.EqualFold(aggOperatorAddress.String(), chainConfig.OperatorAccountAddress) {
		return fmt.Errorf("aggregator operator address mismatch: expected %s, got %s", chainConfig.OperatorAccountAddress, aggOperatorAddress.String())
	}

	// executor operator
	execOperatorPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(chainConfig.ExecOperatorAccountPk)
	if err != nil {
		return fmt.Errorf("failed to convert exec operator private key: %v", err)
	}
	execOperatorAddress := cryptoUtils.DeriveAddress(execOperatorPrivateKey)
	if !strings.EqualFold(execOperatorAddress.String(), chainConfig.ExecOperatorAccountAddress) {
		return fmt.Errorf("executor operator address mismatch: expected %s, got %s", chainConfig.ExecOperatorAccountAddress, execOperatorAddress.String())
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
		aggOperatorAddress,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		[]uint32{0},
		aggregatorPrivateBLSKey,
		"",
		7200,
		"https://some-metadata-uri.com",
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
		execOperatorAddress,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		[]uint32{1},
		executorPrivateBLSKey,
		socket,
		7200,
		"https://some-metadata-uri.com",
		l,
	)
	if err != nil {
		return fmt.Errorf("failed to register executor operator: %v", err)
	}
	l.Sugar().Infow("Executor operator registered successfully",
		zap.String("operatorAddress", aggOperatorAddress.String()),
		zap.String("transactionHash", result.TxHash.String()),
	)
	return nil
}
