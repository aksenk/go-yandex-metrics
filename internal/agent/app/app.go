package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/agent/config"
	"github.com/aksenk/go-yandex-metrics/internal/agent/metrics"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/retry"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type App struct {
	Logger                 *zap.SugaredLogger
	Client                 *http.Client
	Config                 *config.Config
	RuntimeRequiredMetrics []string
	ReadyMetrics           chan []models.Metric
	PollCounter            metrics.PollCounter
	ReportTicker           *time.Ticker
}

var errMetricSend = fmt.Errorf("error sending metric")
var errStatusCode = fmt.Errorf("unexpected response status code")
var errReadBody = fmt.Errorf("error reading response body")

func NewApp(client *http.Client, logger *zap.SugaredLogger, config *config.Config) (*App, error) {
	runtimeRequiredMetrics := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}
	return &App{
		Logger:                 logger,
		Client:                 client,
		Config:                 config,
		RuntimeRequiredMetrics: runtimeRequiredMetrics,
		ReadyMetrics:           make(chan []models.Metric, 1),
		PollCounter:            metrics.PollCounter{},
		ReportTicker:           time.NewTicker(config.ReportInterval),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.Logger.Infof("Starting agent")
	a.Logger.Infof("Collecting metric every %s and sending every %v",
		a.Config.PollInterval, a.Config.ReportInterval)
	go a.GetMetrics(ctx)
	a.WaitMetrics(ctx)
	return nil
}

func (a *App) WaitMetrics(ctx context.Context) {
	for {
		select {
		case <-a.ReportTicker.C:
			resultMetrics := <-a.ReadyMetrics

			withRetry := retry.NewRetryer(a.Logger, a.Config.RetryAttempts, time.Duration(a.Config.RetryWaitTime), func(ctx context.Context) error {
				statusCode, err := a.sendBatchMetrics(resultMetrics)
				if err != nil {
					if statusCode < 200 || statusCode >= 500 {
						return err
					}
				}
				return nil
			})

			err := withRetry.Do(ctx)
			if err != nil {
				a.Logger.Errorf("Can not send metrics: %s", err)
				continue
			}

			// обнуляем счетчик PollCounter после успешной отправки метрик
			a.PollCounter.Reset()

			a.Logger.Debugf("Metrics have been sent successfully")

		case <-ctx.Done():
			a.Logger.Infof("Stopping sending metrics")
			return
		}
	}

}

func (a *App) GetMetrics(ctx context.Context) {
	for {
		systemMetrics := metrics.GetSystemMetrics()
		resultMetrics, err := metrics.RemoveUnnecessaryMetrics(systemMetrics, a.RuntimeRequiredMetrics)
		if err != nil {
			a.Logger.Errorf("Can not remove unnecessary metrics: %s", err)
			continue
		}

		a.PollCounter.Inc()

		pollCountMetric, randomValueMetric := metrics.GenerateCustomMetrics(a.PollCounter.Get())
		resultMetrics = append(resultMetrics, pollCountMetric, randomValueMetric)

		select {
		case <-ctx.Done():
			a.Logger.Info("Stopping receiving metrics")
			return
		// если канал пуст - помещаем туда данные
		case a.ReadyMetrics <- resultMetrics:
		// если в канале уже есть данные
		default:
			// вычитываем их
			<-a.ReadyMetrics
			// помещаем туда новые данные
			a.ReadyMetrics <- resultMetrics
		}
		time.Sleep(a.Config.PollInterval)
	}
}

func (a *App) sendBatchMetrics(metrics []models.Metric) (statusCode int, err error) {
	for i := 0; i < len(metrics); i += a.Config.BatchSize {
		end := i + a.Config.BatchSize
		if end > len(metrics) {
			end = len(metrics)
		}
		a.Logger.Debugf("Processing metrics from %v to %v", i, end)

		batch := metrics[i:end]

		jsonData, err := json.Marshal(batch)
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
			h := hmac.New(sha256.New, []byte(cryptKey))
			h.Write(jsonData)
			sign := h.Sum(nil)
			strSign := hex.EncodeToString(sign[:])
			req.Header.Set("HashSHA256", strSign)
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

		a.Logger.Debug("Batch sended")
	}
	return http.StatusOK, nil
}
