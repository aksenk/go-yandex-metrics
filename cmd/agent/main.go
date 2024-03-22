package main

import (
	"context"
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/agent/app"
	"github.com/aksenk/go-yandex-metrics/internal/agent/config"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log, err := logger.NewLogger("info")
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	log, err = logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}
	client := http.Client{
		Timeout: time.Duration(cfg.ClientTimeout) * time.Second,
	}

	mainCtx, mainCancelCtx := context.WithCancel(context.Background())
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		signal := <-exitSignal
		log.Infof("Received signal %v", signal)
		log.Info("Starting gracefully shutdown")

		shutdownCtx, shutdownCancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancelCtx()

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				log.Fatal("Gracefull shutdown timeout exceeded, force shutdown")
			}
		}()

		mainCancelCtx()
	}()

	agent, err := app.NewApp(&client, log, cfg)
	if err != nil {
		log.Fatalf("Error creating agent: %v", err)
	}

	agent.Run(mainCtx)

	<-mainCtx.Done()
	log.Info("Shutdown completed")
}
