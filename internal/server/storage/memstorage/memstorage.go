package memstorage

import (
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/models"
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

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: map[string]models.Metric{},
	}
}

func (s *MemStorage) SaveMetric(m *models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Metrics[m.Name] = *m
	return nil
}

func (s *MemStorage) GetMetric(name string) (*models.Metric, error) {
	notExistErr := errors.New("metric not found")
	s.mu.Lock()
	defer s.mu.Unlock()
	if metric, ok := s.Metrics[name]; ok {
		return &metric, nil
	}
	return &models.Metric{}, notExistErr
}
