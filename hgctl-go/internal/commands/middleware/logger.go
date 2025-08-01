package middleware

import (
	"context"

	"github.com/urfave/cli/v2"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

// LoggerKey is the context key for storing the logger
const LoggerKey = "logger"

// LoggerBeforeFunc initializes the logger and stores it in the context
func LoggerBeforeFunc(c *cli.Context) error {
	// Initialize logger based on verbose flag
	verbose := c.Bool("verbose")
	logger.InitGlobalLogger(verbose)
	
	// Get the logger instance
	log := logger.GetLogger()
	
	// Store logger in context
	c.Context = context.WithValue(c.Context, LoggerKey, log)
	
	return nil
}

// GetLogger retrieves the logger from the context
func GetLogger(c *cli.Context) logger.Logger {
	if log, ok := c.Context.Value(LoggerKey).(logger.Logger); ok {
		return log
	}
	// Fallback to global logger
	return logger.GetLogger()
}