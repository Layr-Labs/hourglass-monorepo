package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/hooks"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	app := commands.Hgctl()

	actionChain := hooks.NewActionChain()
	actionChain.Use(hooks.WithMetricEmission)

	hooks.ApplyMiddleware(app.Commands, actionChain)

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
