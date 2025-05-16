package operator

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

func RegisterOperatorToOperatorSets(
	ctx context.Context,
	caller contractCaller.IContractCaller,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetIds []uint32,
	operatorPublicBlsKey *bn254.PublicKey,
	operatorPrivateBlsKey *bn254.PrivateKey,
	socket string,
	allocationDelay uint32,
	metadataUri string,
	l *zap.Logger,
) (*types.Receipt, error) {
	g1Point, err := caller.GetOperatorRegistrationMessageHash(ctx, operatorAddress)
	if err != nil {
		l.Sugar().Fatalf("failed to get operator registration message hash: %v", err)
	}

	// Create G1 point from contract coordinates
	hashPoint := bn254.NewG1Point(g1Point.X, g1Point.Y)

	// Sign the hash point
	signature, err := operatorPrivateBlsKey.SignG1Point(hashPoint.G1Affine)
	if err != nil {
		l.Sugar().Fatalf("failed to sign hash point: %v", err)
	}

	return caller.CreateOperatorAndRegisterWithAvs(
		ctx,
		avsAddress,
		operatorAddress,
		operatorSetIds,
		operatorPublicBlsKey,
		signature,
		socket,
		allocationDelay,
		metadataUri,
	)
}
