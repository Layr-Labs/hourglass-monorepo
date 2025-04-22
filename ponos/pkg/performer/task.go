package performer

import (
	v1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
)

type Task struct {
	TaskID   string `json:"taskId"`
	Avs      string `json:"avs"`
	Metadata string `json:"metadata"`
	Payload  string `json:"payload"`
}

func (t *Task) GetMetadataBytes() ([]byte, error) {
	return util.DecodeBase64String(t.Metadata)
}

func (t *Task) GetPayloadBytes() ([]byte, error) {
	return util.DecodeBase64String(t.Payload)
}

func NewTaskFromProto(t *v1.TaskSubmission) *Task {
	return &Task{
		TaskID:   t.TaskId,
		Avs:      t.AvsAddress,
		Metadata: "",
		Payload:  util.EncodeBase64String(t.Payload),
	}
}

type TaskResult struct {
	TaskID string `json:"taskId"`
	Avs    string `json:"avs"`
	Result string `json:"result"`
}

func NewTaskResult(taskID string, avs string, result []byte) *TaskResult {
	return &TaskResult{
		TaskID: taskID,
		Avs:    avs,
		Result: util.EncodeBase64String(result),
	}
}
