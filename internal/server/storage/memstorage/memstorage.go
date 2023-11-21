package memstorage

import (
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"strconv"
	"sync"
)

var (
	errMetricType  = errors.New("incorrect metric type")
	errMetricValue = errors.New("incorrect metric value")
)

type MemStorage struct {
	Metrics map[string]models.Metric
	mu      sync.Mutex
}

type Storage interface {
	AddMetric() error
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Metrics: map[string]models.Metric{},
	}
}

func (s *MemStorage) AddMetric(m models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
