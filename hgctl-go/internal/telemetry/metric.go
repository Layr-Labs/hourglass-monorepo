package telemetry

import (
	"context"
	"errors"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

type MetricsContext struct {
	StartTime  time.Time         `json:"start_time"`
	Metrics    []Metric          `json:"metrics"`
	Properties map[string]string `json:"properties"`
}

type Metric struct {
	Value      float64           `json:"value"`
	Name       string            `json:"name"`
	Dimensions map[string]string `json:"dimensions"`
}

func WithMetricsContext(ctx context.Context, metrics *MetricsContext) context.Context {
	return context.WithValue(ctx, config.MetricsContextKey, metrics)
}

func MetricsFromContext(ctx context.Context) (*MetricsContext, error) {
	metrics, ok := ctx.Value(config.MetricsContextKey).(*MetricsContext)
	if !ok || metrics == nil {
		return nil, errors.New("no metrics context found")
	}
	return metrics, nil
}

func NewMetricsContext() *MetricsContext {
	return &MetricsContext{
		StartTime:  time.Now(),
		Metrics:    make([]Metric, 0),
		Properties: make(map[string]string),
	}
}

func (m *MetricsContext) AddMetric(name string, value float64) {
	m.AddMetricWithDimensions(name, value, make(map[string]string))
}

func (m *MetricsContext) AddMetricWithDimensions(name string, value float64, dimensions map[string]string) {
	m.Metrics = append(m.Metrics, Metric{
		Name:       name,
		Value:      value,
		Dimensions: dimensions,
	})
}

func (m *MetricsContext) AddProperty(key, value string) {
	m.Properties[key] = value
}

func (m *MetricsContext) Duration() time.Duration {
	return time.Since(m.StartTime)
}
