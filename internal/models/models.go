package models

import (
	"errors"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/converter"
	"strconv"
)

var errIncorrectType = errors.New("incorrect metric type")
var errIncorrectValue = errors.New("incorrect metric value")

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

func NewMetric(name, mtype string, value any) (Metric, error) {
	//var flValue float64
	//var intValue int64
	//var err error
	if mtype == "gauge" {
		flValue, err := converter.AnyToFloat64(value)
		if err != nil {
			return Metric{}, err
		}
		//switch ty := value.(type) {
		//case uint32:
		//	flValue, err = strconv.ParseFloat(strconv.Itoa(int(ty)), 64)
		//	if err != nil {
		//		return Metric{}, errIncorrectValue
		//	}
		//case uint64:
		//	flValue, err = strconv.ParseFloat(strconv.FormatUint(ty, 10), 64)
		//	if err != nil {
		//		return Metric{}, errIncorrectValue
		//	}
		//case float64:
		//	flValue = value.(float64)
		//case string:
		//	flValue, err = strconv.ParseFloat(value.(string), 64)
		//	if err != nil {
		//		return Metric{}, errIncorrectValue
		//	}
		//default:
		//	return Metric{}, errIncorrectValue
		//}
		return Metric{ID: name, MType: mtype, Value: &flValue}, nil
	} else if mtype == "counter" {
		intValue, err := converter.AnyToInt64(value)
		if err != nil {
			return Metric{}, err
		}
		//switch v := value.(type) {
		//case int:
		//	intValue = int64(v)
		//case int8:
		//	intValue = int64(v)
		//case int16:
		//	intValue = int64(v)
		//case int32:
		//	intValue = int64(v)
		//case int64:
		//	intValue = v
		//case uint:
		//	intValue = int64(v)
		//case uint8:
		//	intValue = int64(v)
		//case uint16:
		//	intValue = int64(v)
		//case uint32:
		//	intValue = int64(v)
		//case uint64:
		//	// Handle uint64 separately to avoid potential overflow
		//	if v <= uint64(int64(^uint64(0)>>1)) {
		//		intValue = int64(v)
		//	} else {
		//		return Metric{}, fmt.Errorf("value is too large to convert to int64")
		//	}
		//default:
		//	return Metric{}, errIncorrectValue
		//}
		return Metric{ID: name, MType: mtype, Delta: &intValue}, nil
	}
	return Metric{}, errIncorrectType
}

//type Server struct {
//	ListenAddr string
//	ListenURL  string
//	Handler    http.HandlerFunc
//}
