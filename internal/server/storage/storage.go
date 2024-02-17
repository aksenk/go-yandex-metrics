package storage

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
)

// TODO добавить везде контекст
// TODO добавить метод Close() error
type Storager interface {
	SaveMetric(metric models.Metric) error
	GetMetric(name string) (*models.Metric, error)
	GetAllMetrics() map[string]models.Metric
	StartupRestore() error
	FlushMetrics() error
}
