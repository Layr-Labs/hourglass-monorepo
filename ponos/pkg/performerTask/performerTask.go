package performerTask

import (
	v1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
)

type PerformerTask struct {
	TaskID             string
	Avs                string
	Payload            []byte
	Signature          []byte
	AggregatorAddress  string
	OperatorSetId      uint32
	ReferenceTimestamp uint32
}

// NewPerformerTaskFromTaskSubmissionProto creates a new PerformerTask from a TaskSubmission proto
func NewPerformerTaskFromTaskSubmissionProto(t *v1.TaskSubmission) *PerformerTask {
	return &PerformerTask{
		TaskID:             t.TaskId,
		Avs:                t.AvsAddress,
		Payload:            t.Payload,
		Signature:          t.Signature,
		AggregatorAddress:  t.AggregatorAddress,
		OperatorSetId:      t.OperatorSetId,
		ReferenceTimestamp: t.ReferenceTimestamp,
	}
}

type PerformerTaskResult struct {
	TaskID string `json:"taskId"`
	Result []byte `json:"result"`
}

func NewPerformerTaskResult(taskID string, result []byte) *PerformerTaskResult {
	return &PerformerTaskResult{
		TaskID: taskID,
		Result: result,
	}
}

func NewTaskResultFromResultProto(result *performerV1.TaskResponse) *PerformerTaskResult {
	return &PerformerTaskResult{
		TaskID: string(result.TaskId),
		Result: result.Result,
	}
}
