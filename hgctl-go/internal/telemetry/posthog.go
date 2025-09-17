package telemetry

import (
	"context"
	"fmt"
	"os"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/posthog/posthog-go"
)

type PostHogClient struct {
	namespace       string
	client          posthog.Client
	distinctID      string
	operatorAddress string
	enabled         bool
}

func NewPostHogClient(cfg *config.Config, namespace string) (*PostHogClient, error) {
	if !isTelemetryEnabled(cfg) {
		return nil, nil
	}

	apiKey := getPostHogAPIKey(cfg)
	if apiKey == "" {
		return nil, nil
	}

	client, err := posthog.NewWithConfig(apiKey, posthog.Config{
		Endpoint: getPostHogEndpoint(),
	})
	if err != nil {
		return nil, err
	}

	distinctID := getAnonymousID()

	var operatorAddress string
	if cfg != nil && !cfg.TelemetryAnonymous && cfg.CurrentContext != "" {
		if ctx, ok := cfg.Contexts[cfg.CurrentContext]; ok && ctx.OperatorAddress != "" {
			operatorAddress = ctx.OperatorAddress
		}
	}

	return &PostHogClient{
		namespace:       namespace,
		client:          client,
		distinctID:      distinctID,
		operatorAddress: operatorAddress,
		enabled:         true,
	}, nil
}

func (c *PostHogClient) Track(_ context.Context, event string, properties map[string]interface{}) error {
	if c == nil || c.client == nil || !c.enabled {
		return nil
	}

	eventName := fmt.Sprintf("%s_%s", c.namespace, event)

	if c.operatorAddress != "" {
		properties["operator_address"] = c.operatorAddress
	}

	_ = c.client.Enqueue(posthog.Capture{
		DistinctId: c.distinctID,
		Event:      eventName,
		Properties: properties,
	})
	return nil
}

func (c *PostHogClient) AddMetric(_ context.Context, metric Metric) error {
	if c == nil || c.client == nil || !c.enabled {
		return nil
	}

	props := make(map[string]interface{})
	props["metric_name"] = metric.Name
	props["metric_value"] = metric.Value

	if c.operatorAddress != "" {
		props["operator_address"] = c.operatorAddress
	}

	for k, v := range metric.Dimensions {
		props[k] = v
	}

	_ = c.client.Enqueue(posthog.Capture{
		DistinctId: c.distinctID,
		Event:      fmt.Sprintf("%s_metric", c.namespace),
		Properties: props,
	})
	return nil
}

func (c *PostHogClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	_ = c.client.Close()
	return nil
}

func isTelemetryEnabled(cfg *config.Config) bool {
	if envVal := os.Getenv("HGCTL_TELEMETRY_ENABLED"); envVal != "" {
		return envVal == "true" || envVal == "1"
	}

	if cfg != nil && cfg.TelemetryEnabled != nil {
		return *cfg.TelemetryEnabled
	}

	return false
}

func getPostHogAPIKey(cfg *config.Config) string {
	if key := os.Getenv("HGCTL_POSTHOG_KEY"); key != "" {
		return key
	}

	if cfg != nil && cfg.PostHogAPIKey != "" {
		return cfg.PostHogAPIKey
	}

	return embeddedTelemetryApiKey
}

func getPostHogEndpoint() string {
	if endpoint := os.Getenv("HGCTL_POSTHOG_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return "https://us.i.posthog.com"
}
