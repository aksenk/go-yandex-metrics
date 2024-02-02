package metrics

import (
	"github.com/aksenk/go-yandex-metrics/internal/converter"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/fatih/structs"
	"math/rand"
	"runtime"
	"slices"
	"time"
)

func getSystemMetrics() map[string]interface{} {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	// возвращаем преобразованный *Mem.Stats в map
	return structs.Map(m)
}

func getRequiredSystemMetrics(m map[string]interface{}, r []string) []models.Metric {
	log := logger.Log
	var resultMetrics []models.Metric
	for k, v := range m {
		var t models.Metric
		if contains := slices.Contains(r, k); contains {
			float64Value, err := converter.AnyToFloat64(v)
			if err != nil {
				log.Errorf("error: %s", err)
				continue
			}
			t = models.Metric{
				ID:    k,
				MType: "gauge",
				Value: &float64Value,
			}
			resultMetrics = append(resultMetrics, t)
		}
	}
	return resultMetrics
}

func generateCustomMetrics(p *models.Metric, r *models.Metric, c *int64) {
	*c += int64(1)
	rnd := rand.Float64()
	*p = models.Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: c,
	}
	*r = models.Metric{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &rnd,
	}
}

func GetMetrics(c chan []models.Metric, s time.Duration, runtimeRequiredMetrics []string) {
	pollCounter := int64(0)
	var pollCountMetric, randomValueMetric models.Metric
	for {
		systemMetrics := getSystemMetrics()
		resultMetrics := getRequiredSystemMetrics(systemMetrics, runtimeRequiredMetrics)

		generateCustomMetrics(&pollCountMetric, &randomValueMetric, &pollCounter)
		resultMetrics = append(resultMetrics, pollCountMetric, randomValueMetric)
		select {
		// если канал пуст - помещаем туда данные
		case c <- resultMetrics:
		// если в канале уже есть данные
		default:
			// вычитываем их
			<-c
			// помещаем туда новые данные
			c <- resultMetrics
		}
		time.Sleep(s)
	}
}
