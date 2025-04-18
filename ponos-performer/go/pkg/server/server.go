package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos-performer/go/pkg/worker"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type PonosPerformerConfig struct {
	Port    int
	Timeout time.Duration
}

type PonosPerformer struct {
	config     *PonosPerformerConfig
	taskWorker worker.IWorker
	logger     *zap.Logger
}

func NewPonosPerformer(
	cfg *PonosPerformerConfig,
	worker worker.IWorker,
	logger *zap.Logger,
) *PonosPerformer {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &PonosPerformer{
		config:     cfg,
		taskWorker: worker,
		logger:     logger,
	}
}

func (pp *PonosPerformer) WriteJsonError(w http.ResponseWriter, err error, errorCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); err != nil {
		pp.logger.Sugar().Errorw("Failed to write JSON error response",
			zap.Error(err),
		)
	}
}

func (pp *PonosPerformer) WriteJsonResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		pp.logger.Sugar().Errorw("Failed to write JSON response",
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (pp *PonosPerformer) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pp.logger.Sugar().Infow("Received HTTP request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
		)
		next.ServeHTTP(w, r)
	})
}

func (pp *PonosPerformer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := struct {
		Status string `json:"status"`
	}{
		Status: "running",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := pp.WriteJsonResponse(w, health); err != nil {
		pp.WriteJsonError(w, fmt.Errorf("failed to write JSON health response - %v", err), http.StatusInternalServerError)
	}
}

func (pp *PonosPerformer) handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		pp.WriteJsonError(w, fmt.Errorf("invalid request method"), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		pp.WriteJsonError(w, fmt.Errorf("failed to read request body - %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var task *performer.Task
	if err := json.Unmarshal(body, &task); err != nil {
		pp.WriteJsonError(w, fmt.Errorf("failed to parse task json from body - %v", err), http.StatusBadRequest)
		return
	}

	if err := pp.taskWorker.ValidateTask(task); err != nil {
		pp.logger.Sugar().Errorw("Task is invalid",
			zap.String("taskId", task.TaskID),
			zap.String("avs", task.Avs),
			zap.Uint64("operatorSetId", task.OperatorSetID),
			zap.Error(err),
		)
		pp.WriteJsonError(w, fmt.Errorf("task is invalid - %v", err), http.StatusBadRequest)
		return
	}

	result, err := pp.taskWorker.HandleTask(task)
	if err != nil {
		pp.logger.Sugar().Errorw("Failed to handle task",
			zap.String("taskId", task.TaskID),
			zap.String("avs", task.Avs),
			zap.Uint64("operatorSetId", task.OperatorSetID),
			zap.Error(err),
		)
		pp.WriteJsonError(w, fmt.Errorf("failed to handle task - %v", err), http.StatusInternalServerError)
		return
	}
	resultString, err := json.Marshal(result)
	if err != nil {
		pp.logger.Sugar().Errorw("Failed to marshal task result",
			zap.String("taskId", task.TaskID),
			zap.String("avs", task.Avs),
			zap.Uint64("processType", task.OperatorSetID),
			zap.Error(err),
		)
		pp.WriteJsonError(w, fmt.Errorf("failed to marshal task result - %v", err), http.StatusInternalServerError)
		return
	}

	pp.WriteJsonResponse(w, resultString)
}

func (pp *PonosPerformer) StartHttpServer(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", pp.handleTask)
	mux.HandleFunc("/health", pp.handleHealth)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", pp.config.Port),
		Handler:      pp.loggerMiddleware(mux),
		ReadTimeout:  pp.config.Timeout,
		WriteTimeout: pp.config.Timeout,
		IdleTimeout:  pp.config.Timeout,
	}

	go func() {
		pp.logger.Sugar().Infow("Starting HTTP server", zap.Int("port", pp.config.Port))
		if err := server.ListenAndServe(); err != nil {
			pp.logger.Sugar().Errorw("Failed to start HTTP server",
				zap.Int("port", pp.config.Port),
				zap.Error(err),
			)
			return
		}
	}()

	<-ctx.Done()
	pp.logger.Sugar().Infow("Shutting down HTTP server")
	_ = server.Shutdown(context.Background())
	return nil
}
