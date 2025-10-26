package main

import (
	"fmt"
	"os"

	"github.com/ranjbar-dev/loki-notification/internal/config"
	"github.com/ranjbar-dev/loki-notification/internal/httpserver"
	"github.com/ranjbar-dev/loki-notification/internal/logger"
	"github.com/ranjbar-dev/loki-notification/srv"
)

var (
	configPath string = "config/config.yaml"
)

func main() {

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {

		fmt.Fprintf(os.Stderr, "failed to load configuration: %v", err)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.App.LogLevel, cfg.App.Environment)
	if err != nil {

		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v", err)
	}
	defer log.Sync()

	hs := httpserver.NewHttpServer(cfg.Api.Host, cfg.Api.Port, cfg.Api.CertLocation, cfg.Api.KeyLocation)

	srv := srv.NewService(cfg, log, hs)

	log.Info("Loki Notification Server Started")
	err = srv.Start()
	if err != nil {

		fmt.Fprintf(os.Stderr, "failed to start service: %v", err)
	}

	log.Info("Loki Notification Server Stopped")

}
