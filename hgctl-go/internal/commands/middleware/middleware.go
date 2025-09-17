package middleware

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

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

func StandardMiddlewareChain() cli.BeforeFunc {
	return ChainBeforeFuncs(
		func(c *cli.Context) error {
			err, _ := LoggerBeforeFunc(c)
			return err
		},
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
				c.Context = context.WithValue(c.Context, config.ContextKey, ctx)
			}
			return nil
		},
		SecretsBeforeFunc,
		ContractBeforeFunc,
	)
}

func MiddlewareBeforeFunc(c *cli.Context) error {
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
		if err := SecretsBeforeFunc(c); err != nil {
			l.Warn("Failed to load secrets", zap.Error(err))
			// Don't fail here, let it continue
		}

		if err := ContractBeforeFunc(c); err != nil {
			l.Debug("Failed to initialize contract client", zap.Error(err))
		}
	}

	return nil
}

func ExitErrHandler(c *cli.Context, err error) {
	if err == nil {
		return
	}

	var log logger.Logger
	if c != nil {
		log = GetLogger(c)
	} else {
		logger.InitGlobalLogger(false)
		log = logger.GetLogger()
	}

	if c != nil && c.Command != nil {
		log.Error("Command execution failed",
			zap.String("command", c.Command.Name),
			zap.Error(err))
	} else {
		log.Error("Command execution failed", zap.Error(err))
	}

}
