package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	pb "github.com/hourglass/obsidian/api/proto/proxy"
	"github.com/hourglass/obsidian/pkg/config"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedProxyServiceServer
	
	config      *config.ProxyConfig
	backends    map[string]*Backend
	httpClient  *http.Client
	metrics     *ProxyMetrics
	mu          sync.RWMutex
}

type Backend struct {
	Name          string
	URL           string
	Enabled       bool
	Status        pb.BackendStatus
	RateLimiter   *rate.Limiter
	AllowedMethods map[string]bool
	Filters       []Filter
	CircuitBreaker *CircuitBreaker
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Filter interface {
	Apply(req *pb.Request) error
}

type CircuitBreaker struct {
	mu              sync.Mutex
	failures        int
	successes       int
	lastFailureTime time.Time
	state           string
	failureThreshold int
	successThreshold int
	timeout         time.Duration
}

type ProxyMetrics struct {
	mu                   sync.RWMutex
	totalRequests        int64
	successfulRequests   int64
	failedRequests       int64
	rateLimitedRequests  int64
	requestsByMethod     map[string]int64
	requestsByStatus     map[int32]int64
	latencies            []float64
}

func NewServer(config *config.ProxyConfig) (*Server, error) {
	s := &Server{
		config:   config,
		backends: make(map[string]*Backend),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		metrics: &ProxyMetrics{
			requestsByMethod: make(map[string]int64),
			requestsByStatus: make(map[int32]int64),
		},
	}

	if err := s.initializeBackends(); err != nil {
		return nil, fmt.Errorf("failed to initialize backends: %w", err)
	}

	go s.metricsCleanupLoop()

	return s, nil
}

func (s *Server) ProxyRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	start := time.Now()

	backend, err := s.getBackend(req.BackendName)
	if err != nil {
		return nil, err
	}

	if !backend.Enabled {
		return nil, status.Errorf(codes.FailedPrecondition, "backend %s is disabled", req.BackendName)
	}

	if backend.Status == pb.BackendStatus_BACKEND_STATUS_CIRCUIT_OPEN {
		if !backend.CircuitBreaker.shouldTryRequest() {
			s.metrics.recordRequest(req.Method, 503, time.Since(start))
			return nil, status.Errorf(codes.Unavailable, "circuit breaker is open for backend %s", req.BackendName)
		}
	}

	if !backend.RateLimiter.Allow() {
		s.metrics.recordRateLimited()
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded for backend %s", req.BackendName)
	}

	if !backend.AllowedMethods[req.Method] {
		return nil, status.Errorf(codes.PermissionDenied, "method %s not allowed for backend %s", req.Method, req.BackendName)
	}

	for _, filter := range backend.Filters {
		if err := filter.Apply(req); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "request filtered: %v", err)
		}
	}

	httpReq, err := s.buildHTTPRequest(backend.URL, req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to build request: %v", err)
	}

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		backend.CircuitBreaker.recordFailure()
		s.metrics.recordRequest(req.Method, 0, time.Since(start))
		return nil, status.Errorf(codes.Internal, "request failed: %v", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read response body: %v", err)
	}

	if httpResp.StatusCode >= 500 {
		backend.CircuitBreaker.recordFailure()
	} else {
		backend.CircuitBreaker.recordSuccess()
	}

	latency := time.Since(start)
	s.metrics.recordRequest(req.Method, int32(httpResp.StatusCode), latency)

	headers := make(map[string]string)
	for k, v := range httpResp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &pb.Response{
		StatusCode: int32(httpResp.StatusCode),
		Headers:    headers,
		Body:       body,
		Latency:    durationpb.New(latency),
	}, nil
}

func (s *Server) AddBackend(ctx context.Context, req *pb.BackendConfig) (*pb.Backend, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.backends[req.Name]; exists {
		return nil, status.Errorf(codes.AlreadyExists, "backend %s already exists", req.Name)
	}

	backend := s.createBackend(req)
	s.backends[req.Name] = backend

	return s.backendToProto(backend), nil
}

func (s *Server) UpdateBackend(ctx context.Context, req *pb.BackendConfig) (*pb.Backend, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	backend, ok := s.backends[req.Name]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "backend not found: %s", req.Name)
	}

	backend.URL = req.Url
	backend.RateLimiter = rate.NewLimiter(rate.Limit(req.RateLimits.RequestsPerSecond), int(req.RateLimits.Burst))
	backend.AllowedMethods = s.buildAllowedMethods(req.AllowedMethods)
	backend.Filters = s.buildFilters(req.Filters)
	backend.UpdatedAt = time.Now()

	return s.backendToProto(backend), nil
}

func (s *Server) RemoveBackend(ctx context.Context, req *pb.BackendID) (*pb.Backend, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	backend, ok := s.backends[req.Name]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "backend not found: %s", req.Name)
	}

	delete(s.backends, req.Name)

	return s.backendToProto(backend), nil
}

func (s *Server) ListBackends(ctx context.Context, req *pb.ListBackendsRequest) (*pb.BackendList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	backends := make([]*pb.Backend, 0, len(s.backends))
	for _, backend := range s.backends {
		backends = append(backends, s.backendToProto(backend))
	}

	pageSize := 50
	if req.PageSize > 0 {
		pageSize = int(req.PageSize)
	}

	start := 0
	if req.PageToken != "" {
		start = s.parsePageToken(req.PageToken)
	}

	end := start + pageSize
	if end > len(backends) {
		end = len(backends)
	}

	var nextPageToken string
	if end < len(backends) {
		nextPageToken = s.generatePageToken(end)
	}

	return &pb.BackendList{
		Backends:      backends[start:end],
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Server) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.ProxyMetrics, error) {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	avgLatency, p95Latency, p99Latency := s.metrics.calculateLatencyPercentiles()

	requestsByMethod := make(map[string]int64)
	for k, v := range s.metrics.requestsByMethod {
		requestsByMethod[k] = v
	}

	requestsByStatus := make(map[int32]int64)
	for k, v := range s.metrics.requestsByStatus {
		requestsByStatus[k] = v
	}

	return &pb.ProxyMetrics{
		TotalRequests:       s.metrics.totalRequests,
		SuccessfulRequests:  s.metrics.successfulRequests,
		FailedRequests:      s.metrics.failedRequests,
		RateLimitedRequests: s.metrics.rateLimitedRequests,
		AverageLatencyMs:    avgLatency,
		P95LatencyMs:        p95Latency,
		P99LatencyMs:        p99Latency,
		RequestsByMethod:    requestsByMethod,
		RequestsByStatus:    requestsByStatus,
	}, nil
}

func (s *Server) initializeBackends() error {
	for _, cfg := range s.config.Backends {
		backend := s.createBackend(&pb.BackendConfig{
			Name:           cfg.Name,
			Url:            cfg.URL,
			RateLimits:     &pb.RateLimits{
				RequestsPerSecond: int32(cfg.RateLimits.RequestsPerSecond),
				Burst:             int32(cfg.RateLimits.Burst),
			},
			AllowedMethods: cfg.AllowedMethods,
			Filters:        s.convertFilters(cfg.Filters),
		})
		s.backends[cfg.Name] = backend
	}
	return nil
}

func (s *Server) createBackend(config *pb.BackendConfig) *Backend {
	backend := &Backend{
		Name:           config.Name,
		URL:            config.Url,
		Enabled:        true,
		Status:         pb.BackendStatus_BACKEND_STATUS_HEALTHY,
		RateLimiter:    rate.NewLimiter(rate.Limit(config.RateLimits.RequestsPerSecond), int(config.RateLimits.Burst)),
		AllowedMethods: s.buildAllowedMethods(config.AllowedMethods),
		Filters:        s.buildFilters(config.Filters),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if config.CircuitBreaker != nil {
		backend.CircuitBreaker = &CircuitBreaker{
			failureThreshold: int(config.CircuitBreaker.FailureThreshold),
			successThreshold: int(config.CircuitBreaker.SuccessThreshold),
			timeout:          config.CircuitBreaker.Timeout.AsDuration(),
			state:            "closed",
		}
	} else {
		backend.CircuitBreaker = &CircuitBreaker{
			failureThreshold: 5,
			successThreshold: 2,
			timeout:          30 * time.Second,
			state:            "closed",
		}
	}

	return backend
}

func (s *Server) getBackend(name string) (*Backend, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	backend, ok := s.backends[name]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "backend not found: %s", name)
	}

	return backend, nil
}

func (s *Server) buildHTTPRequest(baseURL string, req *pb.Request) (*http.Request, error) {
	url := baseURL + req.Path

	httpReq, err := http.NewRequest(req.Method, url, bytes.NewReader(req.Body))
	if err != nil {
		return nil, err
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	httpReq.Header.Set("X-Container-ID", req.ContainerId)

	return httpReq, nil
}

func (s *Server) buildAllowedMethods(methods []string) map[string]bool {
	allowed := make(map[string]bool)
	for _, method := range methods {
		allowed[method] = true
	}
	return allowed
}

func (s *Server) buildFilters(filters []*pb.Filter) []Filter {
	var result []Filter
	for _, f := range filters {
		switch f.Type {
		case pb.FilterType_FILTER_TYPE_CONTENT_SIZE:
			result = append(result, &ContentSizeFilter{maxSize: s.parseSize(f.Value)})
		case pb.FilterType_FILTER_TYPE_HEADER:
			result = append(result, &HeaderFilter{header: f.Parameter, value: f.Value})
		}
	}
	return result
}

func (s *Server) convertFilters(filters []config.FilterConfig) []*pb.Filter {
	var result []*pb.Filter
	for _, f := range filters {
		filterType := pb.FilterType_FILTER_TYPE_UNSPECIFIED
		switch f.Type {
		case "contentSize":
			filterType = pb.FilterType_FILTER_TYPE_CONTENT_SIZE
		case "header":
			filterType = pb.FilterType_FILTER_TYPE_HEADER
		}
		result = append(result, &pb.Filter{
			Type:      filterType,
			Parameter: f.Parameter,
			Value:     f.Value,
		})
	}
	return result
}

func (s *Server) backendToProto(b *Backend) *pb.Backend {
	return &pb.Backend{
		Name:      b.Name,
		Url:       b.URL,
		Enabled:   b.Enabled,
		Status:    b.Status,
		CreatedAt: timestamppb.New(b.CreatedAt),
		UpdatedAt: timestamppb.New(b.UpdatedAt),
	}
}

func (s *Server) parsePageToken(token string) int {
	return 0
}

func (s *Server) generatePageToken(offset int) string {
	return fmt.Sprintf("%d", offset)
}

func (s *Server) parseSize(size string) int64 {
	return 10485760
}

func (s *Server) metricsCleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.metrics.cleanup()
	}
}

type ContentSizeFilter struct {
	maxSize int64
}

func (f *ContentSizeFilter) Apply(req *pb.Request) error {
	if int64(len(req.Body)) > f.maxSize {
		return fmt.Errorf("content size %d exceeds limit %d", len(req.Body), f.maxSize)
	}
	return nil
}

type HeaderFilter struct {
	header string
	value  string
}

func (f *HeaderFilter) Apply(req *pb.Request) error {
	if req.Headers[f.header] != f.value {
		return fmt.Errorf("header %s does not match required value", f.header)
	}
	return nil
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.failureThreshold {
		cb.state = "open"
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successes++

	if cb.state == "half-open" && cb.successes >= cb.successThreshold {
		cb.state = "closed"
		cb.failures = 0
		cb.successes = 0
	}
}

func (cb *CircuitBreaker) shouldTryRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "closed" {
		return true
	}

	if cb.state == "open" && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = "half-open"
		cb.successes = 0
		return true
	}

	return cb.state == "half-open"
}

func (m *ProxyMetrics) recordRequest(method string, statusCode int32, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++
	m.requestsByMethod[method]++
	m.requestsByStatus[statusCode]++

	if statusCode >= 200 && statusCode < 300 {
		m.successfulRequests++
	} else {
		m.failedRequests++
	}

	m.latencies = append(m.latencies, latency.Seconds()*1000)
}

func (m *ProxyMetrics) recordRateLimited() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rateLimitedRequests++
}

func (m *ProxyMetrics) calculateLatencyPercentiles() (avg, p95, p99 float64) {
	if len(m.latencies) == 0 {
		return 0, 0, 0
	}

	sum := 0.0
	for _, l := range m.latencies {
		sum += l
	}
	avg = sum / float64(len(m.latencies))

	p95Index := int(float64(len(m.latencies)) * 0.95)
	p99Index := int(float64(len(m.latencies)) * 0.99)

	if p95Index < len(m.latencies) {
		p95 = m.latencies[p95Index]
	}
	if p99Index < len(m.latencies) {
		p99 = m.latencies[p99Index]
	}

	return
}

func (m *ProxyMetrics) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.latencies) > 10000 {
		m.latencies = m.latencies[len(m.latencies)-10000:]
	}
}