package telemetry

import (
	"context"
)

type Client interface {
	AddMetric(ctx context.Context, metric Metric) error
	Close() error
}
