package app

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/filestorage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

type App struct {
	storage storage.Storager
	router  *chi.Router
	config  *config.Config
}

func (a *App) Start() error {
	log := logger.Log
	if a.config.Metrics.StartupRestore {
		err := a.storage.StartupRestore()
		if err != nil {
			return err
		}
	}

	if a.config.Metrics.StoreInterval > 0 {
		go a.BackgroundFlusher()
	}

	log.Infof("Starting web server on %v", a.config.Server.ListenAddr)
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
		synchronousFlush := false
		if config.Metrics.StoreInterval == 0 {
			log.Infof("Synchronous flushing is enabled")
			synchronousFlush = true
		}
		s, err := filestorage.NewFileStorage(config.FileStorage.FileName, synchronousFlush)
		if err != nil {
			return nil, fmt.Errorf("can not init fileStorage: %v", err)
		}
		r := handlers.NewRouter(s)
		return &App{
			storage: s,
			router:  &r,
			config:  config,
			//flushLock: &sync.Mutex{},
		}, nil
	default:
		return nil, fmt.Errorf("unknown storage type: %v", config.Storage)
	}
}

func (a *App) BackgroundFlusher() {
	log := logger.Log
	log.Infof("Starting background metric flushing every %v seconds", a.config.Metrics.StoreInterval)
	flushTicker := time.NewTicker(time.Duration(a.config.Metrics.StoreInterval) * time.Second)
	for {
		<-flushTicker.C
		err := a.storage.FlushMetrics()
		if err != nil {
			log.Errorf("FileStorage.BackgroundFlusher error saving metrics to the disk: %v", err)
			continue
		}
	}
}
