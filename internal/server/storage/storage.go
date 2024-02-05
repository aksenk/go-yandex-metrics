package storage

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
)

type MetricSaver interface {
	SaveMetric(metric *models.Metric) error
}

type MetricGetter interface {
	GetMetric(name string) (*models.Metric, error)
}

type AllMetricsGetter interface {
	GetAllMetrics() map[string]models.Metric
}

type Storager interface {
	MetricSaver
	MetricGetter
	AllMetricsGetter
}
