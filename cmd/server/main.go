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
	// хз зачем делать объявление переменной тут,
	// но если в след строчке написать напрямую storage.Storage.NewMemStorage() - это не работает
	s := memstorage.NewMemStorage()
	srv := models.Server{
		ListenAddr: listenAddr,
		ListenURL:  listenPath,
		Handler:    handlers.UpdateMetric(s),
	}
	err := server.NewServer(srv)
	if err != nil {
		log.Fatal(err)
	}
}
