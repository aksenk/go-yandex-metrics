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

func getSystemMetrics() map[string]interface{} {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	// возвращаем преобразованный *Mem.Stats в map
	return structs.Map(m)
}

func convertToFloat64(v interface{}) (float64, error) {
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
		log.Printf("unknown type:%v\n", ty)
		return value, fmt.Errorf("%s: %s", "unknown value type", ty)
	}
}

func getRequiredSystemMetrics(m map[string]interface{}, r []string) []models.Metric {
	var resultMetrics []models.Metric
	for k, v := range m {
		var t models.Metric
		if contains := slices.Contains(r, k); contains {
			float64Value, err := convertToFloat64(v)
			if err != nil {
				log.Printf("error: %s", err)
				continue
			}
			t = models.Metric{
				Name:  k,
				Type:  "gauge",
				Value: float64Value,
			}
			resultMetrics = append(resultMetrics, t)
		}
	}
	return resultMetrics
}

func generateCustomMetrics(p *models.Metric, r *models.Metric, c *int64) {
	*c += int64(1)
	*p = models.Metric{
		Name:  "PollCount",
		Type:  "counter",
		Value: *c,
	}
	*r = models.Metric{
		Name:  "RandomValue",
		Type:  "gauge",
		Value: rand.Float64(),
	}
}

func GetMetrics(c chan []models.Metric, s time.Duration, runtimeRequiredMetrics []string) {
	pollCounter := int64(0)
	var pollCountMetric, randomValueMetric models.Metric

	for {
		systemMetrics := getSystemMetrics()
		resultMetrics := getRequiredSystemMetrics(systemMetrics, runtimeRequiredMetrics)
		log.Printf("system: %+v", resultMetrics)

		generateCustomMetrics(&pollCountMetric, &randomValueMetric, &pollCounter)
		resultMetrics = append(resultMetrics, pollCountMetric, randomValueMetric)
		select {
		// если канал пуст - помещаем туда данные
		case c <- resultMetrics:
			//log.Printf("sent metrics to the channel")
		// если в канале уже есть данные
		default:
			//log.Printf("read")
			// вычитываем их
			<-c
			//log.Printf("sent")
			// помещаем туда новые данные
			c <- resultMetrics
			//log.Printf("exit")
		}
		time.Sleep(s)
	}
}
