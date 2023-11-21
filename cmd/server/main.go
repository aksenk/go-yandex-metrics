package main

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/server"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	"log"
)

func main() {
	listenAddr := "localhost:8080"
	listenPath := "/update/"
	storage := memstorage.NewStorage()
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
