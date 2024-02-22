package memstorage

import (
	"context"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"go.uber.org/zap"
	"sync"
)

type MemStorage struct {
	Metrics map[string]models.Metric
	Logger  *zap.SugaredLogger
	mu      sync.Mutex
}

func NewMemStorage(logger *zap.SugaredLogger) *MemStorage {
	return &MemStorage{
		Metrics: map[string]models.Metric{},
		Logger:  logger,
	}
}

func (s *MemStorage) SaveMetric(ctx context.Context, m models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Metrics[m.ID] = m
	return nil
}

func (s *MemStorage) SaveBatchMetrics(ctx context.Context, metrics []models.Metric) error {
	return nil
}

func (s *MemStorage) GetMetric(ctx context.Context, name string) (*models.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if metric, ok := s.Metrics[name]; ok {
		return &metric, nil
	}
	return &models.Metric{}, storage.ErrMetricNotExist
}

func (s *MemStorage) GetAllMetrics(ctx context.Context) (map[string]models.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Metrics, nil
}

func (s *MemStorage) FlushMetrics() error {
	return nil
}

func (s *MemStorage) StartupRestore(ctx context.Context) error {
	return nil
}

func (s *MemStorage) Close() error {
	return nil
}

func (s *MemStorage) Status(ctx context.Context) error {
	return nil
}
