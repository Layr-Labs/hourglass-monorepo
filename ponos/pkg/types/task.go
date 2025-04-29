package types

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Task struct {
	TaskId        string
	AVSAddress    string
	OperatorSetId uint32
	CallbackAddr  string
	Deadline      *big.Int
	StakeRequired float64
	Payload       []byte
	ChainId       config.ChainId
	BlockNumber   uint64
	BlockHash     string
}

type TaskResult struct {
	TaskId        string
	AvsAddress    string
	CallbackAddr  string
	OperatorSetId uint32
	Output        []byte
	ChainId       config.ChainId
	BlockNumber   uint64
	BlockHash     string
}

func NewTask(
	log *transactionLogParser.DecodedLog,
	block *ethereum.EthereumBlock,
	inboxAddress string,
	chainId config.ChainId,
) (*Task, error) {
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
	operatorSetId, ok = log.OutputData["operatorSetId"].(uint32)
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

	return &Task{
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
