package simulatedChainListener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"io"
	"net/http"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/workQueue"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"go.uber.org/zap"
)

type SimulatedChainListenerConfig struct {
	Port int
}

type SimulatedChainListener struct {
	taskQueue  workQueue.IInputQueue[types.Task]
	config     *SimulatedChainListenerConfig
	logger     *zap.Logger
	httpServer *http.Server
	chainId    *config.ChainId
}

func NewSimulatedChainListener(
	taskQueue workQueue.IInputQueue[types.Task],
	config *SimulatedChainListenerConfig,
	logger *zap.Logger,
	chainId *config.ChainId,
) *SimulatedChainListener {
	return &SimulatedChainListener{
		taskQueue: taskQueue,
		config:    config,
		logger:    logger,
		chainId:   chainId,
	}
}

func (scl *SimulatedChainListener) Start(ctx context.Context) error {
	scl.logger.Sugar().Infow("SimulatedChainListener starting", zap.Int("port", scl.config.Port))

	mux := http.NewServeMux()
	mux.HandleFunc("/events", scl.handleSubmitTaskRoute())

	scl.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", scl.config.Port),
		Handler: scl.httpLoggerMiddleware(mux),
	}

	go func() {
		if err := scl.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			scl.logger.Sugar().Errorw("HTTP server error", zap.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()
		_ = scl.Close()
	}()

	return nil
}

func (scl *SimulatedChainListener) Close() error {
	scl.logger.Sugar().Infow("SimulatedChainListener stopping")
	if scl.httpServer != nil {
		return scl.httpServer.Shutdown(context.Background())
	}
	return nil
}

func (scl *SimulatedChainListener) httpLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scl.logger.Sugar().Infow("Received HTTP request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
		)
		next.ServeHTTP(w, r)
	})
}

func (scl *SimulatedChainListener) handleSubmitTaskRoute() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			scl.logger.Sugar().Errorw("Failed to read request body", zap.Error(err))
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var taskEvent types.TaskEvent
		if err := json.Unmarshal(body, &taskEvent); err != nil {
			scl.logger.Sugar().Errorw("Failed to unmarshal task event", zap.Error(err))
			http.Error(w, "Failed to unmarshal task event", http.StatusBadRequest)
			return
		}

		task := convertEventToTask(&taskEvent, scl.chainId)
		scl.logger.Sugar().Infow("Received simulated task event", "taskID", task.TaskId)

		if err := scl.taskQueue.Enqueue(task); err != nil {
			scl.logger.Sugar().Errorw("Failed to enqueue task event", zap.Error(err))
			http.Error(w, "Failed to enqueue task event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Task event enqueued"))
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
