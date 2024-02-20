package config

import (
	"flag"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"os"
	"strconv"
)

type Config struct {
	Storage         string
	Server          serverConfig
	Metrics         metricsConfig
	FileStorage     fileStorageConfig
	PostgresStorage databaseConfig
}

type serverConfig struct {
	ListenAddr string
}

type metricsConfig struct {
	StoreInterval  int
	StartupRestore bool
}

type fileStorageConfig struct {
	FileName string
}

type databaseConfig struct {
	DSN  string
	Type string
}

func GetConfig() (*Config, error) {
	log := logger.Log
	serverListenAddr := flag.String("a", "localhost:8080", "host:port for server listening")
	metricsStoreInterval := flag.Int("i", 300, "Period in seconds between flushing metrics to the disk (file storage)")
	fileStorageFileName := flag.String("f", "", "Path to the file for storing metrics (file storage)")
	fileStorageStartupRestore := flag.Bool("r", false, "Restoring metrics from the file at startup (file storage)")
	databaseDSN := flag.String("d", "", "Postgres connection DSN string (database storage)")

	flag.Parse()

	if e := os.Getenv("DATABASE_DSN"); e != "" {
		databaseDSN = &e
	}
	if e := os.Getenv("ADDRESS"); e != "" {
		serverListenAddr = &e
	}
	if e := os.Getenv("STORE_INTERVAL"); e != "" {
		v, err := strconv.Atoi(e)
		if err != nil {
			log.Errorf("GetConfig: can not parse value of 'STORE_INTERVAL' (%v) environment variable: %v", e, err)
			return nil, fmt.Errorf("GetConfig: can not parse value of 'STORE_INTERVAL' (%v) environment variable: %v", e, err)
		}
		metricsStoreInterval = &v
	}
	if *metricsStoreInterval < 0 {
		return nil, fmt.Errorf("store interval must be zero or greather")
	}
	if e := os.Getenv("FILE_STORAGE_PATH"); e != "" {
		fileStorageFileName = &e
	}
	if e := os.Getenv("RESTORE"); e != "" {
		v, err := strconv.ParseBool(e)
		if err != nil {
			log.Errorf("GetConfig: can not parse value of 'RESTORE' (%v) environment variable: %v", e, err)
			return nil, fmt.Errorf("GetConfig: can not parse value of 'RESTORE' (%v) environment variable: %v", e, err)
		}
		fileStorageStartupRestore = &v
	}
	storage := "file"
	if databaseDSN != nil && *databaseDSN != "" {
		storage = "postgres"
	} else if fileStorageFileName != nil && *fileStorageFileName != "" {
		storage = "file"
	} else {
		storage = "memory"
	}
	return &Config{
		Storage: storage,
		Server: serverConfig{
			ListenAddr: *serverListenAddr,
		},
		Metrics: metricsConfig{
			StoreInterval:  *metricsStoreInterval,
			StartupRestore: *fileStorageStartupRestore,
		},
		FileStorage: fileStorageConfig{
			FileName: *fileStorageFileName,
		},
		PostgresStorage: databaseConfig{
			DSN: *databaseDSN,
		},
	}, nil
}
