package types

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"math/big"
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
	TaskId        string         `json:"taskId"`
	AVSAddress    string         `json:"avsAddress"`
	OperatorSetId uint32         `json:"operatorSetId"`
	CallbackAddr  string         `json:"callbackAddr"`
	Deadline      *big.Int       `json:"deadline"`
	StakeRequired float64        `json:"stakeRequired"`
	Payload       []byte         `json:"payload"`
	ChainId       config.ChainId `json:"chainId"`
	BlockNumber   uint64         `json:"blockNumber"`
	BlockHash     string         `json:"blockHash"`
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
