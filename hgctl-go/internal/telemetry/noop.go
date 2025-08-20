package telemetry

import (
	"time"

	"github.com/urfave/cli/v2"
)

// NoopClient implements a no-op telemetry client for when telemetry is disabled
type NoopClient struct{}

// NewNoopClient creates a new no-op client
func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

// TrackCommand does nothing
func (n *NoopClient) TrackCommand(c *cli.Context, duration time.Duration, success bool, err error) {
	// No-op
}

// Close does nothing
func (n *NoopClient) Close() error {
	return nil
}