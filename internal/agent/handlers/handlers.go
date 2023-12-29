package handlers

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"io"
	"log"
	"net/http"
	"time"
)

func generateSendURL(m models.Metric, b string) string {
	//if m.Value.(string)
	return fmt.Sprintf("%v/%v/%v/%v", b, m.MType, m.ID, m.Value)
}

func sendMetrics(metrics []models.Metric, serverURL string) error {
	for _, v := range metrics {
		req, err := http.NewRequest(http.MethodPost, generateSendURL(v, serverURL), nil)
		if err != nil {
			return err
		}
		//req.Close = true
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		err = res.Body.Close()
		if err != nil {
			return err
		}
		if res.StatusCode != 200 {
			return fmt.Errorf("unexpected response status code: %v\nError: %v", res.StatusCode, string(body))
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
