package config

import (
	"flag"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"os"
	"strconv"
)

type Config struct {
	Storage         storage.SType
	LogLevel        string
	Server          ServerConfig
	Metrics         MetricsConfig
	FileStorage     FileStorageConfig
	PostgresStorage PostgresConfig
	RetryConfig     RetryConfig
	CryptConfig     CryptConfig
}

type RetryConfig struct {
	RetryAttempts int
	RetryWaitTime int
}

type ServerConfig struct {
	ListenAddr string
}

type MetricsConfig struct {
	StoreInterval  int
	StartupRestore bool
}

type FileStorageConfig struct {
	FileName string
}

type PostgresConfig struct {
	DSN           string
	Type          string
	MigrationsDir string
}

type CryptConfig struct {
	Key string
}

func GetConfig() (*Config, error) {
	log, err := logger.NewLogger("info")
	if err != nil {
		return nil, err
	}
	logLevel := flag.String("l", "info", "Logger level")
	serverListenAddr := flag.String("a", "localhost:8080", "host:port for server listening")
	metricsStoreInterval := flag.Int("i", 300, "Period in seconds between flushing metrics to the disk (file storage)")
	fileStorageFileName := flag.String("f", "", "Path to the file for storing metrics (file storage)")
	fileStorageStartupRestore := flag.Bool("r", true, "Restoring metrics from the file at startup (file storage)")
	databaseDSN := flag.String("d", "", "Postgres connection DSN string (database storage)")
	cryptKey := flag.String("k", "", "Crypt key for signing requests")

	retryAttempts := 3
	retryWaitTime := 2
	migrationsDir := "./migrations/postgres"

	flag.Parse()

	if e := os.Getenv("KEY"); e != "" {
		cryptKey = &e
	}
	if e := os.Getenv("LOG_LEVEL"); e != "" {
		logLevel = &e
	}
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
	s := storage.MemoryStorage
	if databaseDSN != nil && *databaseDSN != "" {
		s = storage.PostgresStorage
	} else if fileStorageFileName != nil && *fileStorageFileName != "" {
		s = storage.FileStorage
	}
	return &Config{
		Storage:  s,
		LogLevel: *logLevel,
		Server: ServerConfig{
			ListenAddr: *serverListenAddr,
		},
		Metrics: MetricsConfig{
			StoreInterval:  *metricsStoreInterval,
			StartupRestore: *fileStorageStartupRestore,
		},
		FileStorage: FileStorageConfig{
			FileName: *fileStorageFileName,
		},
		PostgresStorage: PostgresConfig{
			DSN:           *databaseDSN,
			Type:          "postgres",
			MigrationsDir: migrationsDir,
		},
		RetryConfig: RetryConfig{
			RetryAttempts: retryAttempts,
			RetryWaitTime: retryWaitTime,
		},
		CryptConfig: CryptConfig{
			Key: *cryptKey,
		},
	}, nil
}
