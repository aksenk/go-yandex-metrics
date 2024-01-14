package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
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
		r.Post("/", JSONGetMetricHandler(s))
		r.Get("/", PlainGetMetricHandler(s))
		r.Get("/{type}/", PlainGetMetricHandler(s))
		r.Get("/{type}/{name}", PlainGetMetricHandler(s))
	})
	// TODO вынести работу со storage в middleware?
	r.Route("/update", func(r chi.Router) {
		r.Post("/", JSONUpdaterHandler(s))
		// TODO вернуть
		r.Post("/{type}/", PlainUpdaterHandler(s))
		r.Post("/{type}/{name}/", PlainUpdaterHandler(s))
		r.Post("/{type}/{name}/{value}", PlainUpdaterHandler(s))
	})
	r.Get("/value/{type}/{name}", PlainGetMetricHandler(s))
	return r
}

func ListAllMetrics(storage storage.Storager) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var list []string
		allMetrics := storage.GetAllMetrics()
		for _, v := range allMetrics {
			if v.MType == "gauge" {
				list = append(list, fmt.Sprintf("%v=%v", v.ID, *v.Value))

			} else if v.MType == "counter" {
				list = append(list, fmt.Sprintf("%v=%v", v.ID, *v.Delta))
			} else {
				logger.Log.Errorf("Unknown metric type '%v' for metric '%v' when getting all the metrics",
					v.MType, v.ID)
			}
		}
		slices.Sort(list)
		r := fmt.Sprintf("<html><head><title>all metrics</title></head>"+
			"<body><h1>List of all metrics</h1><p>%v</p></body></html>\n", strings.Join(list, "</p><p>"))
		writer.Header().Set("Content-type", "text/html")
		writer.Write([]byte(r))
		writer.WriteHeader(http.StatusOK)
	}
}

func PlainGetMetricHandler(storage storage.Storager) http.HandlerFunc {
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
		if metric.MType != metricType {
			http.Error(res, "Error receiving metric: metric not found", http.StatusNotFound)
			return
		}
		var responseText string
		var responseCode int
		if metric.MType == "counter" {
			responseText = fmt.Sprintf("%v\n", *metric.Delta)
			responseCode = http.StatusOK
		} else if metric.MType == "gauge" {
			responseText = fmt.Sprintf("%v\n", *metric.Value)
			responseCode = http.StatusOK
		} else {
			responseText = "Unknown metric type\n"
			responseCode = http.StatusBadRequest
		}
		res.Write([]byte(responseText))
		res.WriteHeader(responseCode)
	}
}

func JSONGetMetricHandler(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(res, "Header 'Content-Type: application/json' is required", http.StatusBadRequest)
			return
		}
		var receivedMetric models.Metric
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
			return
		}
		req.Body.Close()
		err = json.Unmarshal(body, &receivedMetric)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
			return
		}
		if receivedMetric.MType == "" {
			http.Error(res, "Missing metric type", http.StatusBadRequest)
			return
		}
		if receivedMetric.ID == "" {
			http.Error(res, "Missing metric name", http.StatusNotFound)
			return
		}
		metric, err := storage.GetMetric(receivedMetric.ID)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error receiving metric: %v", err), http.StatusNotFound)
			return
		}
		if metric.MType != receivedMetric.MType {
			http.Error(res, "Error receiving metric: metric not found", http.StatusNotFound)
			return
		}
		responseText, err := json.Marshal(metric)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error getting metric: %v", err), http.StatusBadRequest)
			return
		}
		res.Write(responseText)
		res.WriteHeader(http.StatusOK)
	}
}

func UpdateMetric(metric models.Metric, storage storage.Storager) (models.Metric, error) {
	newMetric := metric
	if metric.MType == "counter" {
		if currentMetric, err := storage.GetMetric(metric.ID); err == nil {
			if currentMetric.MType == "counter" {
				*newMetric.Delta = *newMetric.Delta + *currentMetric.Delta
			}
		}
	}
	if err := storage.SaveMetric(metric); err != nil {
		return models.Metric{}, err
	}
	return newMetric, nil
}

func PlainUpdaterHandler(storage storage.Storager) http.HandlerFunc {
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
		metric, err := models.NewMetric(metricName, metricType, metricValue)
		if err != nil {
			logger.Log.Errorf("Error handling metric: %v", err)
			http.Error(res, fmt.Sprintf("Error handling metric: %v", err), http.StatusBadRequest)
			return
		}
		newMetric, err := UpdateMetric(metric, storage)
		if err != nil {
			logger.Log.Errorf("Error updating metric: %v", err)
			http.Error(res, fmt.Sprintf("Error updating metric: %v", err), http.StatusInternalServerError)
			return
		}
		res.Write([]byte(fmt.Sprintf("Updated metric: %+v", newMetric)))
		res.WriteHeader(http.StatusOK)
	}
}

func JSONUpdaterHandler(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(res, "Header 'Content-Type: application/json' is required", http.StatusBadRequest)
			return
		}
		var receivedMetric models.Metric
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
			return
		}
		req.Body.Close()
		err = json.Unmarshal(body, &receivedMetric)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
			return
		}
		if receivedMetric.ID == "" {
			http.Error(res, "Field 'id' is required", http.StatusBadRequest)
			return
		}
		if receivedMetric.MType == "gauge" {
			if receivedMetric.Value == nil {
				http.Error(res, "Field 'value' is required for gauge metrics", http.StatusBadRequest)
				return
			}
		} else if receivedMetric.MType == "counter" {
			if receivedMetric.Delta == nil {
				http.Error(res, "Field 'delta' is required for counter metrics", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(res, "Unknown value of field 'type'. Should be 'gauge' or 'counter'", http.StatusBadRequest)
			return
		}

		newMetric, err := UpdateMetric(receivedMetric, storage)
		if err != nil {
			http.Error(res, fmt.Sprintf("Error updating metric: %v", err), http.StatusInternalServerError)
			return
		}
		new, err := json.Marshal(newMetric)
		if err != nil {
			logger.Log.Errorf("Error updating metric: %v", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write(new)
		res.WriteHeader(http.StatusOK)
	}
}
