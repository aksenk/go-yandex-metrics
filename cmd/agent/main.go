package main

import (
	"github.com/aksenk/go-yandex-metrics/internal/agent/config"
	"github.com/aksenk/go-yandex-metrics/internal/agent/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/agent/metrics"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"time"
)

func main() {
	log := logger.Log
	cfg := config.NewConfig()
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
