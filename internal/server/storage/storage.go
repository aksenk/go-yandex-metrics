package storage

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
)

//type Saver interface {
//	SaveMetric(metric *models.Metric) error
//}
//
//type Getter interface {
//	GetMetric(name string) (*models.Metric, error)
//}

type Storager interface {
	SaveMetric(metric *models.Metric) error
	GetMetric(name string) (*models.Metric, error)
	// TODO вроде правильно делать так? но не работает
	//Saver
	//Getter
}
