package tasks

import "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"

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
