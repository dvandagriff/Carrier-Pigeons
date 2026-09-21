// Package main is the entry point for the PigeonCoop application.
package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/internal/config"
	"github.com/dvandagriff/Carrier-Pigeons/pigeoncoop/pkg/app"
)

func main() {
	// Parse configuration flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	config, err := config.LoadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Create application
	app := app.New(config)

	// Start application
	if err := app.Start(); err != nil {
		slog.Error("application failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Handle shutdown
	app.Shutdown()
}
