package telemetry

import (
	"context"
)

type Client interface {
	Track(ctx context.Context, event string, properties map[string]interface{}) error
	AddMetric(ctx context.Context, metric Metric) error
	Close() error
}
