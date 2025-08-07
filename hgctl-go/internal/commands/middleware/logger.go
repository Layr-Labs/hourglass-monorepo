package middleware

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/urfave/cli/v2"
)

// LoggerBeforeFunc initializes the logger and stores it in the context
func LoggerBeforeFunc(c *cli.Context) (error, logger.Logger) {
	// Initialize logger based on verbose flag
	verbose := c.Bool("verbose")
	logger.InitGlobalLogger(verbose)

	// Get the logger instance
	log := logger.GetLogger()

	// Store logger in context
	c.Context = context.WithValue(c.Context, config.LoggerKey, log)

	return nil, log
}

// GetLogger retrieves the logger from the context
func GetLogger(c *cli.Context) logger.Logger {
	// Check if we need to reinitialize the logger with the current writer
	// This is important for tests where app.Writer might be different
	verbose := c.Bool("verbose")

	// Always create a logger with the current context's writer
	// This ensures test output capture works correctly
	return logger.NewLoggerWithWriter(verbose, c.App.Writer)
}
