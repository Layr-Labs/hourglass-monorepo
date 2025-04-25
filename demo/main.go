package main

import (
	"context"
	"encoding/json"
	"fmt"
	performerV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer/server"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type TaskWorker struct {
	logger *zap.Logger
}

func NewTaskWorker(logger *zap.Logger) *TaskWorker {
	return &TaskWorker{
		logger: logger,
	}
}

type TaskRequestPayload struct {
	Message string
}

type TaskResponsePayload struct {
	Message       string
	UnixTimestamp uint64
}

func (tw *TaskWorker) marshalPayload(t *performerV1.Task) (*TaskRequestPayload, error) {
	if len(t.Payload) == 0 {
		return nil, fmt.Errorf("task payload is empty")
	}

	payloadBytes := t.GetPayload()
	var payload *TaskRequestPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal payload")
	}
	return payload, nil
}

func (tw *TaskWorker) ValidateTask(t *performerV1.Task) error {
	tw.logger.Sugar().Infow("Validating task",
		zap.Any("task", t),
	)
	if _, err := tw.marshalPayload(t); err != nil {
		return errors.Wrap(err, "invalid task payload")
	}
	return nil
}

func (tw *TaskWorker) HandleTask(t *performerV1.Task) (*performerV1.TaskResult, error) {
	tw.logger.Sugar().Infow("Handling task",
		zap.Any("task", t),
	)
	payload, err := tw.marshalPayload(t)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal payload")
	}

	responsePayload := &TaskResponsePayload{
		Message:       fmt.Sprintf("Hello %s", payload.Message),
		UnixTimestamp: uint64(time.Now().Unix()),
	}
	responseBytes, err := json.Marshal(responsePayload)

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
		Timeout: 5 * time.Second,
	}, w, l)
	if err != nil {
		panic(fmt.Errorf("failed to create performer: %w", err))
	}

	if err := pp.Start(ctx); err != nil {
		panic(err)
	}
}
