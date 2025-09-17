package telemetry

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/denisbrodbeck/machineid"
)

var (
	embeddedTelemetryApiKey string // Set by build flags
	globalClient            Client
	namespace               = "Hgctl"
)

func Init(cfg *config.Config) {
	client, err := NewPostHogClient(cfg, namespace)
	if err != nil || client == nil {
		globalClient = NewNoopClient()
	} else {
		globalClient = client
	}
}

func GetGlobalClient() Client {
	return globalClient
}

func ContextWithClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, config.TelemetryContextKey, client)
}

func ClientFromContext(ctx context.Context) (Client, bool) {
	client, ok := ctx.Value(config.TelemetryContextKey).(Client)
	return client, ok
}

func Close() {
	if globalClient != nil {
		_ = globalClient.Close()
	}
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
