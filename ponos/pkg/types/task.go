package types

import "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"

type TaskEvent struct {
	TaskId        string         `json:"taskId"`
	AVSAddress    string         `json:"avsAddress"`
	OperatorSetId uint32         `json:"operatorSetId"`
	CallbackAddr  string         `json:"callbackAddress"`
	Metadata      string         `json:"metadata"`
	Payload       []byte         `json:"payload"`
	ChainID       config.ChainId `json:"chainId"`
}

type Task struct {
	TaskId        string
	AVSAddress    string
	OperatorSetId uint32
	CallbackAddr  string
	Deadline      int64
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
