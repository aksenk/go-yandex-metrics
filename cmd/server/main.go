package main

import (
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	"net/http"
)

func main() {
	listenAddr := "localhost:8080"
	listenPath := "/update/"
	s := memstorage.NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(listenPath, handlers.UpdateMetric(s))
	http.ListenAndServe(listenAddr, mux)
}
