package types

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/ethereum/go-ethereum/common"
)

// TaskEvent is a struct that represents a task event as consumed from on-chain events
type TaskEvent struct {
	// The address of who created the task
	CreatorAddress string `json:"creatorAddress"`

	// Unique hash of task metadata to identify the task globally
	TaskId string `json:"taskId"`

	// Address of the AVS
	AVSAddress string `json:"avsAddress"`

	// The ID of the operator set to distribute the task to
	OperatorSetId uint32 `json:"operatorSetId"`

	// The payload of the task
	Payload []byte `json:"payload"`
}

type Task struct {
	TaskId              string         `json:"taskId"`
	AVSAddress          string         `json:"avsAddress"`
	OperatorSetId       uint32         `json:"operatorSetId"`
	CallbackAddr        string         `json:"callbackAddr"`
	DeadlineUnixSeconds *time.Time     `json:"deadline"`
	ThresholdBips       uint16         `json:"stakeRequired"`
	Payload             []byte         `json:"payload"`
	ChainId             config.ChainId `json:"chainId"`
	BlockNumber         uint64         `json:"blockNumber"`
	BlockHash           string         `json:"blockHash"`
}

type TaskResult struct {
	TaskId          string
	AvsAddress      string
	OperatorSetId   uint32
	Output          []byte
	OperatorAddress string
	Signature       []byte
	OutputDigest    []byte
}

func TaskResultFromTaskResultProto(tr *executorV1.TaskResult) *TaskResult {
	return &TaskResult{
		TaskId:          tr.TaskId,
		Output:          tr.Output,
		OperatorAddress: tr.OperatorAddress,
		Signature:       tr.Signature,
		AvsAddress:      tr.AvsAddress,
		OutputDigest:    tr.OutputDigest,
	}
}

func NewTaskFromLog(log *log.DecodedLog, block *ethereum.EthereumBlock, inboxAddress string) (*Task, error) {
	var avsAddress string
	var taskId string

	taskId, ok := log.Arguments[1].Value.(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse task id")
	}

	avsAddr, ok := log.Arguments[2].Value.(common.Address)
	if !ok {
		return nil, fmt.Errorf("failed to parse task event address")
	}
	avsAddress = avsAddr.String()

	// it aint stupid if it works...
	// take the output data, turn it into a json string, then Unmarshal it into a typed struct
	// rather than trying to coerce data types
	outputBytes, err := json.Marshal(log.OutputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output data: %w", err)
	}

	type outputDataType struct {
		ExecutorOperatorSetId uint32
		TaskDeadline          *big.Int `json:"TaskDeadline"`
		Payload               []byte
	}
	var od *outputDataType
	if err := json.Unmarshal(outputBytes, &od); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output data: %w", err)
	}
	if od.TaskDeadline.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		return nil, fmt.Errorf("task deadline too large for duration calculation: %s", od.TaskDeadline.String())
	}
	taskDeadlineTime := time.Now().Add(time.Duration(od.TaskDeadline.Uint64()) * time.Second)

	return &Task{
		TaskId:              taskId,
		AVSAddress:          strings.ToLower(avsAddress),
		OperatorSetId:       od.ExecutorOperatorSetId,
		CallbackAddr:        inboxAddress,
		DeadlineUnixSeconds: &taskDeadlineTime,
		Payload:             od.Payload,
		ChainId:             block.ChainId,
		BlockNumber:         block.Number.Value(),
		BlockHash:           block.Hash.Value(),
	}, nil
}
