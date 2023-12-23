package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
	"slices"
	"strconv"
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
		r.Post("/", JSONUpdaterHandler(s))
		r.Post("/{type}/", ParamsUpdaterHandler(s))
		r.Post("/{type}/{name}/", ParamsUpdaterHandler(s))
		r.Post("/{type}/{name}/{value}", ParamsUpdaterHandler(s))
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

func UpdateMetric(metricType, metricName, metricValue string, storage storage.Storager) (any, error) {
	newMetric, err := models.NewMetric(metricName, metricType, metricValue)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error сreating new metric: %v", err))
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
		return "", err
	}
	if metricType == "counter" {
		return newCounterValue, nil
	}
	return metricValue, nil
}

func ParamsUpdaterHandler(storage storage.Storager) http.HandlerFunc {
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
		UpdateMetric(metricType, metricName, metricValue, storage)
	}
}

func JSONUpdaterHandler(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(res, "Header 'Content-Type: application/json' is required", http.StatusBadRequest)
			return
		}
		var receivedMetric models.Metrics
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
		//responseMetric := receivedMetric

		if receivedMetric.MType == "gauge" {
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!
			// TODO поставил сюда delta тк float64 не хочет конвертироваться в строку, а интернета нету чтобы разобраться
			//UpdateMetric(receivedMetric.MType, receivedMetric.ID, string(*receivedMetric.Delta), storage)
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!
			// TODO !!!!!!!!!!

		} else if receivedMetric.MType == "counter" {
			// TODO загуглить нормальный вариант конвертации int64 в string
			kkk := strconv.Itoa(int(*receivedMetric.Delta))
			//kkk := string(receivedMetric.Delta)
			logger.Log.Info(kkk)
			//newValue, err := UpdateMetric(receivedMetric.MType, receivedMetric.ID, kkk, storage)
			if err != nil {
				http.Error(res, fmt.Sprintf("Error updating metric: %v", err), http.StatusBadRequest)
				return
			}
		}
		res.Write([]byte(fmt.Sprintf("Metric saved successfully\n")))
		res.Write([]byte(fmt.Sprintf("%+v\n", storage)))
		res.WriteHeader(http.StatusOK)
	}
}
