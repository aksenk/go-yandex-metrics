package main

import (
	"flag"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/agent/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/agent/metrics"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"log"
	"time"
)

func main() {
	var serverURL string
	serverHTTPScheme := flag.Bool("s", false, "Use HTTPS connection to the server")
	serverAddr := flag.String("a", "localhost:8080", "Metrics server address (host:port)")
	pollInterval := flag.Int("p", 2, "Interval for scraping metrics (in seconds)")
	reportInterval := flag.Int("r", 10, "Interval for sending metrics (in seconds)")
	flag.Parse()
	if *serverHTTPScheme {
		serverURL = fmt.Sprintf("https://%v/update", *serverAddr)
	} else {
		serverURL = fmt.Sprintf("http://%v/update", *serverAddr)
	}
	if *pollInterval > *reportInterval {
		log.Fatalf("Poll interval can not be more that report interval")
	}
	runtimeRequiredMetrics := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}
	pollIntervalDuration := time.Second * time.Duration(*pollInterval)
	reportIntervalDuration := time.Second * time.Duration(*reportInterval)
	reportTicker := time.NewTicker(reportIntervalDuration)
	metricsChan := make(chan []models.Metric, 1)
	go metrics.GetMetrics(metricsChan, pollIntervalDuration, runtimeRequiredMetrics)
	handlers.HandleMetrics(metricsChan, reportTicker, serverURL)
}
