package manualPushChainPoller

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

type ManualPushChainPollerConfig struct {
	ChainId *config.ChainId
	Port    int
}

type ManualPushChainPoller struct {
	taskQueue  chan *types.Task
	httpServer *http.Server
	config     *ManualPushChainPollerConfig
	logger     *zap.Logger
}

func NewManualPushChainPoller(
	taskQueue chan *types.Task,
	config *ManualPushChainPollerConfig,
	logger *zap.Logger,
) *ManualPushChainPoller {
	return &ManualPushChainPoller{
		taskQueue: taskQueue,
		config:    config,
		logger:    logger,
	}
}

func (scl *ManualPushChainPoller) Start(ctx context.Context) error {
	sugar := scl.logger.Sugar()
	sugar.Infow("ManualPushChainPoller starting", "port", scl.config.Port)

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
		sugar.Infow("ManualPushChainPoller stopping due to context cancellation")
		if scl.httpServer != nil {
			err := scl.httpServer.Shutdown(context.Background())
			if err != nil {
				sugar.Errorw("HTTP server shutdown error", "error", err)
			}
		}
	}()

	return nil
}

func (scl *ManualPushChainPoller) httpLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scl.logger.Sugar().Infow("Received HTTP request",
			"method", r.Method,
			"url", r.URL.String(),
		)
		next.ServeHTTP(w, r)
	})
}

func (scl *ManualPushChainPoller) handleSubmitTaskRoute(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
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
			_, _ = w.Write([]byte("task event enqueued"))
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
	_ = json.Unmarshal(event.Metadata, &parsedMeta)
	deadline := time.Now().Add(time.Duration(parsedMeta.Deadline) * time.Second)
	return &types.Task{
		TaskId:              event.TaskId,
		AVSAddress:          event.AVSAddress,
		OperatorSetId:       event.OperatorSetId,
		Payload:             event.Payload,
		DeadlineUnixSeconds: &deadline,
		StakeRequired:       parsedMeta.StakeWeightRequiredPct,
		ChainId:             *chainId,
	}
}
