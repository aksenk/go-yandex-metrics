package main

import (
	"context"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/app"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := logger.Log
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-signals
		log.Infof("Received %v signal", signal)
		cancel()
	}()

	config, err := config.GetConfig()
	if err != nil {
		log.Fatalf("can not create app config: %v", err)
	}
	app, err := app.NewApp(config)
	if err != nil {
		log.Fatalf("Application initialization error: %v", err)
	}
	go func() {
		err = app.Start()
		if err != nil {
			log.Fatalf("Application launch error: %v", err)
		}
	}()
	<-ctx.Done()
	err = app.Stop()
	if err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}
