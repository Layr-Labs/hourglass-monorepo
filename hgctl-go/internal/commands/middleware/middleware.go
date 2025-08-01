package middleware

import (
	"github.com/urfave/cli/v2"
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
	if err := LoggerBeforeFunc(c); err != nil {
		return err
	}
	
	// Initialize contract client
	if err := ContractBeforeFunc(c); err != nil {
		return err
	}
	
	return nil
}