package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SathyaBhat/tuya-exporter/internal/config"
	"github.com/SathyaBhat/tuya-exporter/internal/exporter"
)

func main() {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("config loaded", "tuya_devices", len(cfg.Tuya.DeviceIDs))

	// Create exporter
	exp, err := exporter.New(cfg)
	if err != nil {
		slog.Error("failed to create exporter", "error", err)
		os.Exit(1)
	}
	defer exp.Close()

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("shutdown signal received", "signal", sig)
		cancel()
	}()

	// Run exporter
	if err := exp.Run(ctx); err != nil && err != context.Canceled {
		slog.Error("exporter error", "error", err)
		os.Exit(1)
	}

	slog.Info("exporter stopped")
}
