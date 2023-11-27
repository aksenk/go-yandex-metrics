package handlers

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func GetMetric(storage storage.Storager) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		return
	}
}

func UpdateMetric(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}
		rawMetricType := chi.URLParam(req, "type")
		rawMetricName := chi.URLParam(req, "name")
		rawMetricValue := chi.URLParam(req, "value")

		if rawMetricType == "" {
			http.Error(res, "Missing metric type", http.StatusBadRequest)
			return
		}
		if rawMetricName == "" {
			http.Error(res, "Missing metric name", http.StatusNotFound)
			return
		}
		if rawMetricValue == "" {
			http.Error(res, "Missing metric value", http.StatusBadRequest)
			return
		}

		newMetric, err := models.NewMetric(rawMetricName, rawMetricType, rawMetricValue)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error handling '%v' metric: %v", rawMetricName, err),
				http.StatusBadRequest)
			return
		}

		var newCounterValue int64
		var oldCounterValue int64

		if rawMetricType == "counter" {
			if currentMetric, err := storage.GetMetric(rawMetricName); err == nil {
				if currentMetric.Type == "counter" {
					oldCounterValue = currentMetric.Value.(int64)
					newCounterValue = newMetric.Value.(int64)
					newCounterValue += oldCounterValue
					newMetric.Value = newCounterValue
				}
			}
		}

		if err := storage.SaveMetric(newMetric); err != nil {
			http.Error(res, fmt.Sprintf("Error saving metric '%v' to storage: %v", newMetric.Name, err),
				http.StatusBadRequest)
			return
		}

		res.Write([]byte(fmt.Sprintf("Metric saved successfully\ntype: %v, name: %v, value: %v\n",
			newMetric.Type, newMetric.Name, newMetric.Value)))
		res.Write([]byte(fmt.Sprintf("%+v\n", storage)))
	}
}
