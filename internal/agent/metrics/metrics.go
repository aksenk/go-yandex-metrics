package metrics

import (
	"fmt"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"github.com/fatih/structs"
	"log"
	"math/rand"
	"runtime"
	"slices"
	"strconv"
	"time"
)

func getSystemMetricsMap() map[string]interface{} {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	// возвращаем преобразованный *Mem.Stats в map
	return structs.Map(m)
}

func convertMetricValueToFloat64(v interface{}) (float64, error) {
	var value float64
	switch ty := v.(type) {
	case uint32:
		value, _ = strconv.ParseFloat(strconv.Itoa(int(ty)), 64)
		return value, nil
	case uint64:
		value, _ = strconv.ParseFloat(strconv.FormatUint(ty, 10), 64)
		return value, nil
	case float64:
		return ty, nil
	default:
		log.Printf("unknown type\n")
		return value, fmt.Errorf("%s: %s", "unknown value type", ty)
	}
}

func getRequiredSystemMetrics(m map[string]interface{}, r []string) []models.Metric {
	var resultMetrics []models.Metric
	for k, v := range m {
		var t models.Metric
		if contains := slices.Contains(r, k); contains {
			//log.Printf("processing %s, %s", k, v)
			float64Value, err := convertMetricValueToFloat64(v)
			if err != nil {
				log.Printf("error: %s", err)
				continue
			}
			t = models.Metric{
				Name:  k,
				Type:  "gauge",
				Value: float64Value,
			}
			//log.Printf("append")
			resultMetrics = append(resultMetrics, t)
		}
		//else {
		//	log.Printf("metric %s is not required", k)
		//}
	}
	return resultMetrics
}

func generateCustomMetrics() (models.Metric, models.Metric) {
	pollCountMetric := models.Metric{
		Name:  "PollCount",
		Type:  "counter",
		Value: 1,
	}
	randomValueMetric := models.Metric{
		Name:  "RandomValue",
		Type:  "gauge",
		Value: rand.Float64(),
	}
	return pollCountMetric, randomValueMetric
}

func GetMetrics(c chan []models.Metric, s time.Duration) {
	for {
		// требуемые метрики по заданию
		runtimeRequiredMetrics := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
			"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
			"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
			"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}
		systemMetrics := getSystemMetricsMap()
		resultMetrics := getRequiredSystemMetrics(systemMetrics, runtimeRequiredMetrics)

		pollCountMetric, randomValueMetric := generateCustomMetrics()
		resultMetrics = append(resultMetrics, pollCountMetric, randomValueMetric)
		log.Printf("metrics received\n")
		select {
		// если канал пуст - помещаем туда данные
		case c <- resultMetrics:
		// если в канале уже есть данные
		default:
			<-c
			c <- resultMetrics
		}
		time.Sleep(s)
	}
}
