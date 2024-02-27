package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/filestorage"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/postgres"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type App struct {
	storage storage.Storager
	router  *chi.Router
	config  *config.Config
	server  *http.Server
	logger  *zap.SugaredLogger
}

func (a *App) Start(ctx context.Context) error {
	if a.config.Storage == storage.FileStorage {
		if a.config.Metrics.StartupRestore {
			err := a.storage.StartupRestore(ctx)
			if err != nil {
				return err
			}
		}
		if a.config.Metrics.StoreInterval > 0 {
			go a.BackgroundFlusher(ctx)
		}
	}

	a.logger.Infof("Starting web server on %v", a.config.Server.ListenAddr)
	err := a.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Infof("Closing web server")
	err := a.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	if a.config.Storage == "file" {
		err = a.storage.FlushMetrics()
		if err != nil {
			return err
		}
	}
	a.logger.Infof("Closing storage")
	err = a.storage.Close()
	if err != nil {
		return err
	}

	a.logger.Infof("Shutdown completed")
	return nil
}

func NewApp(config *config.Config) (*App, error) {
	logger, err := logger.NewLogger(config.LogLevel)
	if err != nil {
		return nil, err
	}
	var router chi.Router
	var s storage.Storager

	logger.Infof("Starting %v storage initialization", config.Storage)
	synchronousFlush := false
	if config.Metrics.StoreInterval == 0 {
		logger.Infof("Synchronous flushing is enabled")
		synchronousFlush = true
	}

	switch config.Storage {

	case storage.MemoryStorage:
		s = memstorage.NewMemStorage(logger)

	case storage.FileStorage:
		s, err = filestorage.NewFileStorage(config.FileStorage.FileName, synchronousFlush, logger)
		if err != nil {
			return nil, fmt.Errorf("can not init fileStorage: %v", err)
		}

	case storage.PostgresStorage:
		pgs, err := postgres.NewPostgresStorage(config.PostgresStorage.DSN, logger)
		if err != nil {
			return nil, fmt.Errorf("can not init postgresStorage: %v", err)
		}
		err = pgs.Status(context.TODO())
		if err != nil {
			return nil, err
		}
		logger.Infof("Starting database migrations")
		version, dirty, err := postgres.RunMigrations("./migrations/postgres", pgs.Conn)
		if err != nil {
			return nil, fmt.Errorf("can not run migrations: %v", err)
		}
		if dirty {
			return nil, fmt.Errorf("database version %v have dirty status", version)
		}
		s = pgs
		logger.Infof("Database is up to date. Version: %v", version)

	default:
		return nil, fmt.Errorf("unknown storage type: %v", config.Storage)
	}

	router = handlers.NewRouter(s, logger)
	srv := &http.Server{
		Addr:    config.Server.ListenAddr,
		Handler: router,
	}
	return &App{
		storage: s,
		router:  &router,
		config:  config,
		server:  srv,
		logger:  logger,
	}, nil
}

func (a *App) BackgroundFlusher(ctx context.Context) {
	a.logger.Infof("Starting background metric flushing every %v seconds", a.config.Metrics.StoreInterval)
	flushTicker := time.NewTicker(time.Duration(a.config.Metrics.StoreInterval) * time.Second)
	for {
		select {
		case <-flushTicker.C:
			err := a.storage.FlushMetrics()
			if err != nil {
				a.logger.Errorf("FileStorage.BackgroundFlusher error saving metrics to the disk: %v", err)
				continue
			}
		case <-ctx.Done():
			a.logger.Infof("BackgroundFlusher stopped")
			return
		}
	}
}
