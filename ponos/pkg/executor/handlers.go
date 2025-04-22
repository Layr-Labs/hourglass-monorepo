package executor

import (
	"context"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/common/v1"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
)

func (e *Executor) SubmitTask(_ context.Context, _ *executorV1.TaskSubmission) (*commonV1.SubmitAck, error) {
	return &commonV1.SubmitAck{Message: "Stubbed message", Success: false}, nil
}
