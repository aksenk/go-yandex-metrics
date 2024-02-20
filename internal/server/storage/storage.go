package storage

import (
	"context"
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/models"
)

var ErrMetricNotExist = errors.New("metric not found")

type Storager interface {
	SaveMetric(ctx context.Context, metric models.Metric) error
	GetMetric(ctx context.Context, name string) (*models.Metric, error)
	GetAllMetrics(ctx context.Context) (map[string]models.Metric, error)
	StartupRestore(ctx context.Context) error
	FlushMetrics() error
	Close() error
	Status(ctx context.Context) error
}
