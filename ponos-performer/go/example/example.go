package main

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos-performer/go/pkg/server"
	"github.com/Layr-Labs/hourglass-monorepo/ponos-performer/go/pkg/task"
	"go.uber.org/zap"
	"time"
)

type TaskWorker struct{}

func (tw *TaskWorker) HandleTask(t *task.Task) (*task.TaskResult, error) {
	return task.NewTaskResult(t.TaskID, t.Avs, t.OperatorSetID, []byte("result")), nil
}

func main() {
	ctx := context.Background()
	l, _ := zap.NewProduction()

	w := &TaskWorker{}

	pp := server.NewPonosPerformer(&server.PonosPerformerConfig{
		Port:    8080,
		Timeout: 5 * time.Second,
	}, w, l)

	if err := pp.StartHttpServer(ctx); err != nil {
		panic(err)
	}
}
