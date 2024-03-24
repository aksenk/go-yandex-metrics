package metrics

import (
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
	mu    sync.Mutex
}

func (pc *PollCounter) Inc() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.value++
}

func (pc *PollCounter) Get() int64 {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.value
}

func (pc *PollCounter) Reset() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.value = 0
}

func GetSystemMetrics() map[string]interface{} {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	// возвращаем преобразованный *Mem.Stats в map
	return structs.Map(m)
}

func RemoveUnnecessaryMetrics(m map[string]interface{}, r []string) ([]models.Metric, error) {
	var resultMetrics []models.Metric
	for k, v := range m {
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

func GenerateCustomMetrics(counter int64) (models.Metric, models.Metric) {
	rnd := rand.Float64()
	pollCountMetric := models.Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: &counter,
	}
	randomValueMetric := models.Metric{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &rnd,
	}
	return pollCountMetric, randomValueMetric
}
