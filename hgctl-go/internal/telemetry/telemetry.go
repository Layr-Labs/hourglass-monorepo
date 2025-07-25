package telemetry

import (
    "crypto/sha256"
    "fmt"
    "os"
    "runtime"
    "time"
    
    "github.com/denisbrodbeck/machineid"
    "github.com/spf13/cobra"
)

var (
    embeddedTelemetryApiKey string // Set by build flags
    client                  *Client
)

type Client struct {
    enabled  bool
    distinct string
}

func Init() {
    // For now, telemetry is disabled
    // TODO: Implement proper telemetry
    client = &Client{enabled: false}
}

func TrackCommand(cmd *cobra.Command, startTime time.Time) func() {
    // Return a no-op cleanup function
    return func() {}
}

func TrackEvent(event string, properties map[string]interface{}) {
    // No-op for now
}

func TrackError(err error, context map[string]interface{}) {
    // No-op for now
}

func Close() {
    // No-op for now
}

func getCommandPath(cmd *cobra.Command) string {
    if cmd == nil {
        return "root"
    }
    
    path := cmd.CommandPath()
    if path == "" {
        path = "root"
    }
    return path
}

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

func getVersion() string {
    return "dev"
}
