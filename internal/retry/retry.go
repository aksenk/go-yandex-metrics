package retry

import (
	"fmt"
	"go.uber.org/zap"
	"time"
)

type Retry struct {
	Logger        *zap.SugaredLogger
	RetryAttempts int
	SleepStep     int
	F             func() error
}

func NewRetry(logger *zap.SugaredLogger, attempts int, sleepStep int, f func() error) *Retry {
	return &Retry{
		Logger:        logger,
		RetryAttempts: attempts,
		SleepStep:     sleepStep,
		F:             f,
	}
}

func (w *Retry) Do() error {
	err := w.F()
	if err == nil {
		return nil
	}
	if w.RetryAttempts == 0 {
		return err
	}
	for i := 1; i <= w.RetryAttempts; i++ {
		sleepTime := i*w.SleepStep - 1
		w.Logger.Errorf("Retrying in %d seconds", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
		err = w.F()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("retry quota exceeded (last origin error: %w)", err)
}
