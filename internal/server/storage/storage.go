package storage

import (
	"errors"
	"strconv"
)

//type MetricType string
//
//const (
//	Gauge   = MetricType("gauge")
//	Counter = MetricType("counter")
//)

var (
	errMetricType  = errors.New("incorrect metric type")
	errMetricValue = errors.New("incorrect metric value")
)

type Metric struct {
	Name  string
	Type  string
	Value any
}

type MemStorage struct {
	Metrics map[string]Metric
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Metrics: map[string]Metric{},
	}
}

func (s MemStorage) NewMetric(metricName string, metricType string, metricValue any) *Metric {
	return &Metric{
		Name:  metricName,
		Type:  metricType,
		Value: metricValue,
	}
}

func (s MemStorage) AddMetric(m Metric) error {
	if m.Type == "gauge" {
		if tmp, err := strconv.ParseFloat(m.Value.(string), 64); err == nil {
			m.Value = tmp
			s.Metrics[m.Name] = m
			return nil
		} else {
			return errMetricValue
		}
	}
	if m.Type == "counter" {
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
