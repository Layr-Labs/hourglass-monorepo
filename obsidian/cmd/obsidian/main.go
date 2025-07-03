package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hourglass/obsidian/pkg/config"
	"github.com/hourglass/obsidian/pkg/gateway"
)

var (
	configPath = flag.String("config", "config/default.yaml", "Path to configuration file")
	version    = flag.Bool("version", false, "Print version information")
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("Obsidian Gateway\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	setupLogging(cfg.Logging)

	log.Printf("Starting Obsidian Gateway v%s", Version)
	log.Printf("Configuration loaded from: %s", *configPath)

	gw, err := gateway.NewGateway(cfg)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}

	ctx := context.Background()
	if err := gw.Start(ctx); err != nil {
		log.Fatalf("Gateway error: %v", err)
	}

	log.Println("Obsidian Gateway shutdown complete")
}

func setupLogging(cfg config.LoggingConfig) {
	switch cfg.Format {
	case "json":
		log.SetFlags(0)
	default:
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	if cfg.OutputPath != "" && cfg.OutputPath != "stdout" {
		file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file %s: %v", cfg.OutputPath, err)
		} else {
			log.SetOutput(file)
		}
	}
}