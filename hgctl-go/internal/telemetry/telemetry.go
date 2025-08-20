package telemetry

import (
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/urfave/cli/v2"
)

// TelemetryClient interface for telemetry operations
type TelemetryClient interface {
	TrackCommand(c *cli.Context, duration time.Duration, success bool, err error)
	Close() error
}

var (
	embeddedTelemetryApiKey string // Set by build flags
	client                  TelemetryClient
)

// Init initializes the telemetry client
func Init() {
	// Try to create PostHog client
	phClient, err := NewPostHogClient()
	if err != nil || phClient == nil || !phClient.enabled {
		// Fall back to noop client
		client = NewNoopClient()
		return
	}

	client = phClient
}

// TrackCommand tracks a command execution with timing and success/failure
func TrackCommand(c *cli.Context, duration time.Duration, success bool, err error) {
	if client == nil {
		return
	}
	
	client.TrackCommand(c, duration, success, err)
}

//nolint:unused
func TrackEvent(event string, properties map[string]interface{}) {
	// No-op for now
}

//nolint:unused
func TrackError(err error, context map[string]interface{}) {
	// No-op for now
}

// Close closes the telemetry client
func Close() {
	if client != nil {
		_ = client.Close()
	}
}

// getCommandPath returns the full command path from cli.Context
func getCommandPath(c *cli.Context) string {
	if c == nil || c.Command == nil {
		return "root"
	}

	// Build command path from context
	path := c.Command.Name
	if c.Command.Category != "" {
		path = c.Command.Category + " " + path
	}
	
	return path
}

//nolint:unused
func getAnonymousID() string {
	// Try to get machine ID
	id, err := machineid.ID()
	if err != nil {
		// Fallback to hostname-based ID
		hostname, _ := os.Hostname()
		id = fmt.Sprintf("%s-%s-%s", runtime.GOOS, runtime.GOARCH, hostname)
	}

	// Hash it for privacy
	hash := sha256.Sum256([]byte(id))
	return fmt.Sprintf("%x", hash[:8])
}

