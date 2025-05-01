package tasks

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
