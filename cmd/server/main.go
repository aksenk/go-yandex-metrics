package main

import (
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/server/server"
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/server/storage"
	"log"
)

func main() {
	listenAddr := "localhost:8080"
	listenPath := "/update/"
	storage := storage.NewStorage()
	s := models.Server{
		ListenAddr: listenAddr,
		ListenURL:  listenPath,
		Handler:    handlers.UpdateMetric(storage),
	}
	err := server.NewServer(s)
	if err != nil {
		log.Fatal(err)
	}
}
