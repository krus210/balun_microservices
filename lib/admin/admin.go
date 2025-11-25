package admin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Config содержит настройки admin HTTP сервера
type Config struct {
	Enabled bool
	Host    string
	Port    int
	Metrics MetricsConfig
	Pprof   PprofConfig
}

// MetricsConfig содержит настройки эндпоинта метрик
type MetricsConfig struct {
	Enabled bool
	Path    string
}

// PprofConfig содержит настройки pprof эндпоинтов
type PprofConfig struct {
	Enabled bool
	Path    string
}

const (
	// DefaultMetricsPath - путь по умолчанию для метрик
	DefaultMetricsPath = "/metrics"
	// DefaultPprofPath - путь по умолчанию для pprof
	DefaultPprofPath = "/debug/pprof"
	// GracefulShutdownTimeout - таймаут для graceful shutdown
	GracefulShutdownTimeout = 30 * time.Second
)

// Init инициализирует admin HTTP сервер и возвращает cleanup функцию
func Init(cfg Config) (*http.Server, func(), error) {
	// Если admin сервер отключен, возвращаем nil
	if !cfg.Enabled {
		return nil, func() {}, nil
	}

	// Валидация
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, nil, fmt.Errorf("invalid admin port: %d", cfg.Port)
	}

	// Создаем HTTP сервер с настроенными handlers
	server := NewServer(cfg)

	// Cleanup функция (вызывается при shutdown)
	cleanup := func() {
		// Здесь можно добавить дополнительную логику очистки если нужно
	}

	return server, cleanup, nil
}

// NewServer создает новый HTTP сервер с настроенными routes
func NewServer(cfg Config) *http.Server {
	mux := http.NewServeMux()

	// Регистрируем handlers
	RegisterHandlers(mux, cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// Serve запускает admin HTTP сервер с graceful shutdown
func Serve(ctx context.Context, server *http.Server) error {
	if server == nil {
		// Сервер не инициализирован (отключен в конфиге)
		<-ctx.Done()
		return ctx.Err()
	}

	lis, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", server.Addr, err)
	}

	errChan := make(chan error, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("admin server failed: %w", err)
		}
	}()

	// Ждем сигнала отмены или ошибки
	select {
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("admin server shutdown failed: %w", err)
		}
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
