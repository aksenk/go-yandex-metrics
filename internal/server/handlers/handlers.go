package handlers

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"net/http"
	"strconv"
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

		//if splitURL[2] != "counter" && splitURL[2] != "gauge" {
		//	http.Error(res, fmt.Sprintf("Incorrect newMetric type: %v", splitURL[2]), http.StatusBadRequest)
		//	return
		//}

		var newMetric models.Metric

		rawMetricName := splitURL[3]
		rawMetricType := splitURL[2]
		rawMetricValue := splitURL[4]

		switch rawMetricType {
		case "gauge":
			if newFloat64Value, err := strconv.ParseFloat(rawMetricValue, 64); err == nil {
				newMetric = models.Metric{
					Name:  splitURL[3],
					Type:  splitURL[2],
					Value: newFloat64Value,
				}
			} else {
				http.Error(res, fmt.Sprintf("Can not parse metric value: %v. Error: %v",
					rawMetricValue, err), http.StatusBadRequest)
			}
		case "counter":
			var intValueNew int64
			var intValueOld int64

			switch newFloat64Value := s.Metrics[m.Name].Value.(type) {
			case int64:
				intValueOld = tmp
			}

			if tmp, err := strconv.ParseInt(m.Value.(string), 10, 64); err == nil {
				intValueNew = tmp
			} else {
				return errMetricValue
			}

			newValue := intValueNew + intValueOld
			m.Value = newValue
			s.Metrics[m.Name] = m
		default:
			http.Error(res, fmt.Sprintf("Incorrect metric type: %v", rawMetricType), http.StatusBadRequest)
		}

		if err := storage.AddMetric(newMetric); err != nil {
			http.Error(res, "can not convert newMetric value", http.StatusBadRequest)
			return
		}

		res.Write([]byte(fmt.Sprintf("type: %storage, name: %storage, value: %storage\n",
			m.Type, m.Name, m.Value)))
		res.Write([]byte(fmt.Sprintf("%+v\n", storage)))
	}
}
