package metrics

import (
	"context"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/converter"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/fatih/structs"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"runtime"
	"slices"
	"sync"
	"time"
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

func GetCustomMetrics(ctx context.Context, counter int64, pollTicker *time.Ticker) chan []models.Metric {
	resultChan := make(chan []models.Metric, 1)

	go func() {
		defer close(resultChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-pollTicker.C:
				var metrics []models.Metric

				rnd := rand.Float64()

				pollCountMetric := models.Metric{
					ID:    "PollCount",
					MType: "counter",
					Delta: &counter,
				}
				metrics = append(metrics, pollCountMetric)

				randomValueMetric := models.Metric{
					ID:    "RandomValue",
					MType: "gauge",
					Value: &rnd,
				}
				metrics = append(metrics, randomValueMetric)

				select {
				case resultChan <- metrics:
				default:
					<-resultChan
					resultChan <- metrics
				}
			}
		}
	}()

	return resultChan
}

func GetPSUtilMetrics(ctx context.Context, pollTicker *time.Ticker) chan []models.Metric {
	resultChan := make(chan []models.Metric, 1)

	go func() {
		defer close(resultChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-pollTicker.C:
				var metrics []models.Metric

				cpuCount := runtime.NumCPU()
				v, _ := mem.VirtualMemoryWithContext(ctx)
				c, _ := cpu.PercentWithContext(ctx, 10, true)

				TotalMemoryMetric, _ := models.NewMetric("TotalMemory", "gauge", v.Total)
				metrics = append(metrics, TotalMemoryMetric)

				FreeMemoryMetric, _ := models.NewMetric("FreeMemory", "gauge", v.Free)
				metrics = append(metrics, FreeMemoryMetric)

				for i := 0; i < cpuCount; i++ {
					CPUUtilizationMetric, _ := models.NewMetric(fmt.Sprintf("CPUutilization%v", i+1), "gauge", c[i])
					metrics = append(metrics, CPUUtilizationMetric)
				}

				select {
				case resultChan <- metrics:
				default:
					<-resultChan
					resultChan <- metrics
				}
			}
		}
	}()

	return resultChan
}

func GetRuntimeMetrics(ctx context.Context, r []string, pollTicker *time.Ticker) chan []models.Metric {
	resultChan := make(chan []models.Metric, 1)

	go func() {
		defer close(resultChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-pollTicker.C:
				m := &runtime.MemStats{}
				runtime.ReadMemStats(m)

				var metrics []models.Metric

				for k, v := range structs.Map(m) {
					if contains := slices.Contains(r, k); contains {
						float64Value, err := converter.AnyToFloat64(v)
						if err != nil {
							continue
						}
						t, _ := models.NewMetric(k, "gauge", float64Value)

						metrics = append(metrics, t)
					}
				}

				select {
				case resultChan <- metrics:
				default:
					<-resultChan
					resultChan <- metrics
				}
			}
		}
	}()

	return resultChan
}
