package telemetry

import "context"

type NoopClient struct{}

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (c *NoopClient) AddMetric(_ context.Context, _ Metric) error {
	return nil
}

func (c *NoopClient) Close() error {
	return nil
}

func IsNoopClient(client Client) bool {
	_, isNoop := client.(*NoopClient)
	return isNoop
}
