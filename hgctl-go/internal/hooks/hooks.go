package hooks

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/version"

	"github.com/urfave/cli/v2"
)

type ActionChain struct {
	Processors []func(action cli.ActionFunc) cli.ActionFunc
}

func NewActionChain() *ActionChain {
	return &ActionChain{
		Processors: make([]func(action cli.ActionFunc) cli.ActionFunc, 0),
	}
}

func (ac *ActionChain) Use(processor func(action cli.ActionFunc) cli.ActionFunc) {
	ac.Processors = append(ac.Processors, processor)
}

func (ac *ActionChain) Wrap(action cli.ActionFunc) cli.ActionFunc {
	for i := len(ac.Processors) - 1; i >= 0; i-- {
		action = ac.Processors[i](action)
	}
	return action
}

func ApplyMiddleware(commands []*cli.Command, chain *ActionChain) {
	for _, cmd := range commands {
		if cmd.Action != nil {
			cmd.Action = chain.Wrap(cmd.Action)
		}
		if len(cmd.Subcommands) > 0 {
			ApplyMiddleware(cmd.Subcommands, chain)
		}
	}
}

func getFlagValue(ctx *cli.Context, name string) interface{} {
	if !ctx.IsSet(name) {
		return nil
	}

	if ctx.Bool(name) {
		return ctx.Bool(name)
	}
	if ctx.String(name) != "" {
		return ctx.String(name)
	}
	if ctx.Int(name) != 0 {
		return ctx.Int(name)
	}
	if ctx.Float64(name) != 0 {
		return ctx.Float64(name)
	}
	return nil
}

func collectFlagValues(ctx *cli.Context) map[string]interface{} {
	flags := make(map[string]interface{})

	for _, flag := range ctx.App.Flags {
		flagName := flag.Names()[0]
		if ctx.IsSet(flagName) {
			flags[flagName] = getFlagValue(ctx, flagName)
		}
	}

	for _, flag := range ctx.Command.Flags {
		flagName := flag.Names()[0]
		if ctx.IsSet(flagName) {
			flags[flagName] = getFlagValue(ctx, flagName)
		}
	}

	return flags
}

func setupTelemetry() telemetry.Client {
	cfg, err := config.LoadConfig()
	if err != nil {
		return telemetry.NewNoopClient()
	}

	telemetry.Init(cfg)

	client := telemetry.GetGlobalClient()
	if client == nil {
		return telemetry.NewNoopClient()
	}

	return client
}

func WithMetricEmission(action cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		client := setupTelemetry()
		ctx.Context = telemetry.ContextWithClient(ctx.Context, client)

		err := action(ctx)

		emitTelemetryMetrics(ctx, err)

		return err
	}
}

func emitTelemetryMetrics(ctx *cli.Context, actionError error) {
	metrics, err := telemetry.MetricsFromContext(ctx.Context)
	if err != nil {
		return
	}
	metrics.Properties["command"] = ctx.Command.HelpName
	result := "Success"
	dimensions := map[string]string{}
	if actionError != nil {
		result = "Failure"
		dimensions["error"] = actionError.Error()
	}
	metrics.AddMetricWithDimensions(result, 1, dimensions)

	duration := time.Since(metrics.StartTime).Milliseconds()
	metrics.AddMetric("DurationMilliseconds", float64(duration))

	client, ok := telemetry.ClientFromContext(ctx.Context)
	if !ok {
		return
	}
	defer client.Close()

	for _, metric := range metrics.Metrics {
		mDimensions := metric.Dimensions
		for k, v := range metrics.Properties {
			mDimensions[k] = v
		}
		_ = client.AddMetric(ctx.Context, metric)
	}
}

func WithCommandMetricsContext(ctx *cli.Context) error {
	metrics := telemetry.NewMetricsContext()
	ctx.Context = telemetry.WithMetricsContext(ctx.Context, metrics)

	cfg, err := config.LoadConfig()
	if err != nil {
		l := config.LoggerFromContext(ctx.Context)
		l.Error(err.Error())
		return nil
	}

	metrics.Properties["cli_version"] = version.GetVersion()
	metrics.Properties["os"] = runtime.GOOS
	metrics.Properties["arch"] = runtime.GOARCH

	if cfg != nil && cfg.CurrentContext != "" {
		if currentCtx, ok := cfg.Contexts[cfg.CurrentContext]; ok {
			if currentCtx.OperatorAddress != "" {
				isAnonymous := cfg.TelemetryAnonymous != nil && *cfg.TelemetryAnonymous
				if !isAnonymous {
					metrics.Properties["operator_address"] = currentCtx.OperatorAddress
				}
			}
		}
	}

	// Set flags in metrics
	for k, v := range collectFlagValues(ctx) {
		metrics.Properties[k] = fmt.Sprintf("%v", v)
	}

	metrics.AddMetric("Count", 1)
	return nil
}
