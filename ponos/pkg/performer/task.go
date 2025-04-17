package performer

import "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"

type Task struct {
	TaskID        string `json:"taskId"`
	Avs           string `json:"avs"`
	OperatorSetID uint64 `json:"operatorSetId"`
	ChainId       string `json:"chainId"`
	Metadata      string `json:"metadata"`
	Payload       string `json:"payload"`
}

func (t *Task) GetMetadataBytes() ([]byte, error) {
	return util.DecodeBase64String(t.Metadata)
}

func (t *Task) GetPayloadBytes() ([]byte, error) {
	return util.DecodeBase64String(t.Payload)
}

type TaskResult struct {
	TaskID        string `json:"taskId"`
	Avs           string `json:"avs"`
	OperatorSetID uint64 `json:"operatorSetId"`
	Result        string `json:"result"`
}

func NewTaskResult(taskID string, avs string, operatorSetID uint64, result []byte) *TaskResult {
	return &TaskResult{
		TaskID:        taskID,
		Avs:           avs,
		OperatorSetID: operatorSetID,
		Result:        util.EncodeBase64String(result),
	}
}
