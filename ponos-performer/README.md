## HTTP/JSON interface

Task payload request
```protobuf
message Task {
	string task_id = 1;
	string avs_address = 2;
	bytes metadata = 3;
	bytes payload = 4;
}
```

Task payload response
```protobuf
message TaskResult {
	string task_id = 1;
	string avs_address = 2;
	bytes result = 3;
}

```

### Example Go implementation

Simply implement the `IWorker` interface and pass it to the `PonosPerformer` server.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer/server"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

// Example task worker that implements the IWorker interface
type TaskWorker struct {
	logger *zap.Logger
}

func NewTaskWorker(logger *zap.Logger) *TaskWorker {
	return &TaskWorker{
		logger: logger,
	}
}

func (tw *TaskWorker) ValidateTask(t *performerV1.Task) error {
	tw.logger.Sugar().Infow("Validating task",
		zap.Any("task", t),
	)

	// ------------------------------------
	// verify payload structure
	// ------------------------------------
	
	
	return nil
}

func (tw *TaskWorker) HandleTask(t *performerV1.Task) (*performerV1.TaskResult, error) {
	tw.logger.Sugar().Infow("Handling task",
		zap.Any("task", t),
	)
	
	// ------------------------------------
	// Logic to handle the task goes here
	// ------------------------------------

	return &performerV1.TaskResult{
		TaskId:     t.TaskId,
		AvsAddress: t.AvsAddress,
		Result:     responseBytes,
	}, nil
}

func main() {
	ctx := context.Background()
	l, _ := zap.NewProduction()

	w := NewTaskWorker(l)

	pp, err := server.NewPonosPerformerWithRpcServer(&server.PonosPerformerConfig{
		Port:    8080,
	}, w, l)
	if err != nil {
		panic(fmt.Errorf("failed to create performer: %w", err))
	}

	if err := pp.Start(ctx); err != nil {
		panic(err)
	}
}
```
