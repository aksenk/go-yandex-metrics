package storage

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
)

type Storager interface {
	SaveMetric(metric models.Metric) error
	GetMetric(name string) (*models.Metric, error)
	GetAllMetrics() map[string]models.Metric
	StartupRestore() error
	FlushMetrics() error
}
