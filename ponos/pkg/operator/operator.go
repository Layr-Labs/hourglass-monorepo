package operator

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

func RegisterOperatorToOperatorSets(
	ctx context.Context,
	avsContractCaller contractCaller.IContractCaller,
	operatorContractCaller contractCaller.IContractCaller,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetIds []uint32,
	privateKey *bn254.PrivateKey,
	socket string,
	allocationDelay uint32,
	metadataUri string,
	l *zap.Logger,
) (*types.Receipt, error) {
	l.Sugar().Infow("Registering operator to AVS operator sets",
		zap.String("avsAddress", avsAddress.String()),
		zap.String("operatorAddress", operatorAddress.String()),
		zap.Uint32s("operatorSetIds", operatorSetIds),
		zap.Any("publicKey", privateKey.Public()),
	)
	keyData, err := operatorContractCaller.EncodeBN254KeyData(privateKey.Public())
	if err != nil {
		l.Sugar().Fatalf("failed to encode BN254 key data: %v", err)
		return nil, fmt.Errorf("failed to get key data: %w", err)
	}

	for _, operatorSetId := range operatorSetIds {
		tx, err := avsContractCaller.ConfigureAVSOperatorSet(ctx, avsAddress, operatorSetId, contractCaller.CurveTypeBN254)
		if err != nil {
			l.Sugar().Fatalf("failed to configure AVS operator set %d: %v", operatorSetId, err)
			return nil, err
		}
		l.Sugar().Infow("Configured AVS operator set",
			zap.String("avsAddress", avsAddress.String()),
			zap.Uint32("operatorSetId", operatorSetId),
			zap.String("txHash", tx.TxHash.String()),
		)

		messageHash, err := operatorContractCaller.GetOperatorRegistrationMessageHash(ctx, operatorAddress, avsAddress, operatorSetId, keyData)
		if err != nil {
			l.Sugar().Fatalf("failed to get operator registration message hash: %v", err)
		}

		sig, err := privateKey.SignSolidityCompatible(messageHash)
		if err != nil {
			l.Sugar().Fatalf("failed to sign message hash: %v", err)
			return nil, err
		}

		l.Sugar().Infow("Registering key for operator set",
			zap.String("avsAddress", avsAddress.String()),
			zap.Uint32("operatorSetId", operatorSetId),
			zap.String("operatorAddress", operatorAddress.String()),
		)

		txReceipt, err := operatorContractCaller.RegisterKeyWithKeyRegistrar(
			ctx,
			operatorAddress,
			avsAddress,
			operatorSetId,
			sig,
			keyData,
		)
		if err != nil {
			l.Sugar().Fatalf("failed to register key with key registrar: %v", err)
			return nil, err
		}
		l.Sugar().Infow("Registered key with registrar",
			zap.String("avsAddress", avsAddress.String()),
			zap.Uint32("operatorSetId", operatorSetId),
			zap.String("operatorAddress", operatorAddress.String()),
			zap.String("transactionHash", txReceipt.TxHash.String()),
		)
	}

	return operatorContractCaller.CreateOperatorAndRegisterWithAvs(
		ctx,
		avsAddress,
		operatorAddress,
		operatorSetIds,
		socket,
		allocationDelay,
		metadataUri,
	)
}
