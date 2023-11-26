package models

import (
	"errors"
	"net/http"
	"strconv"
)

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
	var valueErr error = errors.New("incorrect metric value")
	var typeErr error = errors.New("incorrect metric type")
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
			return &newMetric, valueErr
		}
	default:
		return &newMetric, typeErr
	}
	return &newMetric, nil
}
