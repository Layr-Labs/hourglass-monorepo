package telemetry

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/posthog/posthog-go"
	"github.com/urfave/cli/v2"
)

const (
	namespace = "hgctl"
	endpoint  = "https://us.i.posthog.com"
)

// PostHogClient wraps the PostHog client for telemetry
type PostHogClient struct {
	client    posthog.Client
	enabled   bool
	sessionID string
}

// NewPostHogClient creates a new PostHog client
func NewPostHogClient() (*PostHogClient, error) {
	// Check if telemetry is enabled
	enabled := isEnabled()
	if !enabled {
		return &PostHogClient{enabled: false}, nil
	}

	// Get API key from environment
	apiKey := os.Getenv("HGCTL_POSTHOG_KEY")
	if apiKey == "" {
		// No API key, return disabled client
		return &PostHogClient{enabled: false}, nil
	}

	// Create PostHog client
	client, err := posthog.NewWithConfig(apiKey, posthog.Config{
		Endpoint: endpoint,
	})
	if err != nil {
		// Fail silently, return disabled client
		return &PostHogClient{enabled: false}, nil
	}

	return &PostHogClient{
		client:    client,
		enabled:   true,
		sessionID: getAnonymousID(),
	}, nil
}

// TrackCommand tracks a command execution
func (p *PostHogClient) TrackCommand(c *cli.Context, duration time.Duration, success bool, err error) {
	if !p.enabled || p.client == nil {
		return
	}

	// Build properties
	properties := posthog.NewProperties()
	properties.Set("command", getFullCommandPath(c))
	properties.Set("success", success)
	properties.Set("duration_ms", duration.Milliseconds())
	properties.Set("version", getVersion())
	properties.Set("os", runtime.GOOS)
	properties.Set("arch", runtime.GOARCH)

	// Add sanitized error if failed
	if !success && err != nil {
		properties.Set("error", sanitizeError(err))
	}

	// Send event (fire and forget)
	_ = p.client.Enqueue(posthog.Capture{
		DistinctId: p.sessionID,
		Event:      namespace,
		Properties: properties,
	})
}

// Close closes the PostHog client
func (p *PostHogClient) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// getFullCommandPath builds the full command path from context
func getFullCommandPath(c *cli.Context) string {
	if c == nil {
		return "unknown"
	}

	// Build full command path
	var parts []string
	
	// Add parent command if exists
	if c.App != nil && c.App.Name != "" {
		parts = append(parts, c.App.Name)
	}

	// Add command name
	if c.Command != nil && c.Command.Name != "" {
		parts = append(parts, c.Command.Name)
	}

	// Add subcommand if present
	for _, arg := range c.Args().Slice() {
		// Only add if it looks like a subcommand (not a flag or value)
		if !strings.HasPrefix(arg, "-") {
			parts = append(parts, arg)
			break // Only take first non-flag arg
		}
	}

	if len(parts) == 0 {
		return "root"
	}

	return strings.Join(parts, " ")
}

// sanitizeError removes sensitive information from error messages
func sanitizeError(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()
	
	// Remove file paths
	msg = strings.ReplaceAll(msg, os.Getenv("HOME"), "~")
	
	// Truncate if too long
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}

	return msg
}

// isEnabled checks if telemetry is enabled via config or environment
func isEnabled() bool {
	// Import cycle prevention - check env var directly here
	// The config package will also check this env var
	if env := os.Getenv("HGCTL_TELEMETRY_ENABLED"); env != "" {
		return env == "true" || env == "1"
	}
	
	// Default to disabled if env var not set
	// Config file check happens in config package
	return false
}

// getVersion returns the hgctl version
func getVersion() string {
	// TODO: Get from version package
	return "dev"
}