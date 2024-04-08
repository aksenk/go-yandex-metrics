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

func (a *App) resultHandler(jobs chan<- models.Metric, results <-chan Response) {
	for result := range results {
		if result.Err == nil {
			a.Logger.Errorf("Metric %v is successfully sent", result.Name)
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

	jobs := make(chan models.Metric, a.Config.RateLimit)
	results := make(chan Response, a.Config.RateLimit)

	for i := 0; i < a.Config.RateLimit; i++ {
		go a.worker(jobs, results)
	}

	go a.resultHandler(jobs, results)

	readyMetrics := a.GetMetrics(ctx, a.PollCounter)
	for {
		select {
		case <-ctx.Done():
			a.Logger.Infof("Stopping agent")
			return nil
		case <-a.ReportTicker.C:
			a.Logger.Infof("report ticker")
			metricCount := len(readyMetrics)
			a.Logger.Infof("metric count: %v", metricCount)
			if metricCount > 0 {
				m := <-readyMetrics
				for _, metric := range m {
					jobs <- metric
				}
			} else {
				a.Logger.Infof("No metrics to send")
			}
		}
	}
}

func (a *App) GetMetrics(ctx context.Context, counter *metrics.PollCounter) chan []models.Metric {
	pollInterval := a.Config.PollInterval
	pollTicker := time.NewTicker(pollInterval)
	reportInterval := a.Config.ReportInterval
	reportTicker := time.NewTicker(reportInterval)

	resultChan := make(chan []models.Metric, 1)

	go func() {
		defer pollTicker.Stop()
		defer reportTicker.Stop()

		for {
			a.PollCounter.Inc()

			select {

			case <-ctx.Done():
				a.Logger.Info("Stopping receiving metrics")
				return

			case <-pollTicker.C:
				var result []models.Metric

				runtimeMetrics, err := metrics.GetRuntimeMetrics(a.RuntimeRequiredMetrics)
				if err != nil {
					a.Logger.Errorf("Can not get runtime metrics: %s", err)
					continue
				}
				result = append(result, runtimeMetrics...)

				customMetrics := metrics.GetCustomMetrics(counter.Get())
				result = append(result, customMetrics...)

				psutilMetrics, err := metrics.GetPSUtilMetrics(ctx)
				if err != nil {
					a.Logger.Errorf("Can not get psutil metrics: %s", err)
					continue
				}
				result = append(result, psutilMetrics...)

				select {
				// если канал пуст записываем данные
				case resultChan <- result:
				// если в канале уже есть данные
				default:
					// вычитываем их
					<-resultChan
					// помещаем туда новые данные
					resultChan <- result
				}

				//a.Logger.Infof("reset counter")
				//a.PollCounter.Reset()
			}
		}
	}()
	return resultChan
}

//func (a *App) sendBatchMetrics(metrics []models.Metric) (statusCode int, err error) {
//	for i := 0; i < len(metrics); i += a.Config.BatchSize {
//		end := i + a.Config.BatchSize
//		if end > len(metrics) {
//			end = len(metrics)
//		}
//		a.Logger.Debugf("Processing metrics from %v to %v", i, end)
//
//		batch := metrics[i:end]
//
//		jsonData, err := json.Marshal(batch)
//		if err != nil {
//			return 0, fmt.Errorf("can not marshal data: %v", err)
//		}
//
//		var gzippedBody bytes.Buffer
//		w := gzip.NewWriter(&gzippedBody)
//
//		if _, err = w.Write(jsonData); err != nil {
//			return 0, fmt.Errorf("can not gzip data: %v", err)
//		}
//
//		if err = w.Close(); err != nil {
//			return 0, fmt.Errorf("can not close gzip writer: %v", err)
//		}
//
//		req, err := http.NewRequest(http.MethodPost, a.Config.ServerURL, &gzippedBody)
//		if err != nil {
//			return 0, fmt.Errorf("can not create request: %v", err)
//		}
//		req.Header.Set("Content-Type", "application/json")
//		req.Header.Set("Content-Encoding", "gzip")
//
//		cryptKey := a.Config.CryptKey
//		if cryptKey != "" {
//			sign := signature.GetSignature(jsonData, cryptKey)
//			req.Header.Set(signature.SignHeader, sign)
//		}
//
//		res, err := a.Client.Do(req)
//		if err != nil {
//			return 0, fmt.Errorf("%w: %s", errMetricSend, err)
//		}
//
//		var resBody []byte
//		_, err = res.Body.Read(resBody)
//		res.Body.Close()
//		if err != nil {
//			return 0, fmt.Errorf("%w: %s", errReadBody, err)
//		}
//
//		if res.StatusCode != http.StatusOK {
//			return res.StatusCode, fmt.Errorf("%w: %v, response: %v", errStatusCode, res.StatusCode, string(resBody))
//		}
//
//		a.Logger.Debug("Batch sended")
//	}
//	return http.StatusOK, nil
//}

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
	var metrics []models.Metric
	metrics = append(metrics, metric)

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
