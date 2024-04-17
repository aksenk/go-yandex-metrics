package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/agent/config"
	"github.com/aksenk/go-yandex-metrics/internal/agent/metrics"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/signature"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type App struct {
	Logger                 *zap.SugaredLogger
	Client                 *http.Client
	Config                 *config.Config
	RuntimeRequiredMetrics []string
	PSUtilRequiredMetrics  []string
	ReadyMetrics           chan []models.Metric
	PSUtilMetrics          chan []models.Metric
	PollCounter            *metrics.PollCounter
	ReportTicker           *time.Ticker
}

type Response struct {
	Name       string
	StatusCode int
	Err        error
}

var errMetricSend = fmt.Errorf("error sending metric")
var errStatusCode = fmt.Errorf("unexpected response status code")
var errReadBody = fmt.Errorf("error reading response body")

func NewApp(client *http.Client, logger *zap.SugaredLogger, config *config.Config) (*App, error) {
	runtimeRequiredMetrics := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}
	psutilRequiredMetrics := []string{"TotalMemory", "FreeMemory", "CPUutilization1"}

	return &App{
		Logger:                 logger,
		Client:                 client,
		Config:                 config,
		RuntimeRequiredMetrics: runtimeRequiredMetrics,
		PSUtilRequiredMetrics:  psutilRequiredMetrics,
		ReadyMetrics:           make(chan []models.Metric, 1),
		PollCounter:            &metrics.PollCounter{},
		ReportTicker:           time.NewTicker(config.ReportInterval),
	}, nil
}

func (a *App) resultHandler(results <-chan Response) {
	for result := range results {
		if result.Err == nil {
			a.Logger.Infof("Metric %v is successfully sent", result.Name)
		}
		if result.Err != nil {
			a.Logger.Errorf("Error sending %v metric: %v", result.Name, result.Err)
		}
		if result.StatusCode != 200 {
			a.Logger.Errorf("Unexpected status code while sending %v metric: %v", result.Name, result.StatusCode)
		}
	}
}

func (a *App) Run(ctx context.Context) error {
	a.Logger.Infof("Starting agent")
	a.Logger.Infof("Collecting metric every %s and sending every %v", a.Config.PollInterval, a.Config.ReportInterval)

	reportInterval := a.Config.ReportInterval
	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	pollInterval := a.Config.PollInterval
	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	// канал куда складываем метрики для отправки на сервер
	jobs := make(chan models.Metric, a.Config.RateLimit)
	// канал куда получаем результаты отправки метрик
	results := make(chan Response, a.Config.RateLimit)

	for i := 0; i < a.Config.RateLimit; i++ {
		go a.worker(jobs, results)
	}

	go a.resultHandler(results)

	runtimeMetrics := metrics.GetRuntimeMetrics(ctx, a.RuntimeRequiredMetrics, pollTicker)
	customMetrics := metrics.GetCustomMetrics(ctx, a.PollCounter.Get(), pollTicker)
	psutilMetrics := metrics.GetPSUtilMetrics(ctx, pollTicker)

	for {
		select {
		case <-ctx.Done():
			a.Logger.Infof("Stopping agent")
			return nil
		case <-a.ReportTicker.C:
			var allMetrics []models.Metric
			a.Logger.Debug("Report ticker")
			if len(runtimeMetrics) > 0 {
				m := <-runtimeMetrics
				allMetrics = append(allMetrics, m...)
			}
			if len(customMetrics) > 0 {
				m := <-customMetrics
				allMetrics = append(allMetrics, m...)
			}
			if len(psutilMetrics) > 0 {
				m := <-psutilMetrics
				allMetrics = append(allMetrics, m...)
			}
			if len(allMetrics) > 0 {
				a.Logger.Infof("Sending %v metrics", len(allMetrics))
				for _, m := range allMetrics {
					jobs <- m
				}
			} else {
				a.Logger.Infof("No metrics to send")
			}
		}
	}
}

func (a *App) worker(jobs <-chan models.Metric, results chan<- Response) {
	for j := range jobs {
		a.Logger.Infof("Sending metric %v", j.ID)
		statusCode, err := a.sendMetric(j)
		resp := Response{
			StatusCode: statusCode,
			Err:        err,
			Name:       j.ID,
		}
		results <- resp
	}
}

func (a *App) sendMetric(metric models.Metric) (statusCode int, err error) {
	metrics := []models.Metric{metric}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return 0, fmt.Errorf("can not marshal data: %v", err)
	}

	var gzippedBody bytes.Buffer
	w := gzip.NewWriter(&gzippedBody)

	if _, err = w.Write(jsonData); err != nil {
		return 0, fmt.Errorf("can not gzip data: %v", err)
	}

	if err = w.Close(); err != nil {
		return 0, fmt.Errorf("can not close gzip writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.Config.ServerURL, &gzippedBody)
	if err != nil {
		return 0, fmt.Errorf("can not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	cryptKey := a.Config.CryptKey
	if cryptKey != "" {
		sign := signature.GetSignature(jsonData, cryptKey)
		req.Header.Set(signature.SignHeader, sign)
	}

	res, err := a.Client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", errMetricSend, err)
	}

	var resBody []byte
	_, err = res.Body.Read(resBody)
	res.Body.Close()
	if err != nil {
		return 0, fmt.Errorf("%w: %s", errReadBody, err)
	}

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, fmt.Errorf("%w: %v, response: %v", errStatusCode, res.StatusCode, string(resBody))
	}

	return http.StatusOK, nil
}
