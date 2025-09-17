package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	telemetry.Init(cfg)
	defer telemetry.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
