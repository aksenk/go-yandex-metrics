package models

import (
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/converter"
	"strconv"
)

// TODO вопрос зачем делать указатели на int64 float64 ?
type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m Metric) String() string {
	if m.Delta != nil {
		return strconv.FormatInt(*m.Delta, 10)
	} else if m.Value != nil {
		return fmt.Sprintf("%g", *m.Value)
	}
	return ""
}

type MType string

const (
	Gauge   MType = "gauge"
	Counter MType = "counter"
)

func (m MType) String() string {
	return string(m)
}

func NewMetric(name, mtype string, value any) (Metric, error) {
	if mtype == "gauge" {
		flValue, err := converter.AnyToFloat64(value)
		if err != nil {
			return Metric{}, err
		}
		return Metric{ID: name, MType: mtype, Value: &flValue}, nil
	} else if mtype == "counter" {
		intValue, err := converter.AnyToInt64(value)
		if err != nil {
			return Metric{}, err
		}
		return Metric{ID: name, MType: mtype, Delta: &intValue}, nil
	}
	return Metric{}, fmt.Errorf("incorrect metric type")
}
