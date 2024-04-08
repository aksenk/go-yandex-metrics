package metrics

import (
	"context"
	"github.com/aksenk/go-yandex-metrics/internal/converter"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/fatih/structs"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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

func GetPSUtilMetrics(ctx context.Context) ([]models.Metric, error) {
	var resultMetrics []models.Metric

	v, _ := mem.VirtualMemoryWithContext(ctx)

	c, err := cpu.PercentWithContext(ctx, 10, false)

	if err != nil {
		return nil, err
	}

	TotalMemoryMetric, err := models.NewMetric("TotalMemory", "gauge", v.Total)
	if err != nil {
		return nil, err
	}
	resultMetrics = append(resultMetrics, TotalMemoryMetric)

	FreeMemoryMetric, err := models.NewMetric("FreeMemory", "gauge", v.Free)
	if err != nil {
		return nil, err
	}
	resultMetrics = append(resultMetrics, FreeMemoryMetric)

	CPUutilization1Metric, err := models.NewMetric("CPUutilization1", "gauge", c[0])
	if err != nil {
		return nil, err
	}
	resultMetrics = append(resultMetrics, CPUutilization1Metric)

	return resultMetrics, nil
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
