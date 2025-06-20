package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer/server"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"go.uber.org/zap"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

type TaskWorker struct {
	logger    *zap.Logger
	taskDelay time.Duration
}

func NewTaskWorker(logger *zap.Logger, delay time.Duration) *TaskWorker {
	return &TaskWorker{
		logger:    logger,
		taskDelay: delay,
	}
}

func parseHexBytesToBigInt(payload []byte) (*big.Int, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("payload is empty")
	}

	payloadStr := strings.TrimPrefix(string(payload), "0x")

	i, success := new(big.Int).SetString(payloadStr, 16)
	if !success {
		return nil, fmt.Errorf("failed to convert hex string to big.Int")
	}
	return i, nil
}

func parseBigIntToHex(i *big.Int) []byte {
	if i == nil {
		return nil
	}
	hexStr := i.Text(16)
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	return []byte("0x" + hexStr)
}

func (tw *TaskWorker) ValidateTask(t *performerV1.TaskRequest) error {
	tw.logger.Sugar().Infow("Validating task",
		zap.Any("task", t),
	)
	_, err := parseHexBytesToBigInt(t.Payload)

	return err
}

func (tw *TaskWorker) HandleTask(t *performerV1.TaskRequest) (*performerV1.TaskResponse, error) {
	tw.logger.Sugar().Infow("Handling task",
		zap.Any("task", t),
	)
	time.Sleep(tw.taskDelay)

	i, err := parseHexBytesToBigInt(t.Payload)
	if err != nil {
		return nil, err
	}

	squaredNumber := new(big.Int).Exp(i, big.NewInt(2), nil)

	tw.logger.Sugar().Infow("Task result",
		zap.Uint64("originalInput", i.Uint64()),
		zap.Uint64("squaredResult", squaredNumber.Uint64()),
	)

	return &performerV1.TaskResponse{
		TaskId: t.TaskId,
		Result: parseBigIntToHex(squaredNumber),
	}, nil
}

func main() {
	ctx := context.Background()
	l, _ := zap.NewProduction()

	// Read task delay from environment variable
	var taskDelay time.Duration
	delayMs := os.Getenv("TASK_DELAY_MS")
	l.Sugar().Infow("Checking TASK_DELAY_MS environment variable",
		zap.String("TASK_DELAY_MS", delayMs),
		zap.Bool("isSet", delayMs != ""),
	)

	if delayMs != "" {
		if ms, err := strconv.Atoi(delayMs); err == nil {
			taskDelay = time.Duration(ms) * time.Millisecond
			l.Sugar().Infow("Task delay configured",
				zap.Duration("delay", taskDelay),
			)
		} else {
			l.Sugar().Warnw("Failed to parse TASK_DELAY_MS",
				zap.String("value", delayMs),
				zap.Error(err),
			)
		}
	}

	w := NewTaskWorker(l, taskDelay)

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
