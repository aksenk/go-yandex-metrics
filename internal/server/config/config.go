package config

import (
	"flag"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"os"
	"strconv"
)

type Config struct {
	Storage     string
	Server      serverConfig
	Metrics     metricsConfig
	FileStorage fileStorageConfig
	Database    databaseConfig
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
	// TODO storage type должно иметь фиксированные значения
	storage := flag.String("s", "file", "Storage type")
	serverListenAddr := flag.String("a", "localhost:8080", "host:port for server listening")
	metricsStoreInterval := flag.Int("i", 300, "Period in seconds between flushing metrics to the disk")
	fileStorageFileName := flag.String("f", "/tmp/metrics-db.json", "Path to the file for storing metrics")
	fileStorageStartupRestore := flag.Bool("r", true, "Restoring metrics from the file at startup")
	databaseDSN := flag.String("d", "", "Database DSN string")
	// TODO database type должно иметь фиксированные значения
	databaseType := flag.String("dt", "postgres", "Database type")

	flag.Parse()

	if e := os.Getenv("DATABASE_TYPE"); e != "" {
		databaseType = &e
	}
	if e := os.Getenv("DATABASE_DSN"); e != "" {
		databaseDSN = &e
	}
	if e := os.Getenv("STORAGE"); e != "" {
		storage = &e
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
	if *databaseType != "postgres" {
		return nil, fmt.Errorf("only postgres database is supported")
	}
	return &Config{
		Storage: *storage,
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
		Database: databaseConfig{
			DSN:  *databaseDSN,
			Type: *databaseType,
		},
	}, nil
}
