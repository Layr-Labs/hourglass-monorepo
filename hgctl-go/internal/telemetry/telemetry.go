package telemetry

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/version"
	"github.com/denisbrodbeck/machineid"
	"github.com/spf13/cobra"
	"github.com/urfave/cli/v2"
)

var (
	embeddedTelemetryApiKey string // Set by build flags
	globalClient            Client
	namespace               = "hgctl"
)

func Init(cfg *config.Config) {
	client, err := NewPostHogClient(cfg, namespace)
	if err != nil || client == nil {
		globalClient = NewNoopClient()
	} else {
		globalClient = client
	}
}

func TrackCommand(cmd *cobra.Command, startTime time.Time) func() {
	if globalClient == nil {
		return func() {}
	}

	metricsCtx := NewMetricsContext()
	metricsCtx.StartTime = startTime

	return func() {
		duration := time.Since(startTime).Milliseconds()

		properties := map[string]interface{}{
			"command":     getCommandPath(cmd),
			"duration_ms": duration,
			"version":     version.GetVersion(),
			"commit":      version.GetCommit(),
			"os":          runtime.GOOS,
			"arch":        runtime.GOARCH,
			"go_version":  runtime.Version(),
			"success":     true,
		}

		if len(metricsCtx.Metrics) > 0 {
			properties["metrics"] = metricsCtx.Metrics
		}

		_ = globalClient.Track(context.Background(), "command_executed", properties)
	}
}

func TrackCLICommand(c *cli.Context, commandName string, startTime time.Time) func() {
	if globalClient == nil {
		return func() {}
	}

	metricsCtx := NewMetricsContext()
	metricsCtx.StartTime = startTime
	ctx := WithMetricsContext(c.Context, metricsCtx)
	c.Context = ctx

	return func() {
		duration := time.Since(startTime).Milliseconds()

		properties := map[string]interface{}{
			"command":     commandName,
			"duration_ms": duration,
			"version":     version.GetVersion(),
			"commit":      version.GetCommit(),
			"os":          runtime.GOOS,
			"arch":        runtime.GOARCH,
			"go_version":  runtime.Version(),
			"success":     true, // Can be overridden if error occurred
		}

		if metrics, err := MetricsFromContext(ctx); err == nil && len(metrics.Metrics) > 0 {
			properties["metrics"] = metrics.Metrics
			for k, v := range metrics.Properties {
				properties[k] = v
			}
		}

		_ = globalClient.Track(context.Background(), "command_executed", properties)
	}
}

func TrackEvent(event string, properties map[string]interface{}) {
	if globalClient == nil {
		return
	}

	if properties == nil {
		properties = make(map[string]interface{})
	}

	properties["version"] = version.GetVersion()
	properties["commit"] = version.GetCommit()
	properties["os"] = runtime.GOOS
	properties["arch"] = runtime.GOARCH

	_ = globalClient.Track(context.Background(), event, properties)
}

func TrackError(err error, context map[string]interface{}) {
	if globalClient == nil || err == nil {
		return
	}

	if context == nil {
		context = make(map[string]interface{})
	}

	context["error"] = err.Error()
	context["error_type"] = fmt.Sprintf("%T", err)

	TrackEvent("error_occurred", context)
}

func EmitMetric(ctx context.Context, name string, value float64, dimensions map[string]string) {
	if metricsCtx, err := MetricsFromContext(ctx); err == nil {
		metricsCtx.AddMetricWithDimensions(name, value, dimensions)
	}

	if globalClient != nil {
		metric := Metric{
			Name:       name,
			Value:      value,
			Dimensions: dimensions,
		}
		_ = globalClient.AddMetric(ctx, metric)
	}
}

func Close() {
	if globalClient != nil {
		_ = globalClient.Close()
	}
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
	id, err := machineid.ID()
	if err != nil {
		hostname, _ := os.Hostname()
		id = fmt.Sprintf("%s-%s-%s", runtime.GOOS, runtime.GOARCH, hostname)
	}

	hash := sha256.Sum256([]byte(id))
	return fmt.Sprintf("%x", hash[:8])
}
