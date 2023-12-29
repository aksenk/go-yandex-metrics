package models

import (
	"fmt"
	"net/http"
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

//func (m Metric) StringDelta() string {
//	return strconv.FormatInt(*m.Delta, 10)
//}
//
//func (m Metric) StringValue() string {
//	return fmt.Sprintf("%g", *m.Value)
//}

type Server struct {
	ListenAddr string
	ListenURL  string
	Handler    http.HandlerFunc
}
