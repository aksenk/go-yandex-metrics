package main

import (
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/app"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
)

func main() {
	log := logger.Log

	config, err := config.GetConfig()
	if err != nil {
		log.Fatalf("can not create app config: %v", err)
	}
	app, err := app.NewApp(config)
	if err != nil {
		log.Fatalf("Application initialization error: %v", err)
	}
	err = app.Start()
	if err != nil {
		log.Fatalf("Application launch error: %v", err)
	}
}
