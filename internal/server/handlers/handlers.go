package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/compress"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/aksenk/go-yandex-metrics/internal/signature"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"io"
	"net/http"
	"slices"
	"strings"
)

func NewRouter(s storage.Storager, log *zap.SugaredLogger, cryptKey string) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(logger.Middleware(log))
	//r.Use(middleware.Timeout(time.Second * 10))
	r.Use(compress.Middleware)
	if cryptKey != "" {
		r.Use(signature.Middleware(cryptKey, log))
	}
	r.Get("/", ListAllMetrics(s))
	r.Get("/ping", Ping(s))
	// TODO почему-то в ответе дублируется текст "Allow: POST" например при запросе GET /update/
	r.Route("/value", func(r chi.Router) {
		r.Post("/", JSONGetMetricHandler(s))
		r.Get("/", PlainGetMetricHandler(s))
		r.Get("/{type}/", PlainGetMetricHandler(s))
		r.Get("/{type}/{name}", PlainGetMetricHandler(s))
	})
	r.Post("/updates/", JSONBatchUpdaterHandler(s))
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

func Ping(storage storage.Storager) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		err := storage.Status(ctx)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func ListAllMetrics(storage storage.Storager) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var list []string

		ctx := request.Context()

		log, err := logger.FromContext(ctx)
		if err != nil {
			http.Error(writer, "internal logger error", http.StatusInternalServerError)
			return
		}

		allMetrics, err := storage.GetAllMetrics(ctx)
		if err != nil {
			log.Errorf("Error receiving metrics: %v", err)
			http.Error(writer, fmt.Sprintf("Error receiving metrics: %v", err), http.StatusInternalServerError)
			return
		}

		for _, v := range allMetrics {
			if v.MType == "gauge" {
				list = append(list, fmt.Sprintf("%v=%v", v.ID, *v.Value))

			} else if v.MType == "counter" {
				list = append(list, fmt.Sprintf("%v=%v", v.ID, *v.Delta))
			} else {
				log.Errorf("Unknown metric type '%v' for metric '%v' when getting all the metrics",
					v.MType, v.ID)
				http.Error(writer, "Unknown metric type", http.StatusInternalServerError)
				return
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
		ctx := req.Context()

		log, err := logger.FromContext(ctx)
		if err != nil {
			http.Error(res, "internal logger error", http.StatusInternalServerError)
			return
		}

		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		if metricType == "" {
			log.Errorf("Missing metric type")
			http.Error(res, "Missing metric type", http.StatusBadRequest)
			return
		}
		if metricName == "" {
			log.Errorf("Missing metric name")
			http.Error(res, "Missing metric name", http.StatusNotFound)
			return
		}

		metric, err := storage.GetMetric(ctx, metricName)
		if err != nil {
			log.Errorf("Error receiving metric: %v", err)
			http.Error(res, fmt.Sprintf("Error receiving metric: %v", err), http.StatusNotFound)
			return
		}
		if metric.MType != metricType {
			log.Errorf("Error receiving metric: metric not found")
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
		ctx := req.Context()

		log, err := logger.FromContext(ctx)
		if err != nil {
			http.Error(res, "internal logger error", http.StatusInternalServerError)
			return
		}

		if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
			log.Errorf("Header 'Content-Type: application/json' is required")
			http.Error(res, "Header 'Content-Type: application/json' is required", http.StatusBadRequest)
			return
		}

		var receivedMetric models.Metric

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Errorf("Error reading body: %v", err)
			http.Error(res, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
			return
		}
		req.Body.Close()

		err = json.Unmarshal(body, &receivedMetric)
		if err != nil {
			log.Errorf("Error parsing JSON: %v", err)
			http.Error(res, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
			return
		}

		if receivedMetric.MType == "" {
			log.Errorf("Missing metric type")
			http.Error(res, "Missing metric type", http.StatusBadRequest)
			return
		}

		if receivedMetric.ID == "" {
			log.Errorf("Missing metric name")
			http.Error(res, "Missing metric name", http.StatusNotFound)
			return
		}

		metric, err := storage.GetMetric(ctx, receivedMetric.ID)
		if err != nil {
			log.Errorf("Error receiving metric: %v", err)
			http.Error(res, fmt.Sprintf("Error receiving metric: %v", err), http.StatusNotFound)
			return
		}

		if metric.MType != receivedMetric.MType {
			log.Errorf("Error receiving metric: metric not found")
			http.Error(res, "Error receiving metric: metric not found", http.StatusNotFound)
			return
		}

		responseText, err := json.Marshal(metric)
		if err != nil {
			log.Errorf("Error getting metric: %v", err)
			http.Error(res, fmt.Sprintf("Error getting metric: %v", err), http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write(responseText)
		res.WriteHeader(http.StatusOK)
	}
}

func CalculateCounter(ctx context.Context, metric models.Metric, s storage.Storager) (models.Metric, error) {
	newMetric := metric
	currentMetric, err := s.GetMetric(ctx, metric.ID)
	if err == nil || errors.Is(err, storage.ErrMetricNotExist) {
		if currentMetric.MType == "counter" {
			*newMetric.Delta = *newMetric.Delta + *currentMetric.Delta
		}
		return newMetric, nil
	}
	return newMetric, err
}

func UpdateMetric(ctx context.Context, metric models.Metric, storage storage.Storager) (models.Metric, error) {
	newMetric := metric
	var err error
	if metric.MType == "counter" {
		newMetric, err = CalculateCounter(ctx, metric, storage)
		if err != nil {
			return newMetric, err
		}
	}
	return newMetric, storage.SaveMetric(ctx, metric)
}

func UpdateBatchMetrics(ctx context.Context, metrics []models.Metric, storage storage.Storager) ([]models.Metric, error) {
	var newMetrics []models.Metric
	var err error
OuterLoop:
	for _, metric := range metrics {
		isExist := false
		newMetric := metric
		for _, m := range newMetrics {
			// если метрика уже встречалась в батче
			if m.ID == newMetric.ID {
				isExist = true
				if m.MType == "counter" {
					// если это counter, то суммируем с предыдущим значением, которое было в метрике из этого же батча
					*m.Delta += *newMetric.Delta
				} else {
					// если это gauge, то заменяем его значение на новое из этого же батча
					*m.Value = *newMetric.Value
				}
				continue OuterLoop
			}
		}
		// если метрика еще не встречалась в батче, то рассчитываем ее значение (для counter) и сохраняем в список
		if !isExist {
			newMetric, err = CalculateCounter(ctx, metric, storage)
			if err != nil {
				return nil, fmt.Errorf("error calculating counter: %w", err)
			}
			newMetrics = append(newMetrics, newMetric)
		}
	}
	err = storage.SaveBatchMetrics(ctx, newMetrics)
	if err != nil {
		return nil, fmt.Errorf("error saving metrics: %w", err)
	}
	return newMetrics, nil
}

func PlainUpdaterHandler(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		log, err := logger.FromContext(ctx)
		if err != nil {
			http.Error(res, "internal logger error", http.StatusInternalServerError)
			return
		}

		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		metricValue := chi.URLParam(req, "value")

		if metricType == "" {
			log.Errorf("Missing metric type")
			http.Error(res, "Missing metric type", http.StatusBadRequest)
			return
		}

		if metricName == "" {
			log.Errorf("Missing metric name")
			http.Error(res, "Missing metric name", http.StatusNotFound)
			return
		}

		if metricValue == "" {
			log.Errorf("Missing metric value")
			http.Error(res, "Missing metric value", http.StatusBadRequest)
			return
		}

		metric, err := models.NewMetric(metricName, metricType, metricValue)
		if err != nil {
			log.Errorf("Error handling metric: %v", err)
			http.Error(res, fmt.Sprintf("Error handling metric: %v", err), http.StatusBadRequest)
			return
		}

		newMetric, err := UpdateMetric(ctx, metric, storage)
		if err != nil {
			log.Errorf("Error updating metric: %v", err)
			http.Error(res, fmt.Sprintf("Error updating metric: %v", err), http.StatusInternalServerError)
			return
		}

		res.Write([]byte(fmt.Sprintf("Updated metric: %+v", newMetric)))
		res.WriteHeader(http.StatusOK)
	}
}

func JSONUpdaterHandler(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		log, err := logger.FromContext(ctx)
		if err != nil {
			http.Error(res, "internal logger error", http.StatusInternalServerError)
			return
		}

		if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
			log.Errorf("Header 'Content-Type: application/json' is required")
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
			log.Errorf("Error parsing JSON: %v", err)
			http.Error(res, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
			return
		}

		if receivedMetric.ID == "" {
			log.Errorf("Field 'id' is required")
			http.Error(res, "Field 'id' is required", http.StatusBadRequest)
			return
		}

		if receivedMetric.MType == "gauge" {
			if receivedMetric.Value == nil {
				log.Errorf("Field 'value' is required for gauge metrics")
				http.Error(res, "Field 'value' is required for gauge metrics", http.StatusBadRequest)
				return
			}
		} else if receivedMetric.MType == "counter" {
			if receivedMetric.Delta == nil {
				log.Errorf("Field 'delta' is required for counter metrics")
				http.Error(res, "Field 'delta' is required for counter metrics", http.StatusBadRequest)
				return
			}
		} else {
			log.Errorf("Unknown value of field 'type'. Should be 'gauge' or 'counter'")
			http.Error(res, "Unknown value of field 'type'. Should be 'gauge' or 'counter'", http.StatusBadRequest)
			return
		}

		newMetric, err := UpdateMetric(ctx, receivedMetric, storage)
		if err != nil {
			log.Errorf("Error updating metric: %v", err)
			http.Error(res, fmt.Sprintf("Error updating metric: %v", err), http.StatusInternalServerError)
			return
		}

		newJSONMetric, err := json.Marshal(newMetric)
		if err != nil {
			log.Errorf("Error updating metric: %v", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write(newJSONMetric)
		res.WriteHeader(http.StatusOK)
	}
}

func checkMetricIsCorrect(metric models.Metric) error {
	if metric.ID == "" {
		return fmt.Errorf("field 'id' is required")
	}
	if metric.MType == "counter" {
		if metric.Value != nil {
			return fmt.Errorf("value field is not allowed for counter metrics")
		}
	} else if metric.MType == "gauge" {
		if metric.Delta != nil {
			return fmt.Errorf("delta field is not allowed for gauge metrics")
		}
	} else {
		return fmt.Errorf("unknown metric type")
	}
	return nil
}

func JSONBatchUpdaterHandler(storage storage.Storager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		log, err := logger.FromContext(ctx)
		if err != nil {
			http.Error(res, "internal logger error", http.StatusInternalServerError)
			return
		}

		if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
			log.Errorf("Received request with incorrect header 'Content-Type: %v'", contentType)
			http.Error(res, "Header 'Content-Type: application/json' is required", http.StatusBadRequest)
			return
		}

		var receivedMetric []models.Metric
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Errorf("Error reading body: %v", err)
			http.Error(res, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
			return
		}
		req.Body.Close()

		if err = json.Unmarshal(body, &receivedMetric); err != nil {
			log.Errorf("Error parsing JSON: %v", err)
			http.Error(res, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
			return
		}

		for _, m := range receivedMetric {
			if err = checkMetricIsCorrect(m); err != nil {
				log.Errorf("Metric '%v' is incorrect: %v", m.ID, err)
				http.Error(res, fmt.Sprintf("Metric '%v' is incorrect: %v", m.ID, err), http.StatusBadRequest)
				return
			}
		}

		newMetrics, err := UpdateBatchMetrics(ctx, receivedMetric, storage)
		if err != nil {
			log.Errorf("Error updating metric: %v", err)
			http.Error(res, fmt.Sprintf("Error updating metric: %v", err), http.StatusInternalServerError)
			return
		}

		newMetricsJSON, err := json.Marshal(newMetrics)
		if err != nil {
			log.Errorf("Error updating metric: %v", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(newMetricsJSON)
	}
}
