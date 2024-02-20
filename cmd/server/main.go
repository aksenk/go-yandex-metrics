package main

import (
	"context"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/server/app"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger, err := logger.NewLogger("info")
	if err != nil {
		logrus.Fatalf("Can not initialize logger: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-signals
		logger.Infof("Received %v signal", signal)
		cancel()
	}()

	config, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("can not create app config: %v", err)
	}
	app, err := app.NewApp(config)
	if err != nil {
		logger.Fatalf("Application initialization error: %v", err)
	}
	go func() {
		err = app.Start(ctx)
		if err == http.ErrServerClosed {
			return
		}
		if err != nil {
			logger.Fatalf("Application launch error: %v", err)
		}
	}()
	<-ctx.Done()
	// TODO возможно сюда нужно передавать контекст для закрытия веб сервера, но не знаю какой
	err = app.Stop()
	if err != nil {
		logger.Fatalf("Shutdown error: %v", err)
	}
}
