package convert

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func ConvertTask(
	log *log.DecodedLog,
	block *ethereum.EthereumBlock,
	inboxAddress string,
	chainId config.ChainId,
) (*types.Task, error) {
	var avsAddress common.Address
	var operatorSetId uint32
	var taskDeadline *big.Int
	var taskId string
	var payload []byte

	taskId, ok := log.Arguments[1].Value.(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse task id")
	}
	avsAddress, ok = log.Arguments[2].Value.(common.Address)
	if !ok {
		return nil, fmt.Errorf("failed to parse task event address")
	}
	operatorSetId, ok = log.OutputData["executorOperatorSetId"].(uint32)
	if !ok {
		return nil, fmt.Errorf("failed to parse operator set id")
	}
	taskDeadline, ok = log.OutputData["taskDeadline"].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to parse task event deadline")
	}
	payload, ok = log.OutputData["payload"].([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to parse task payload")
	}

	return &types.Task{
		TaskId:        taskId,
		AVSAddress:    avsAddress.String(),
		OperatorSetId: operatorSetId,
		CallbackAddr:  inboxAddress,
		Deadline:      taskDeadline,
		Payload:       payload,
		ChainId:       chainId,
		BlockNumber:   block.Number.Value(),
		BlockHash:     block.Hash.Value(),
	}, nil
}
