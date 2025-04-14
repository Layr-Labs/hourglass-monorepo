package simulatedChainListener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainListener"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type SimulatedChainListenerConfig struct {
	Port int
}

// SimulatedChainListener implements the chain listener interface but doesnt actually listen to a chain.
//
// Instead, it exposed an HTTP server that allows the developer to manually push event data
// as if it were coming from the chain to make testing easier.
type SimulatedChainListener struct {
	config *SimulatedChainListenerConfig
	logger *zap.Logger
}

func NewSimulatedChainListener(
	config *SimulatedChainListenerConfig,
	logger *zap.Logger,
) *SimulatedChainListener {
	return &SimulatedChainListener{
		config: config,
		logger: logger,
	}
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

func (scl *SimulatedChainListener) handleEventsRoute(queue chan *chainListener.Event) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			scl.logger.Sugar().Errorw("Failed to read request body",
				zap.Error(err),
			)
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var event *chainListener.Event
		if err := json.Unmarshal(body, &event); err != nil {
			scl.logger.Sugar().Errorw("Failed to unmarshal event",
				zap.Error(err),
			)
			http.Error(w, "Failed to unmarshal event", http.StatusBadRequest)
			return
		}

		fmt.Printf("Received event: %+v\n", event)
		queue <- event

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Event published to queue"))
	}
}

func (scl *SimulatedChainListener) ListenForInboxEvents(
	ctx context.Context,
	queue chan *chainListener.Event,
	chainId string,
) error {
	scl.logger.Sugar().Infow("Simulated chain listener started",
		zap.Int("port", scl.config.Port),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/events", scl.handleEventsRoute(queue))
	handler := scl.httpLoggerMiddleware(mux)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", scl.config.Port), handler); err != nil {
			scl.logger.Sugar().Errorw("Failed to start HTTP server",
				zap.Int("port", scl.config.Port),
				zap.Error(err),
			)
			cancelCtx, ok := ctx.Value("cancelFunc").(context.CancelFunc)
			if ok {
				scl.logger.Sugar().Infow("Cancelling context due to HTTP server error")
				cancelCtx()
			}
		}
	}()
	select {
	case <-ctx.Done():
		scl.logger.Sugar().Infow("Context done, stopping Ethereum Chain Listener")
		return nil
	}
}
