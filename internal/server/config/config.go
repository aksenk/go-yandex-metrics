package config

import (
	"flag"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"os"
	"strconv"
)

type Config struct {
	ServerListenAddr      string
	MetricsStoreInterval  int
	MetricsFileName       string
	MetricsStartupRestore bool
}

func GetConfig() (*Config, error) {
	log := logger.Log
	serverListenAddr := flag.String("a", "localhost:8080", "host:port for server listening")
	metricsStoreInterval := flag.Int("i", 300, "Period in seconds between flushing metrics to the disk")
	metricsFileName := flag.String("f", "/tmp/metrics-db.json", "Path to the file for storing metrics")
	metricsStratupRestore := flag.Bool("r", true, "Restoring metrics from the file at startup")
	flag.Parse()
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
	if e := os.Getenv("FILE_STORAGE_PATH"); e != "" {
		metricsFileName = &e
	}
	if e := os.Getenv("RESTORE"); e != "" {
		v, err := strconv.ParseBool(e)
		if err != nil {
			log.Errorf("GetConfig: can not parse value of 'RESTORE' (%v) environment variable: %v", e, err)
			return nil, fmt.Errorf("GetConfig: can not parse value of 'RESTORE' (%v) environment variable: %v", e, err)
		}
		metricsStratupRestore = &v
	}
	return &Config{
		ServerListenAddr:      *serverListenAddr,
		MetricsStoreInterval:  *metricsStoreInterval,
		MetricsFileName:       *metricsFileName,
		MetricsStartupRestore: *metricsStratupRestore,
	}, nil
}
