package storage

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
)

// сделал, но пока не уверен, что понял зачем он нужен
type Storage interface {
	AddMetric(m models.Metric) error
}
