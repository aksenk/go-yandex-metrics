package models

import (
	"errors"
	"net/http"
	"strconv"
)

// TODO вопрос зачем делать указатели на int64 float64 ?
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type Metric struct {
	Name  string
	Type  string
	Value any
}

type Server struct {
	ListenAddr string
	ListenURL  string
	Handler    http.HandlerFunc
}

func NewMetric(metricName, metricType, metricValue string) (*Metric, error) {
	valueErr := errors.New("incorrect metric value")
	typeErr := errors.New("incorrect metric type")
	var newMetric Metric

	switch metricType {
	case "gauge":
		if newValue, err := strconv.ParseFloat(metricValue, 64); err == nil {
			newMetric = Metric{
				Name:  metricName,
				Type:  metricType,
				Value: newValue,
			}
		} else {
			return &newMetric, valueErr
		}
	case "counter":
		if newValue, err := strconv.ParseInt(metricValue, 10, 64); err == nil {
			newMetric = Metric{
				Name:  metricName,
				Type:  metricType,
				Value: newValue,
			}
		} else {
			return &newMetric, err
		}
	default:
		return &newMetric, typeErr
	}
	return &newMetric, nil
}
