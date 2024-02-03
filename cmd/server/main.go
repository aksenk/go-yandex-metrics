package main

import (
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/filestorage"
	"net/http"
)

func main() {
	log := logger.Log
	config, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Can not create app config: %v", err)
	}
	s, err := filestorage.NewFileStorage(config.MetricsFileName)
	if err != nil {
		log.Fatalf("Can not init storage: %v", err)
	}
	r := handlers.NewRouter(s)
	log.Infof("Starting web server on %v", config.ServerListenAddr)
	if err := http.ListenAndServe(config.ServerListenAddr, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
