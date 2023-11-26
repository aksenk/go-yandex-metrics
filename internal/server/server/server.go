package server

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"net/http"
)

// шляпа какая-то, но не знаю как по другому сделать
// вообще если честно хз зачем это в отдельный модуль выносить
// и как тестировать в такой отстойной реализации
func NewServer(s models.Server) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.ListenURL, s.Handler)
	return http.ListenAndServe(s.ListenAddr, mux)
}
