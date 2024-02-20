package storage

import (
	"context"
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/models"
)

var ErrMetricNotExist = errors.New("metric not found")

// TODO добавить везде контекст
type Storager interface {
	SaveMetric(metric models.Metric) error
	// TODO нужно ли передавать еще тип метрики?
	GetMetric(name string) (*models.Metric, error)
	GetAllMetrics() (map[string]models.Metric, error)
	StartupRestore() error
	FlushMetrics() error
	Close() error
	Status(ctx context.Context) error
}
