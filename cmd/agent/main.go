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

	pollIntervalDuration := time.Second * time.Duration(pollInterval)
	reportIntervalDuration := time.Second * time.Duration(reportInterval)
	reportTicker := time.NewTicker(reportIntervalDuration)
	metricsChan := make(chan []models.Metric, 1)
	go metrics.GetMetrics(metricsChan, pollIntervalDuration)
	handlers.HandleMetrics(metricsChan, reportTicker, serverURL)
}
