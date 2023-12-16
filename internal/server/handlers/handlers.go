package handlers

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"slices"
	"strings"
	"time"
)

func NewRouter(s storage.Storager) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.Middleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(time.Second * 5))
	r.Get("/", ListAllMetrics(s))
	// TODO почему-то в ответе дублируется текст "Allow: POST" например при запросе GET /update/
	r.Route("/value", func(r chi.Router) {
		r.Get("/", GetMetric(s))
		r.Get("/{type}/", GetMetric(s))
		r.Get("/{type}/{name}", GetMetric(s))
	})
	// TODO вынести работу со storage в middleware?
	r.Route("/update", func(r chi.Router) {
		r.Post("/", UpdateMetric(s))
		r.Post("/{type}/", UpdateMetric(s))
		r.Post("/{type}/{name}/", UpdateMetric(s))
		r.Post("/{type}/{name}/{value}", UpdateMetric(s))
	})
	r.Get("/value/{type}/{name}", GetMetric(s))
	return r
}

func ListAllMetrics(storage storage.Storager) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var list []string
		allMetrics := storage.GetAllMetrics()
		for _, v := range allMetrics {
			list = append(list, fmt.Sprintf("%v=%v", v.Name, v.Value))
		}
		slices.Sort(list)
		r := fmt.Sprintf("<html><head><title>all metrics</title></head>"+
			"<body><h1>List of all metrics</h1><p>%v</p></body></html>\n", strings.Join(list, "</p><p>"))
		writer.Header().Set("Content-type", "text/html")
		writer.Write([]byte(r))
		writer.WriteHeader(http.StatusOK)
	}
}

func GetMetric(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		if metricType == "" {
			http.Error(res, "Missing metric type", http.StatusBadRequest)
			return
		}
		if metricName == "" {
			http.Error(res, "Missing metric name", http.StatusNotFound)
			return
		}
		metric, err := storage.GetMetric(metricName)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error receiving metric: %v", err), http.StatusNotFound)
			return
		}
		if metric.Type != metricType {
			http.Error(res, "Error receiving metric: metric not found", http.StatusNotFound)
			return
		}
		responseText := fmt.Sprintf("%v\n", metric.Value)
		res.Write([]byte(responseText))
		res.WriteHeader(http.StatusOK)
	}
}

func UpdateMetric(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		metricValue := chi.URLParam(req, "value")

		if metricType == "" {
			http.Error(res, "Missing metric type", http.StatusBadRequest)
			return
		}
		if metricName == "" {
			http.Error(res, "Missing metric name", http.StatusNotFound)
			return
		}
		if metricValue == "" {
			http.Error(res, "Missing metric value", http.StatusBadRequest)
			return
		}

		newMetric, err := models.NewMetric(metricName, metricType, metricValue)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error handling '%v' metric: %v", metricName, err),
				http.StatusBadRequest)
			return
		}

		var newCounterValue int64
		var oldCounterValue int64

		if metricType == "counter" {
			if currentMetric, err := storage.GetMetric(metricName); err == nil {
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
		res.WriteHeader(http.StatusOK)

	}
}
