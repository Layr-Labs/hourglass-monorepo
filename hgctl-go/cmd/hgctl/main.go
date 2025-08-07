package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry"
)

func main() {
	// Initialize telemetry
	telemetry.Init()
	defer telemetry.Close()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	app := commands.Hgctl()

	if err := app.RunContext(ctx, os.Args); err != nil {
		// Error already logged by middleware
		os.Exit(1)
	}
}
