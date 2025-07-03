package gateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hourglass/obsidian/api/proto/orchestrator"
	"github.com/hourglass/obsidian/api/proto/proxy"
	"github.com/hourglass/obsidian/api/proto/registry"
	orchSvc "github.com/hourglass/obsidian/pkg/orchestrator"
	proxySvc "github.com/hourglass/obsidian/pkg/proxy"
	regSvc "github.com/hourglass/obsidian/pkg/registry"
	"github.com/hourglass/obsidian/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Gateway struct {
	config           *config.Config
	grpcServer       *grpc.Server
	httpServer       *http.Server
	orchestrator     *orchSvc.Server
	registry         *regSvc.Server
	proxy            *proxySvc.Server
}

func NewGateway(config *config.Config) (*Gateway, error) {
	orchestrator, err := orchSvc.NewServer(&config.Orchestrator)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}

	registry, err := regSvc.NewServer(&config.Registry)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	proxy, err := proxySvc.NewServer(&config.Proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy: %w", err)
	}

	g := &Gateway{
		config:       config,
		orchestrator: orchestrator,
		registry:     registry,
		proxy:        proxy,
	}

	g.setupGRPCServer()
	g.setupHTTPServer()

	return g, nil
}

func (g *Gateway) setupGRPCServer() {
	g.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(10 * 1024 * 1024),
		grpc.MaxSendMsgSize(10 * 1024 * 1024),
	)

	orchestrator.RegisterOrchestratorServiceServer(g.grpcServer, g.orchestrator)
	registry.RegisterRegistryServiceServer(g.grpcServer, g.registry)
	proxy.RegisterProxyServiceServer(g.grpcServer, g.proxy)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(g.grpcServer, healthServer)

	healthServer.SetServingStatus("obsidian.orchestrator.v1.OrchestratorService", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("obsidian.registry.v1.RegistryService", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("obsidian.proxy.v1.ProxyService", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(g.grpcServer)
}

func (g *Gateway) setupHTTPServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", g.healthHandler)
	mux.HandleFunc("/ready", g.readyHandler)
	
	if g.config.Monitoring.MetricsEnabled {
		mux.Handle("/metrics", promhttp.Handler())
	}

	g.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", g.config.Server.Port),
		Handler:      mux,
		ReadTimeout:  g.config.Server.ReadTimeout,
		WriteTimeout: g.config.Server.WriteTimeout,
	}
}

func (g *Gateway) Start(ctx context.Context) error {
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", g.config.Server.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	errChan := make(chan error, 2)

	go func() {
		fmt.Printf("Starting gRPC server on port %d\n", g.config.Server.GRPCPort)
		if err := g.grpcServer.Serve(grpcListener); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	go func() {
		fmt.Printf("Starting HTTP server on port %d\n", g.config.Server.Port)
		if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled, shutting down...")
	case sig := <-sigChan:
		fmt.Printf("Received signal %v, shutting down...\n", sig)
	case err := <-errChan:
		return err
	}

	return g.Shutdown()
}

func (g *Gateway) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), g.config.Server.ShutdownTimeout)
	defer cancel()

	errChan := make(chan error, 2)

	go func() {
		g.grpcServer.GracefulStop()
		errChan <- nil
	}()

	go func() {
		errChan <- g.httpServer.Shutdown(ctx)
	}()

	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	g.orchestrator.Shutdown()

	return nil
}

func (g *Gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func (g *Gateway) readyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health, err := g.orchestrator.GetHealth(ctx, nil)
	if err != nil || !health.Healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}