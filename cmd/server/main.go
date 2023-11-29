package main

import (
	"flag"
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	"log"
	"net/http"
)

// TODO просто вопрос: как запускать тесты сразу по всем директориям?
func main() {
	listenAddr := flag.String("a", "localhost:8080", "host:port for server listening")
	flag.Parse()
	s := memstorage.NewMemStorage()
	r := handlers.NewRouter(s)
	log.Printf("Starting web server on %v", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
