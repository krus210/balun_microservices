package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sskorolev/balun_microservices/lib/grpc/interceptors"
)

// Option определяет функциональную опцию для конфигурации клиента
type Option func(*config)

// config хранит внутреннюю конфигурацию gRPC клиента
type config struct {
	// Timeout для каждого RPC вызова
	timeout time.Duration

	// Retry конфигурация
	retryEnabled     bool
	retryMaxAttempts int
	retryBackoff     RetryBackoffConfig
	retryableCodes   []string

	// Circuit Breaker конфигурация
	circuitBreakerEnabled bool
	circuitBreakerConfig  CircuitBreakerConfig

	// TLS конфигурация
	insecure bool

	// Дополнительные interceptors
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
}

// RetryBackoffConfig конфигурирует exponential backoff для retry
type RetryBackoffConfig struct {
	Base   time.Duration
	Max    time.Duration
	Jitter bool
}

// CircuitBreakerConfig конфигурирует circuit breaker
type CircuitBreakerConfig struct {
	FailuresForOpen  int
	Window           time.Duration
	HalfOpenMaxCalls int
	OpenStateFor     time.Duration
}

// NewClient создает новое gRPC клиентское подключение с настроенными интерсепторами
// Возвращает:
// - *grpc.ClientConn - готовое подключение
// - cleanup func() - функция для graceful shutdown
// - error - ошибка при создании подключения
func NewClient(ctx context.Context, target string, opts ...Option) (*grpc.ClientConn, func(), error) {
	// Создаем конфигурацию с дефолтными значениями
	cfg := &config{
		timeout:               5 * time.Second,
		retryEnabled:          false,
		retryMaxAttempts:      3,
		retryBackoff:          RetryBackoffConfig{Base: 100 * time.Millisecond, Max: 2 * time.Second, Jitter: true},
		retryableCodes:        []string{"UNAVAILABLE", "DEADLINE_EXCEEDED", "RESOURCE_EXHAUSTED", "ABORTED"},
		circuitBreakerEnabled: false,
		circuitBreakerConfig:  CircuitBreakerConfig{FailuresForOpen: 5, Window: 30 * time.Second, HalfOpenMaxCalls: 5, OpenStateFor: 60 * time.Second},
		insecure:              true,
	}

	// Применяем все переданные опции
	for _, opt := range opts {
		opt(cfg)
	}

	// Создаем цепочку unary interceptors
	unaryInterceptors := make([]grpc.UnaryClientInterceptor, 0)

	// 1. Timeout interceptor (всегда активен)
	if cfg.timeout > 0 {
		unaryInterceptors = append(unaryInterceptors, interceptors.TimeoutUnaryInterceptor(cfg.timeout))
	}

	// 2. Circuit Breaker interceptor
	if cfg.circuitBreakerEnabled {
		cbInterceptor, err := interceptors.CircuitBreakerUnaryInterceptor(
			target,
			cfg.circuitBreakerConfig.FailuresForOpen,
			cfg.circuitBreakerConfig.Window,
			cfg.circuitBreakerConfig.HalfOpenMaxCalls,
			cfg.circuitBreakerConfig.OpenStateFor,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create circuit breaker interceptor: %w", err)
		}
		unaryInterceptors = append(unaryInterceptors, cbInterceptor)
	}

	// 3. Retry interceptor
	if cfg.retryEnabled {
		unaryInterceptors = append(unaryInterceptors, interceptors.RetryUnaryInterceptor(
			cfg.retryMaxAttempts,
			cfg.retryBackoff.Base,
			cfg.retryBackoff.Max,
			cfg.retryBackoff.Jitter,
			cfg.retryableCodes,
		))
	}

	// 4. Добавляем пользовательские interceptors
	unaryInterceptors = append(unaryInterceptors, cfg.unaryInterceptors...)

	// Создаем dial опции
	dialOpts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
	}

	// Добавляем stream interceptors если есть
	if len(cfg.streamInterceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithChainStreamInterceptor(cfg.streamInterceptors...))
	}

	// Настройка TLS
	if cfg.insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Создаем подключение
	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Cleanup функция для graceful shutdown
	cleanup := func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close gRPC connection to %s: %v", target, err)
		}
	}

	return conn, cleanup, nil
}
