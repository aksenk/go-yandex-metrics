package main

import (
	_ "encoding/json"
	"flag"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/handlers"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	_ "github.com/mailru/easyjson"
	"net/http"
	"os"
)

// TODO просто вопрос: как запускать тесты сразу по всем директориям?
func main() {
	log := logger.Log
	listenAddr := flag.String("a", "localhost:8080", "host:port for server listening")
	flag.Parse()
	if e := os.Getenv("ADDRESS"); e != "" {
		listenAddr = &e
	}
	s := memstorage.NewMemStorage()
	r := handlers.NewRouter(s)
	log.Infof("Starting web server on %v", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
