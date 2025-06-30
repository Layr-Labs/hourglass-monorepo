package testUtils

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/ethereum/go-ethereum/common"
)

func SetupTaskMailbox(
	ctx context.Context,
	avsAddress common.Address,
	taskHookAddress common.Address,
	executorOperatorSetIds []uint32,
	curveTypes []string,
	// This contractCaller instance needs to be one with the AVSs private key loaded
	avsContractCaller contractCaller.IContractCaller,
) error {
	return avsContractCaller.SetupTaskMailboxForAvs(
		ctx,
		avsAddress,
		taskHookAddress,
		executorOperatorSetIds,
		curveTypes,
	)
}
