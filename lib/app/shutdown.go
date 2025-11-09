package app

import (
	"context"
	"time"
)

const (
	// GracefulShutdownTimeout - таймаут для graceful shutdown всех компонентов
	GracefulShutdownTimeout = 30 * time.Second
)

// WaitForShutdown ожидает завершения fn. После отмены ctx начинает отсчет timeout
// и возвращает context.DeadlineExceeded, если fn не завершился вовремя.
func WaitForShutdown(ctx context.Context, fn func() error, timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-done:
		return err
	case <-timer.C:
		return context.DeadlineExceeded
	}
}
