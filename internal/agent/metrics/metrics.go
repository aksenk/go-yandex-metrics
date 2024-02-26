package metrics

import (
	"github.com/aksenk/go-yandex-metrics/internal/converter"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/fatih/structs"
	"math/rand"
	"runtime"
	"slices"
)

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

func GenerateCustomMetrics(p *models.Metric, r *models.Metric, c *int64) {
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
