package metrics

import (
	"context"
	"github.com/aksenk/go-yandex-metrics/internal/converter"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/fatih/structs"
	"math/rand"
	"runtime"
	"slices"
	"sync"
)

type PollCounter struct {
	value int64
	mu    sync.RWMutex
}

func (pc *PollCounter) Inc() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.value++
}

func (pc *PollCounter) Get() int64 {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.value
}

func (pc *PollCounter) Reset() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.value = 0
}

func GetCustomMetrics(counter int64) []models.Metric {
	rnd := rand.Float64()

	var resultMetrics []models.Metric

	pollCountMetric := models.Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: &counter,
	}
	resultMetrics = append(resultMetrics, pollCountMetric)

	randomValueMetric := models.Metric{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &rnd,
	}
	resultMetrics = append(resultMetrics, randomValueMetric)

	return resultMetrics
}

//func GetCustomMetrics(ctx context.Context, ticker *time.Ticker, counter *PollCounter, logger *zap.SugaredLogger) chan []models.Metric {
//	resultChan := make(chan []models.Metric, 1)
//
//	go func() {
//		defer close(resultChan)
//
//		for {
//			select {
//			case <-ctx.Done():
//				logger.Info("Stopping receiving runtime metrics")
//				return
//			case <-ticker.C:
//				counter.Inc()
//				metrics := getCustomMetrics(counter.Get())
//				select {
//				case resultChan <- metrics:
//				// если в канале уже есть данные
//				default:
//					// вычитываем их
//					<-resultChan
//					// помещаем туда новые данные
//					resultChan <- metrics
//				}
//			}
//		}
//	}()
//
//	return resultChan
//}

func GetPSUtilMetrics(ctx context.Context) {
}

func GetRuntimeMetrics(r []string) ([]models.Metric, error) {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	var resultMetrics []models.Metric

	for k, v := range structs.Map(m) {
		var t models.Metric

		if contains := slices.Contains(r, k); contains {
			float64Value, err := converter.AnyToFloat64(v)
			if err != nil {
				return nil, err
			}

			t = models.Metric{
				ID:    k,
				MType: "gauge",
				Value: &float64Value,
			}
			resultMetrics = append(resultMetrics, t)
		}
	}
	return resultMetrics, nil
}

//func GetRuntimeMetrics(ctx context.Context, ticker *time.Ticker, requiredMetrics []string, logger *zap.SugaredLogger) chan []models.Metric {
//	resultChan := make(chan []models.Metric, 1)
//
//	go func() {
//		defer close(resultChan)
//
//		for {
//			select {
//			case <-ctx.Done():
//				logger.Info("Stopping receiving runtime metrics")
//				return
//			case <-ticker.C:
//				metrics, err := getRuntimeMetrics(requiredMetrics)
//				if err != nil {
//					logger.Errorf("Can not get run metrics: %s", err)
//					continue
//				}
//				select {
//				case resultChan <- metrics:
//				// если в канале уже есть данные
//				default:
//					// вычитываем их
//					<-resultChan
//					// помещаем туда новые данные
//					resultChan <- metrics
//				}
//			}
//		}
//	}()
//
//	return resultChan
//}
