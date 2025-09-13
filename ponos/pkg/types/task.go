package types

import (
	"encoding/binary"
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
	TaskId                 string         `json:"taskId"`
	AVSAddress             string         `json:"avsAddress"`
	OperatorSetId          uint32         `json:"operatorSetId"`
	CallbackAddr           string         `json:"callbackAddr"`
	DeadlineUnixSeconds    *time.Time     `json:"deadline"`
	ThresholdBips          uint16         `json:"stakeRequired"`
	Payload                []byte         `json:"payload"`
	ChainId                config.ChainId `json:"chainId"`
	SourceBlockNumber      uint64         `json:"sourceBlockNumber"`
	L1ReferenceBlockNumber uint64         `json:"l1ReferenceBlockNumber"`
	ReferenceTimestamp     uint32         `json:"referenceTimestamp"`
	BlockHash              string         `json:"blockHash"`
	Version                uint32         `json:"version"`
}

type TaskResult struct {
	TaskId          string
	AvsAddress      string
	OperatorSetId   uint32
	Output          []byte
	OperatorAddress string
	ResultSignature []byte // Signs: hash(Output) - for aggregation
	AuthSignature   []byte // Signs: hash(TaskId || AvsAddress || OperatorAddress || OperatorSetId || ResultSigDigest)
}

// AuthSignatureData represents the authentication data that binds operator identity
// Each operator signs a unique message including their address
type AuthSignatureData struct {
	TaskId          string
	AvsAddress      string
	OperatorAddress string
	OperatorSetId   uint32
	ResultSigDigest [32]byte
}

// ToSigningBytes creates deterministic bytes for signing using ABI encoding
// Format: taskId(bytes32) || avsAddress(address) || operatorAddress(address) || operatorSetId(uint32) || resultSigDigest(bytes32)
func (asd *AuthSignatureData) ToSigningBytes() []byte {
	// Convert taskId to 32 bytes
	taskIdBytes := common.HexToHash(asd.TaskId).Bytes() // 32 bytes

	// Convert addresses to 20 bytes, then pad to 32
	avsAddr := common.HexToAddress(asd.AvsAddress).Bytes()       // 20 bytes
	operAddr := common.HexToAddress(asd.OperatorAddress).Bytes() // 20 bytes

	// Convert operatorSetId to 32 bytes (uint32 padded)
	operSetId := make([]byte, 32)
	binary.BigEndian.PutUint32(operSetId[28:], asd.OperatorSetId) // uint32 padded to 32 bytes

	// Concatenate all components
	// Total: 32 + 32 + 32 + 32 + 32 = 160 bytes
	result := make([]byte, 0, 160)
	result = append(result, taskIdBytes...)
	result = append(result, common.LeftPadBytes(avsAddr, 32)...)
	result = append(result, common.LeftPadBytes(operAddr, 32)...)
	result = append(result, operSetId...)
	result = append(result, asd.ResultSigDigest[:]...)

	return result
}

func TaskResultFromTaskResultProto(tr *executorV1.TaskResult) *TaskResult {
	return &TaskResult{
		TaskId:          tr.TaskId,
		Output:          tr.Output,
		OperatorAddress: tr.OperatorAddress,
		ResultSignature: tr.ResultSignature,
		AuthSignature:   tr.AuthSignature,
		AvsAddress:      tr.AvsAddress,
		OperatorSetId:   tr.OperatorSetId,
	}
}

func NewTaskFromLog(log *log.DecodedLog, block *ethereum.EthereumBlock, inboxAddress string) (*Task, error) {
	var avsAddress string
	var taskId string

	// validate log.Arguments length matches the indexed parameters (3 indexed params)
	if len(log.Arguments) < 3 {
		return nil, fmt.Errorf("invalid log arguments length: expected at least 3, got %d", len(log.Arguments))
	}

	taskId, ok := log.Arguments[1].Value.(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse task id")
	}

	avsAddr, ok := log.Arguments[2].Value.(common.Address)
	if !ok {
		return nil, fmt.Errorf("failed to parse task event address")
	}
	avsAddress = avsAddr.String()

	outputBytes, err := json.Marshal(log.OutputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output data: %w", err)
	}

	type outputDataType struct {
		ExecutorOperatorSetId           uint32   `json:"ExecutorOperatorSetId"`
		OperatorTableReferenceTimestamp uint32   `json:"OperatorTableReferenceTimestamp"`
		TaskDeadline                    *big.Int `json:"TaskDeadline"`
		Payload                         []byte   `json:"Payload"`
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
		SourceBlockNumber:   block.Number.Value(),
		ReferenceTimestamp:  od.OperatorTableReferenceTimestamp,
		BlockHash:           block.Hash.Value(),
		Version:             1,
	}, nil
}
