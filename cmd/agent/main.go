package main

import (
	"flag"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/agent/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/agent/metrics"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerUseHTTPS bool
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func GetConfig() *Config {
	var serverURL string
	var err error
	serverUseHTTPS := flag.String("s", "false", "Use HTTPS connection to the server")
	serverAddr := flag.String("a", "localhost:8080", "Metrics server address (host:port)")
	pollInterval := flag.String("p", "2", "Interval for scraping metrics (in seconds)")
	reportInterval := flag.String("r", "10", "Interval for sending metrics (in seconds)")

	flag.Parse()
	if e := os.Getenv("USE_HTTPS"); e != "" {
		serverUseHTTPS = &e
	}
	if e := os.Getenv("ADDRESS"); e != "" {
		serverAddr = &e
	}
	if e := os.Getenv("POLL_INTERVAL"); e != "" {
		pollInterval = &e
	}
	if e := os.Getenv("REPORT_INTERVAL"); e != "" {
		reportInterval = &e
	}
	reportIntervalInt, err := strconv.Atoi(*reportInterval)
	if err != nil {
		log.Fatal(err)
	}
	pollIntervalInt, err := strconv.Atoi(*pollInterval)
	if err != nil {
		log.Fatal(err)
	}
	serverUseHTTPSBool, err := strconv.ParseBool(*serverUseHTTPS)
	if err != nil {
		log.Fatal(err)
	}
	if serverUseHTTPSBool {
		serverURL = fmt.Sprintf("https://%v/update", *serverAddr)
	} else {
		serverURL = fmt.Sprintf("http://%v/update", *serverAddr)
	}

	return &Config{
		ServerUseHTTPS: serverUseHTTPSBool,
		ServerURL:      serverURL,
		PollInterval:   time.Second * time.Duration(pollIntervalInt),
		ReportInterval: time.Second * time.Duration(reportIntervalInt),
	}
}

func main() {
	log := logger.Log
	cfg := GetConfig()
	if cfg.PollInterval > cfg.ReportInterval {
		log.Fatalf("Poll interval can not be more that report interval")
	}
	runtimeRequiredMetrics := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}
	reportTicker := time.NewTicker(cfg.ReportInterval)
	metricsChan := make(chan []models.Metric, 1)
	log.Infof("Agent started")
	go metrics.GetMetrics(metricsChan, cfg.PollInterval, runtimeRequiredMetrics)
	handlers.HandleMetrics(metricsChan, reportTicker, cfg.ServerURL)
}
