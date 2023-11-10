package server

import (
	"net/http"
)

func NewServer(listenAddr string, listenPath string, handler http.HandlerFunc) error {
	mux := http.NewServeMux()
	mux.HandleFunc(listenPath, handler)
	return http.ListenAndServe(listenAddr, mux)
}
