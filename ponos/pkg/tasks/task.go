package tasks

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"time"
)

type Task struct {
	TaskId              string         `json:"taskId"`
	AVSAddress          string         `json:"avsAddress"`
	OperatorSetId       uint32         `json:"operatorSetId"`
	CallbackAddr        string         `json:"callbackAddr"`
	DeadlineUnixSeconds *time.Time     `json:"deadline"`
	StakeRequired       float64        `json:"stakeRequired"`
	Payload             []byte         `json:"payload"`
	ChainId             config.ChainId `json:"chainId"`
	BlockNumber         uint64         `json:"blockNumber"`
	BlockHash           string         `json:"blockHash"`
}
