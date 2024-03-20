package retry

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type Worker func(ctx context.Context) error

type Retry struct {
	Logger        *zap.SugaredLogger
	RetryAttempts int
	SleepStep     time.Duration
	Worker        Worker
}

func NewRetryer(logger *zap.SugaredLogger, attempts int, sleepStep time.Duration, worker Worker) *Retry {
	return &Retry{
		Logger:        logger,
		RetryAttempts: attempts,
		SleepStep:     sleepStep,
		Worker:        worker,
	}
}

func (w *Retry) Do(ctx context.Context) error {
	err := w.Worker(ctx)
	if err == nil {
		return nil
	}
	if w.RetryAttempts == 0 {
		return err
	}
	for i := 1; i <= w.RetryAttempts; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			sleepTime := time.Duration(i)*w.SleepStep - 1
			w.Logger.Errorf("Retrying in %d seconds", sleepTime)
			time.Sleep(sleepTime * time.Second)
			err = w.Worker(ctx)
			if err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("retry quota exceeded (last origin error: %w)", err)
}
