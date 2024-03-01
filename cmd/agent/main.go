package main

import (
	"github.com/aksenk/go-yandex-metrics/internal/agent/app"
	"github.com/aksenk/go-yandex-metrics/internal/agent/config"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"net/http"
	"time"
)

func main() {
	log, err := logger.NewLogger("info")
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}
	cfgLog, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}
	client := http.Client{
		Timeout: time.Duration(cfg.ClientTimeout) * time.Second,
	}

	agent, err := app.NewApp(&client, cfgLog, cfg)
	if err != nil {
		log.Fatalf("Error creating agent: %v", err)
	}

	agent.Run()
}
