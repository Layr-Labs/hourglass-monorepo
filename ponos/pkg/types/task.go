package types

import (
	"encoding/json"
	"fmt"
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"time"
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

	// Metadata of the task, sourced from the on-chain AVS config
	Metadata []byte `json:"metadata"`
}

type Task struct {
	TaskId              string                      `json:"taskId"`
	AVSAddress          string                      `json:"avsAddress"`
	OperatorSetId       uint32                      `json:"operatorSetId"`
	CallbackAddr        string                      `json:"callbackAddr"`
	RecipientOperators  []*peering.OperatorPeerInfo `json:"recipientOperators"`
	DeadlineUnixSeconds *time.Time                  `json:"deadline"`
	StakeRequired       float64                     `json:"stakeRequired"`
	Payload             []byte                      `json:"payload"`
	ChainId             config.ChainId              `json:"chainId"`
	BlockNumber         uint64                      `json:"blockNumber"`
	BlockHash           string                      `json:"blockHash"`
}

type TaskResult struct {
	TaskId          string
	AvsAddress      string
	CallbackAddr    string
	OperatorSetId   uint32
	Output          []byte
	ChainId         config.ChainId
	BlockNumber     uint64
	BlockHash       string
	OperatorAddress string
	Signature       []byte
}

func TaskResultFromTaskResultProto(tr *aggregatorV1.TaskResult) *TaskResult {
	return &TaskResult{
		TaskId:          tr.TaskId,
		Output:          tr.Output,
		OperatorAddress: tr.OperatorAddress,
		Signature:       tr.Signature,
		AvsAddress:      tr.AvsAddress,
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
		TaskDeadline          uint64
		Payload               []byte
	}
	var od *outputDataType
	if err := json.Unmarshal(outputBytes, &od); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output data: %w", err)
	}
	parsedTaskDeadline := new(big.Int).SetUint64(od.TaskDeadline)
	taskDeadlineTime := time.Now().Add(time.Duration(parsedTaskDeadline.Int64()) * time.Second)

	return &Task{
		TaskId:              taskId,
		AVSAddress:          strings.ToLower(avsAddress),
		OperatorSetId:       od.ExecutorOperatorSetId,
		CallbackAddr:        inboxAddress,
		DeadlineUnixSeconds: &taskDeadlineTime,
		Payload:             []byte(od.Payload),
		ChainId:             block.ChainId,
		BlockNumber:         block.Number.Value(),
		BlockHash:           block.Hash.Value(),
	}, nil
}
