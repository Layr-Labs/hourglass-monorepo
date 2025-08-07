package middleware

import (
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
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

// MiddlewareBeforeFunc combines logger and contract client initialization
func MiddlewareBeforeFunc(c *cli.Context) error {
	// Initialize logger first
	var l logger.Logger
	var err error
	if err, l = LoggerBeforeFunc(c); err != nil {
		return err
	}

	// Initialize contract client
	if err := ContractBeforeFunc(c); err != nil {
		l.Error("Failed to initialize contract client", zap.Error(err))
		return err
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
}
