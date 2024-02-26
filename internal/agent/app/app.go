package app

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/agent/config"
	"github.com/aksenk/go-yandex-metrics/internal/agent/metrics"
	"github.com/aksenk/go-yandex-metrics/internal/models"
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
	PollCounter            int64
	ReportTicker           *time.Ticker
}

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
		PollCounter:            0,
		ReportTicker:           time.NewTicker(config.ReportInterval),
	}, nil
}

func (a *App) Run() error {
	a.Logger.Infof("Starting agent")
	a.Logger.Infof("Collecting metric every %s and sending every %v",
		a.Config.PollInterval, a.Config.ReportInterval)
	go a.GetMetrics()
	a.WaitMetrics()
	return nil
}

func (a *App) WaitMetrics() {
	for {
		<-a.ReportTicker.C
		resultMetrics := <-a.ReadyMetrics
		err := a.sendBatchMetrics(resultMetrics)
		if err != nil {
			a.Logger.Errorf("Can not send metrics: %s", err)
			continue
		}
		a.Logger.Debugf("Metrics have been sent successfully")
	}
}

func (a *App) GetMetrics() {
	pollCounter := int64(0)
	var pollCountMetric, randomValueMetric models.Metric
	for {
		systemMetrics := metrics.GetSystemMetrics()
		resultMetrics, err := metrics.RemoveUnnecessaryMetrics(systemMetrics, a.RuntimeRequiredMetrics)
		if err != nil {
			a.Logger.Errorf("Can not remove unnecessary metrics: %s", err)
			continue
		}

		metrics.GenerateCustomMetrics(&pollCountMetric, &randomValueMetric, &pollCounter)
		resultMetrics = append(resultMetrics, pollCountMetric, randomValueMetric)
		select {
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

func (a *App) sendBatchMetrics(metrics []models.Metric) error {
	for i := 0; i < len(metrics); i += a.Config.BatchSize {
		end := i + a.Config.BatchSize
		if end > len(metrics) {
			end = len(metrics)
		}
		a.Logger.Debugf("Processing metrics from %v to %v", i, end)

		batch := metrics[i:end]

		jsonData, err := json.Marshal(batch)
		if err != nil {
			return fmt.Errorf("can not marshal data: %v", err)
		}

		var gzippedBody bytes.Buffer
		w := gzip.NewWriter(&gzippedBody)

		if _, err = w.Write(jsonData); err != nil {
			return fmt.Errorf("can not gzip data: %v", err)
		}

		if err = w.Close(); err != nil {
			return fmt.Errorf("can not close gzip writer: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, a.Config.ServerURL, &gzippedBody)
		if err != nil {
			return fmt.Errorf("can not create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		res, err := a.Client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending metric: %s", err)
		}

		var resBody []byte
		_, err = res.Body.Read(resBody)
		res.Body.Close()
		if err != nil {
			return fmt.Errorf("error reading response body: %s", err)
		}

		if res.StatusCode != 200 {
			return fmt.Errorf("unexpected response status code: %v, response: %v", res.StatusCode, string(resBody))
		}

		a.Logger.Debugf("Batch sended")
	}
	return nil
}
