package manualPushChainPoller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"io"
	"net/http"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"go.uber.org/zap"
)

type ManualPushChainPollerConfig struct {
	ChainId *config.ChainId
	Port    int
}

type ManualPushChainPoller struct {
	chainEventsChan chan *chainPoller.LogWithBlock
	httpServer      *http.Server
	config          *ManualPushChainPollerConfig
	logger          *zap.Logger
}

func NewManualPushChainPoller(
	chainEventsChan chan *chainPoller.LogWithBlock,
	config *ManualPushChainPollerConfig,
	logger *zap.Logger,
) *ManualPushChainPoller {
	return &ManualPushChainPoller{
		chainEventsChan: chainEventsChan,
		config:          config,
		logger:          logger,
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

		var lwb *chainPoller.LogWithBlock
		if err := json.Unmarshal(body, &lwb); err != nil {
			scl.logger.Sugar().Errorw("Failed to unmarshal task event", "error", err)
			http.Error(w, "Failed to unmarshal task event", http.StatusBadRequest)
			return
		}

		scl.logger.Sugar().Infow("Received task event",
			zap.Any("event", lwb),
		)

		select {
		case scl.chainEventsChan <- lwb:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("task event enqueued"))
		case <-time.After(1 * time.Second):
			scl.logger.Sugar().Errorw("Failed to enqueue task (channel full or closed)",
				zap.Any("lwb", lwb),
				zap.Error(err),
			)
			http.Error(w, "Failed to enqueue task", http.StatusInternalServerError)
		case <-ctx.Done():
			http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		}
	}
}
