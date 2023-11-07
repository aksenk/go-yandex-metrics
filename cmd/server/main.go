package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type MetricType string

const (
	Gauge   = MetricType("gauge")
	Counter = MetricType("counter")
)

type Metric struct {
	Name  string
	Type  MetricType
	Value any
}

type MemStorage struct {
	Metrics map[string]Metric
}

var (
	errMetricType  = errors.New("incorrect metric type")
	errMetricValue = errors.New("incorrect metric value")
)

func (s MemStorage) AddMetric(m Metric) error {

	if m.Type == Gauge {
		if tmp, err := strconv.ParseFloat(m.Value.(string), 64); err == nil {
			m.Value = tmp
			s.Metrics[m.Name] = m
			return nil
		} else {
			return errMetricValue
		}
	}
	if m.Type == Counter {
		var intValueNew int64
		var intValueOld int64

		switch tmp := s.Metrics[m.Name].Value.(type) {
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
		return nil
	}
	return errMetricType
}

func updateMetric(storage *MemStorage) http.HandlerFunc {
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

		var metricType MetricType
		switch mt := splitURL[2]; mt {
		case "gauge":
			metricType = Gauge
		case "counter":
			metricType = Counter
		default:
			http.Error(res, "incorrect metric type", http.StatusBadRequest)
			return
		}
		metricName := splitURL[3]
		metricValue := splitURL[4]

		m := Metric{
			Name:  metricName,
			Type:  metricType,
			Value: metricValue,
		}

		err := storage.AddMetric(m)
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
	storage := MemStorage{
		Metrics: map[string]Metric{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", updateMetric(&storage))
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
