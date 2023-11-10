package server

import (
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"net/http"
)

func NewServer(s models.Server) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.ListenURL, s.Handler)
	// TODO шляпа какая-то, но пока хз как по другому сделать
	return http.ListenAndServe(s.ListenAddr, mux)
}
