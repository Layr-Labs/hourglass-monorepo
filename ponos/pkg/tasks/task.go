package tasks

import (
	v1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	performerV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/performer"
)

type Task struct {
	TaskID   string
	Avs      string
	Metadata []byte
	Payload  []byte
}

func NewTaskFromTaskSubmissionProto(t *v1.TaskSubmission) *Task {
	return &Task{
		TaskID:   t.TaskId,
		Avs:      t.AvsAddress,
		Metadata: []byte{},
		Payload:  t.Payload,
	}
}

type TaskResult struct {
	TaskID string `json:"taskId"`
	Avs    string `json:"avs"`
	Result []byte `json:"result"`
}

func NewTaskResult(taskID string, avs string, result []byte) *TaskResult {
	return &TaskResult{
		TaskID: taskID,
		Avs:    avs,
		Result: result,
	}
}

func NewTaskResultFromResultProto(result *performerV1.TaskResult) *TaskResult {
	return &TaskResult{
		TaskID: result.TaskId,
		Avs:    result.AvsAddress,
		Result: result.Result,
	}
}
