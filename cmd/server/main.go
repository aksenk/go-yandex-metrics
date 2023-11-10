package main

import (
	"fmt"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/server/server"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/server/storage"
	"log"
	"net/http"
	"strings"
)

func updateMetric(storage *storage.MemStorage) http.HandlerFunc {
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

		var metricType string
		switch mt := splitURL[2]; mt {
		case "gauge":
			metricType = "gauge"
		case "counter":
			metricType = "counter"
		default:
			http.Error(res, "incorrect metric type", http.StatusBadRequest)
			return
		}
		metricName := splitURL[3]
		metricValue := splitURL[4]

		m := storage.NewMetric(metricName, metricType, metricValue)

		err := storage.AddMetric(*m)
		if err != nil {
			http.Error(res, "can not convert metric value", http.StatusBadRequest)
			return
		}

		res.Write([]byte(fmt.Sprintf("type: %storage, name: %storage, value: %storage\n", metricType, metricName, metricValue)))
		res.Write([]byte(fmt.Sprintf("%+v\n", storage)))

	}
}

func main() {
	listenAddr := "localhost:8080"
	listenPath := "/update/"
	storage := storage.NewStorage()
	// TODO не нравится как сделано с сервером, но пока не знаю как по другому
	err := server.NewServer(listenAddr, listenPath, updateMetric(storage))
	if err != nil {
		log.Fatal(err)
	}
}
