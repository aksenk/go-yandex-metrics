package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"io"
	"net/http"
	"time"
)

func sendMetrics(metrics []models.Metric, serverURL string) error {
	log := logger.Log
	var req *http.Request
	for _, v := range metrics {
		client := http.Client{
			Timeout: 10 * time.Second,
		}

		jsonData, err := json.Marshal(v)
		if err != nil {
			log.Errorf("Can not marshal data: %v", err)
			return fmt.Errorf("can not marshal data: %v", err)
		}

		var gzippedBody bytes.Buffer
		w := gzip.NewWriter(&gzippedBody)
		if _, err := w.Write(jsonData); err != nil {
			log.Errorf("Can not gzip data: %v", err)
			return fmt.Errorf("can not gzip data: %v", err)
		}
		if err := w.Close(); err != nil {
			log.Errorf("Can not close writer: %v", err)
			return fmt.Errorf("can not close writer: %v", err)
		}
		req, err = http.NewRequest(http.MethodPost, serverURL, &gzippedBody)
		if err != nil {
			log.Errorf("Can not create http request: %v", err)
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

func HandleMetrics(metricsChan chan []models.Metric, ticker *time.Ticker, serverURL string) {
	log := logger.Log
	for {
		<-ticker.C
		resultMetrics := <-metricsChan
		err := sendMetrics(resultMetrics, serverURL)
		if err != nil {
			log.Errorf("Can not send handlers: %s", err)
			continue
		}
		log.Debugf("Metrics have been sent successfully")
	}
}
