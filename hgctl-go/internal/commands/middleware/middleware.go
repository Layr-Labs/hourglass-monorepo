package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// ChainBeforeFuncs chains multiple BeforeFuncs together
func ChainBeforeFuncs(funcs ...cli.BeforeFunc) cli.BeforeFunc {
	return func(c *cli.Context) error {
		for _, fn := range funcs {
			if err := fn(c); err != nil {
				return err
			}
		}
		return nil
	}
}

// ChainAfterFuncs chains multiple AfterFuncs together
func ChainAfterFuncs(funcs ...cli.AfterFunc) cli.AfterFunc {
	return func(c *cli.Context) error {
		for _, fn := range funcs {
			if err := fn(c); err != nil {
				return err
			}
		}
		return nil
	}
}

// StandardMiddlewareChain returns the standard middleware chain with proper ordering:
func StandardMiddlewareChain() cli.BeforeFunc {
	return ChainBeforeFuncs(
		// Initialize logger first (always needed)
		func(c *cli.Context) error {
			err, _ := LoggerBeforeFunc(c)
			return err
		},
		// Load context configuration
		func(c *cli.Context) error {
			ctxConfig, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if ctxConfig != nil && ctxConfig.CurrentContext != "" {
				ctx, exists := ctxConfig.Contexts[ctxConfig.CurrentContext]
				if !exists {
					return fmt.Errorf("current context '%s' not found", ctxConfig.CurrentContext)
				}
				// Store the context in the CLI context
				c.Context = context.WithValue(c.Context, config.ContextKey, ctx)
			}
			return nil
		},
		// Load secrets from EnvSecretsPath (must be before contract client)
		SecretsBeforeFunc,
		// Initialize contract client (may need secrets)
		ContractBeforeFunc,
	)
}

// MiddlewareBeforeFunc combines logger, secrets loading, and contract client initialization
func MiddlewareBeforeFunc(c *cli.Context) error {
	// Initialize logger first
	var l logger.Logger
	var err error
	if err, l = LoggerBeforeFunc(c); err != nil {
		return err
	}

	ctxConfig, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if ctxConfig != nil {
		// Load secrets from EnvSecretsPath BEFORE initializing contract client
		// This ensures environment variables are available for GetOperatorPrivateKey
		if err := SecretsBeforeFunc(c); err != nil {
			l.Warn("Failed to load secrets", zap.Error(err))
			// Don't fail here, let it continue
		}

		// Initialize contract client (which may need secrets)
		if err := ContractBeforeFunc(c); err != nil {
			l.Debug("Failed to initialize contract client", zap.Error(err))
		}
	}

	return nil
}

// TelemetryBeforeFunc starts telemetry tracking for the command
func TelemetryBeforeFunc(c *cli.Context) error {
	// Skip telemetry for help commands
	if c.Command == nil || c.Command.Name == "help" {
		return nil
	}

	// Get the full command path
	commandPath := c.Command.FullName()
	if commandPath == "" && c.Command != nil {
		commandPath = c.Command.Name
	}

	// Start tracking
	startTime := time.Now()
	finishFunc := telemetry.TrackCLICommand(c, commandPath, startTime)

	// Store the finish function in context to be called after command execution
	c.Context = context.WithValue(c.Context, config.TelemetryContextKey, finishFunc)

	return nil
}

// TelemetryAfterFunc completes telemetry tracking
func TelemetryAfterFunc(c *cli.Context) error {
	// Retrieve and call the finish function if it exists
	if finishFunc, ok := c.Context.Value(config.TelemetryContextKey).(func()); ok {
		finishFunc()
	}
	return nil
}

func ExitErrHandler(c *cli.Context, err error) {
	if err == nil {
		return
	}

	// Try to get logger from context, or create a new one
	var log logger.Logger
	if c != nil {
		log = GetLogger(c)
	} else {
		logger.InitGlobalLogger(false)
		log = logger.GetLogger()
	}

	// Log the error with appropriate context
	if c != nil && c.Command != nil {
		log.Error("Command execution failed",
			zap.String("command", c.Command.Name),
			zap.Error(err))
	} else {
		log.Error("Command execution failed", zap.Error(err))
	}

	// Track error in telemetry
	if c != nil && c.Command != nil {
		telemetry.TrackError(err, map[string]interface{}{
			"command": c.Command.Name,
		})
	}
}
