package middleware

import (
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry"
	"github.com/urfave/cli/v2"
)

// WithTelemetry wraps a command action to track telemetry
func WithTelemetry(action cli.ActionFunc) cli.ActionFunc {
	if action == nil {
		return nil
	}

	return func(c *cli.Context) error {
		// Record start time
		start := time.Now()

		// Execute the actual command
		err := action(c)

		// Track telemetry (fire and forget)
		go func() {
			duration := time.Since(start)
			success := err == nil
			telemetry.TrackCommand(c, duration, success, err)
		}()

		// Always return the original error
		return err
	}
}

// InitTelemetry initializes the telemetry client at startup
func InitTelemetry(c *cli.Context) error {
	telemetry.Init()
	return nil
}