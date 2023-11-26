package storage

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
)

type Storage interface {
	SaveMetric(metric *models.Metric) error
	GetMetric(name string) (*models.Metric, error)
}
