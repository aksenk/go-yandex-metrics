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
		if a.config.StartupRestore {
			err := a.fileStorage.StartupRestore()
			if err != nil {
				return err
			}
		}
	}

	//if a.config.Metrics.StoreInterval > 0 {
	//	go a.BackgroundFlushing()
	//}

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

//func (a *App) BackgroundFlushing() {
//	log := logger.Log
//	log.Infof("Starting background metric flushing every %v seconds to the file %v", f.flushInterval, f.file.Name())
//	flushTicker := time.NewTicker(time.Duration(f.flushInterval) * time.Second)
//	for {
//		<-flushTicker.C
//		err := f.FlushMetrics()
//		if err != nil {
//			log.Errorf("FileStorage.BackgroundFlushing error saving metrics to the disk: %v", err)
//			continue
//		}
//	}
//
//}

//func (a *App) FlushMetrics() error {
//	log := logger.Log
//	counter := 0
//	f.flushLock.Lock()
//	defer f.flushLock.Unlock()
//	log.Infof("Start collecting metrics for flushing to the file")
//	for _, v := range f.memStorage.Metrics {
//		jsonMetric, err := json.Marshal(v)
//		if err != nil {
//			log.Errorf("FileStorage.FlushMetrics: can not marsgal metric '%v': %v", v, err)
//			return fmt.Errorf("FileStorage.FlushMetrics: can not marshal metric '%v': %v", v, err)
//		}
//		_, err = f.writer.Write(jsonMetric)
//		if err != nil {
//			log.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
//			return fmt.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
//		}
//		err = f.writer.WriteByte('\n')
//		if err != nil {
//			log.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
//			return fmt.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
//		}
//		counter++
//	}
//	f.writer.Flush()
//	log.Infof("Successfully flushed %v metrics to the file", counter)
//	return nil
//}

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
