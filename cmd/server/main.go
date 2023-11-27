package main

import (
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	"net/http"
)

// TODO как запускать тесты сразу по всем директориям?
func main() {
	listenAddr := "localhost:8080"
	s := memstorage.NewMemStorage()
	r := handlers.NewRouter(s)
	http.ListenAndServe(listenAddr, r)
}
