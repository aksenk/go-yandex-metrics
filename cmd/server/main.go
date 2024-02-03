package main

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/filestorage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type App struct {
	storage *filestorage.FileStorage
	router  *chi.Router
	config  *config.Config
}

func (a *App) Start() {
	log := logger.Log
	log.Infof("Starting web server on %v", a.config.ServerListenAddr)
	if err := http.ListenAndServe(a.config.ServerListenAddr, *a.router); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func NewApp() (*App, error) {
	config, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("can not create app config: %v", err)
	}
	s, err := filestorage.NewFileStorage(config.MetricsFileName)
	if err != nil {
		return nil, fmt.Errorf("can not init storage: %v", err)
	}
	r := handlers.NewRouter(s)
	return &App{
		storage: s,
		router:  &r,
		config:  config,
	}, nil
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Error starting application: %v", err)
	}
	app.Start()
}
