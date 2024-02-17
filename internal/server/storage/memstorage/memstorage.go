package memstorage

import (
	"context"
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"sync"
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

func (s *MemStorage) SaveMetric(m models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Metrics[m.ID] = m
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

func (s *MemStorage) GetAllMetrics() map[string]models.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Metrics
}

func (s *MemStorage) FlushMetrics() error {
	return nil
}

func (s *MemStorage) StartupRestore() error {
	return nil
}

func (s *MemStorage) Close() error {
	return nil
}

func (s *MemStorage) Status(ctx context.Context) error {
	return nil
}
