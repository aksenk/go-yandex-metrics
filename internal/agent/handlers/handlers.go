package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"io"
	"log"
	"net/http"
	"time"
)

func sendMetrics(metrics []models.Metric, serverURL string) error {
	for _, v := range metrics {
		//url := fmt.Sprintf("%v/%v/%v/%v", serverURL, v.MType, v.ID, v.String())
		url := fmt.Sprintf("%v/update", serverURL)

		marshaledMetric, err := json.Marshal(v)
		requestBody := bytes.NewBuffer(marshaledMetric)
		if err != nil {
			return err
		}
		res, err := http.Post(url, "application/json", requestBody)
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
	for {
		<-ticker.C
		resultMetrics := <-metricsChan
		err := sendMetrics(resultMetrics, serverURL)
		if err != nil {
			log.Printf("Can not send handlers: %s\n", err)
			continue
		}
		log.Printf("Metrics have been sent successfully\n")
	}
}
