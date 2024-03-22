package main

import (
	"context"
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/app"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger, err := logger.NewLogger("info")
	if err != nil {
		logrus.Fatalf("Can not initialize logger: %v", err)
	}

	mainCtx, mainStopCtx := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	config, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("can not create app config: %v", err)
	}

	app, err := app.NewApp(config)
	if err != nil {
		logger.Fatalf("Application initialization error: %v", err)
	}

	go func() {
		signal := <-signals
		logger.Infof("Received %v signal", signal)
		logger.Info("Starting graceful shutdown")

		// контекст для graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(mainCtx, 10*time.Second)
		defer cancel()

		// если таймаут истек, то завершаем приложение с ошибкой
		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				logger.Fatal("Graceful shutdown timeout. Forcing exit")
			}
		}()

		// специально переобъявляем эту переменную, что избежать возможного data race с такой же переменной с основной функции
		if err := app.Stop(shutdownCtx); err != nil {
			logger.Fatalf("Shutdown error: %v", err)
		}

		mainStopCtx()
	}()

	if err = app.Start(mainCtx); err != nil {
		logger.Fatalf("Application launch error: %v", err)
	}
	<-mainCtx.Done()
}
