package models

import "net/http"

type Metric struct {
	Name  string
	Type  string
	Value any
}

type Server struct {
	ListenAddr string
	ListenURL  string
	Handler    http.HandlerFunc
}
