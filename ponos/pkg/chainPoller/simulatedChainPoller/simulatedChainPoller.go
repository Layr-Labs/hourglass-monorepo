package simulatedChainPoller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

type SimulatedChainListenerConfig struct {
	ChainId         *config.ChainId
	Port            int
	PollingInterval time.Duration
}

type SimulatedChainListener struct {
	taskQueue  chan *types.Task
	config     *SimulatedChainListenerConfig
	logger     *zap.Logger
	httpServer *http.Server
}

func NewSimulatedChainListener(
	taskQueue chan *types.Task,
	config *SimulatedChainListenerConfig,
	logger *zap.Logger,
) *SimulatedChainListener {
	return &SimulatedChainListener{
		taskQueue: taskQueue,
		config:    config,
		logger:    logger,
	}
}

func (scl *SimulatedChainListener) Start(ctx context.Context) error {
	sugar := scl.logger.Sugar()
	sugar.Infow("SimulatedChainListener starting", "port", scl.config.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/events", scl.handleSubmitTaskRoute(ctx))

	scl.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", scl.config.Port),
		Handler: scl.httpLoggerMiddleware(mux),
	}

	go func() {
		if err := scl.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sugar.Errorw("HTTP server error", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		sugar.Infow("SimulatedChainListener stopping due to context cancellation")
		if scl.httpServer != nil {
			err := scl.httpServer.Shutdown(context.Background())
			if err != nil {
				sugar.Errorw("HTTP server shutdown error", "error", err)
			}
		}
	}()

	if scl.config.PollingInterval > 0 {
		go func() {
			scl.generatePeriodicTasks(ctx)
		}()
	} else {
		return fmt.Errorf("polling interval must be greater than 0")
	}

	return nil
}

func (scl *SimulatedChainListener) generatePeriodicTasks(ctx context.Context) {
	ticker := time.NewTicker(scl.config.PollingInterval)
	defer ticker.Stop()

	sugar := scl.logger.Sugar()
	sugar.Infow("Starting periodic task generation")

	for {
		select {
		case <-ctx.Done():
			sugar.Infow("Stopping periodic task generation")
			return
		case <-ticker.C:
			task := &types.Task{
				TaskId:        fmt.Sprintf("periodic-task-%d", time.Now().UnixNano()),
				AVSAddress:    "0xPeriodicTaskAVS",
				OperatorSetId: 123456,
				CallbackAddr:  "0xPeriodicTaskCallback",
				Payload:       []byte(`{"type":"periodic","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `}`),
				Deadline:      time.Now().Add(1 * time.Hour).UnixMilli(),
				StakeRequired: 0.75,
				ChainId:       *scl.config.ChainId,
			}

			select {
			case scl.taskQueue <- task:
				sugar.Infow("Generated periodic task", "taskID", task.TaskId)
			case <-time.After(1 * time.Second):
				sugar.Warnw("Failed to enqueue periodic task (channel full or closed)", "taskID", task.TaskId)
			case <-ctx.Done():
				return
			}
		}
	}
}

func (scl *SimulatedChainListener) httpLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scl.logger.Sugar().Infow("Received HTTP request",
			"method", r.Method,
			"url", r.URL.String(),
		)
		next.ServeHTTP(w, r)
	})
}

func (scl *SimulatedChainListener) handleSubmitTaskRoute(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			scl.logger.Sugar().Errorw("Failed to read request body", "error", err)
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var taskEvent types.TaskEvent
		if err := json.Unmarshal(body, &taskEvent); err != nil {
			scl.logger.Sugar().Errorw("Failed to unmarshal task event", "error", err)
			http.Error(w, "Failed to unmarshal task event", http.StatusBadRequest)
			return
		}

		task := convertEventToTask(&taskEvent, scl.config.ChainId)
		scl.logger.Sugar().Infow("Received simulated task event", "taskID", task.TaskId)

		select {
		case scl.taskQueue <- task:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Task event enqueued"))
		case <-time.After(1 * time.Second):
			scl.logger.Sugar().Errorw("Failed to enqueue task (channel full or closed)", "taskID", task.TaskId)
			http.Error(w, "Failed to enqueue task", http.StatusInternalServerError)
		case <-ctx.Done():
			http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		}
	}
}

func convertEventToTask(event *types.TaskEvent, chainId *config.ChainId) *types.Task {
	var parsedMeta struct {
		Deadline               int64   `json:"deadline"`
		StakeWeightRequiredPct float64 `json:"stakeWeightRequiredPct"`
	}
	_ = json.Unmarshal([]byte(event.Metadata), &parsedMeta)

	return &types.Task{
		TaskId:        event.TaskId,
		AVSAddress:    event.AVSAddress,
		OperatorSetId: event.OperatorSetId,
		CallbackAddr:  event.CallbackAddr,
		Payload:       event.Payload,
		Deadline:      time.Now().Unix() + parsedMeta.Deadline,
		StakeRequired: parsedMeta.StakeWeightRequiredPct,
		ChainId:       *chainId,
	}
}
