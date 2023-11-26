package handlers

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"net/http"
	"strings"
)

func UpdateMetric(storage storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		splitURL := strings.Split(req.URL.Path, "/")
		if len(splitURL) < 5 {
			http.Error(res, "Incorrect fields in request. "+
				"It should be like '/update/metric_type/metric_name/metric_value'", http.StatusNotFound)
			return
		}
		rawMetricName := splitURL[3]
		rawMetricType := splitURL[2]
		rawMetricValue := splitURL[4]

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
