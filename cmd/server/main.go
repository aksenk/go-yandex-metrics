package main

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/filestorage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"sync"
	"time"
)

type App struct {
	storageType string
	fileStorage *filestorage.FileStorage
	router      *chi.Router
	config      *config.Config
	flushLock   *sync.Mutex
}

func (a *App) Start() error {
	log := logger.Log
	log.Infof("Starting web server on %v", a.config.Server.ListenAddr)

	switch a.config.Storage {
	case "file":
		if a.config.Metrics.StartupRestore {
			err := a.fileStorage.StartupRestore()
			if err != nil {
				return err
			}
		}
	}

	// TODO доделать для 0
	if a.config.Metrics.StoreInterval > 0 {
		go a.BackgroundFlushing()
	}

	if err := http.ListenAndServe(a.config.Server.ListenAddr, *a.router); err != nil {
		return err
	}
	return nil
}

func NewApp(config *config.Config) (*App, error) {
	log := logger.Log
	switch config.Storage {
	case "file":
		log.Infof("Starting %v storage initialization", config.Storage)
		s, err := filestorage.NewFileStorage(&config.FileStorage.FileName)
		if err != nil {
			return nil, fmt.Errorf("can not init fileStorage: %v", err)
		}
		r := handlers.NewRouter(s)
		return &App{
			storageType: config.Storage,
			fileStorage: s,
			router:      &r,
			config:      config,
			flushLock:   &sync.Mutex{},
		}, nil
	default:
		return nil, fmt.Errorf("unknown storage type: %v", config.Storage)
	}
}

// TODO эта реализация только для filestorage
func (a *App) BackgroundFlushing() {
	log := logger.Log
	log.Infof("Starting background metric flushing every %v seconds to the file %v",
		a.config.Metrics.StoreInterval, a.config.FileStorage.FileName)
	flushTicker := time.NewTicker(time.Duration(a.config.Metrics.StoreInterval) * time.Second)
	for {
		<-flushTicker.C
		err := a.FlushMetrics()
		if err != nil {
			log.Errorf("FileStorage.BackgroundFlushing error saving metrics to the disk: %v", err)
			continue
		}
	}
}

func (a *App) FlushMetrics() error {
	switch a.storageType {
	case "file":
		err := a.fileStorage.FlushMetrics()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("flush method is not defined for this type of storage")
	}
	return nil
}

func main() {
	log := logger.Log
	config, err := config.GetConfig()
	if err != nil {
		log.Fatalf("can not create app config: %v", err)
	}
	app, err := NewApp(config)
	if err != nil {
		log.Fatalf("Application initialization error: %v", err)
	}
	err = app.Start()
	if err != nil {
		log.Fatalf("Application launch error: %v", err)
	}
}
