package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/sskorolev/balun_microservices/lib/config"
)

// ServeHTTP запускает HTTP сервер на указанном порту с graceful shutdown
func ServeHTTP(ctx context.Context, handler http.Handler, httpCfg config.HTTPConfig) error {
	listenAddr := fmt.Sprintf(":%d", httpCfg.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	server := &http.Server{Handler: handler}

	errChan := make(chan error, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("http server failed: %w", err)
		}
	}()

	// Ждем сигнала отмены или ошибки
	select {
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown failed: %w", err)
		}
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
