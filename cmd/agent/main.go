package main

import (
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/agent/handlers"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/agent/metrics"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"time"
)

func main() {
	serverURL := "http://localhost:8080/update"
	pollInterval := 2
	reportInterval := 10
	runtimeRequiredMetrics := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}
	pollIntervalDuration := time.Second * time.Duration(pollInterval)
	reportIntervalDuration := time.Second * time.Duration(reportInterval)
	reportTicker := time.NewTicker(reportIntervalDuration)
	metricsChan := make(chan []models.Metric, 1)
	go metrics.GetMetrics(metricsChan, pollIntervalDuration, runtimeRequiredMetrics)
	handlers.HandleMetrics(metricsChan, reportTicker, serverURL)
}
