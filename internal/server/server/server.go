package server

import (
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"time"
)

func NewRouter(s storage.Storager) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(10 * time.Second))
	// TODO вынести работу со storage в middleware?
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetric(s))
		r.Post("/{type}/", handlers.UpdateMetric(s))
		r.Post("/{type}/{name}/", handlers.UpdateMetric(s))
		r.Post("/{type}/{name}/{value}", handlers.UpdateMetric(s))
	})
	r.Get("/value/{type}/{name}", handlers.GetMetric(s))
	return r
}
