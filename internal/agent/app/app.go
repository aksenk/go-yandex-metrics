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
	"io"
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
		err := a.sendMetrics(resultMetrics)
		if err != nil {
			a.Logger.Errorf("Can not send handlers: %s", err)
			continue
		}
		a.Logger.Debugf("RuntimeRequiredMetrics have been sent successfully")
	}
}

func (a *App) GetMetrics() {
	pollCounter := int64(0)
	var pollCountMetric, randomValueMetric models.Metric
	for {
		systemMetrics := metrics.GetSystemMetrics()
		a.Logger.Infof("RuntimeRequiredMetrics have been collected successfully")
		resultMetrics, err := metrics.RemoveUnnecessaryMetrics(systemMetrics, a.RuntimeRequiredMetrics)
		if err != nil {
			a.Logger.Errorf("Can not remove unnecessary metrics: %s", err)
			continue
		}
		a.Logger.Infof("RuntimeRequiredMetrics have been filtered successfully")

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

func (a *App) sendMetrics(metrics []models.Metric) error {
	var req *http.Request
	for _, v := range metrics {
		client := http.Client{
			Timeout: 10 * time.Second,
		}

		jsonData, err := json.Marshal(v)
		if err != nil {
			a.Logger.Errorf("Can not marshal data: %v", err)
			return fmt.Errorf("can not marshal data: %v", err)
		}

		var gzippedBody bytes.Buffer
		w := gzip.NewWriter(&gzippedBody)
		if _, err := w.Write(jsonData); err != nil {
			a.Logger.Errorf("Can not gzip data: %v", err)
			return fmt.Errorf("can not gzip data: %v", err)
		}
		if err = w.Close(); err != nil {
			a.Logger.Errorf("Can not close writer: %v", err)
			return fmt.Errorf("can not close writer: %v", err)
		}
		req, err = http.NewRequest(http.MethodPost, a.Config.ServerURL, &gzippedBody)
		if err != nil {
			a.Logger.Errorf("Can not create http request: %v", err)
			return fmt.Errorf("can not create http request: %v", err)
		}
		req.Header.Set("Content-Encoding", "gzip")

		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		err = res.Body.Close()
		if err != nil {
			return err
		}
		if res.StatusCode != 200 {
			return fmt.Errorf("unexpected response status code: %v\nError: %v", res.StatusCode, string(responseBody))
		}
	}
	return nil
}
