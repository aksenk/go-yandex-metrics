package handlers

import (
	"fmt"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/server/storage"
	"net/http"
	"strings"
)

func UpdateMetric(storage *storage.MemStorage) http.HandlerFunc {
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

		m := models.Metric{
			Name:  splitURL[3],
			Type:  splitURL[2],
			Value: splitURL[4],
		}

		if err := storage.AddMetric(m); err != nil {
			http.Error(res, "can not convert metric value", http.StatusBadRequest)
			return
		}

		res.Write([]byte(fmt.Sprintf("type: %storage, name: %storage, value: %storage\n",
			m.Type, m.Name, m.Value)))
		res.Write([]byte(fmt.Sprintf("%+v\n", storage)))
	}
}
