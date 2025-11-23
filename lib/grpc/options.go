package grpc

import (
	"time"

	"google.golang.org/grpc"
)

// WithTimeout устанавливает таймаут для каждого RPC вызова
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = timeout
	}
}

// WithRetry включает retry логику с указанными параметрами
// Retry автоматически применяется только для идемпотентных операций (Get*, List*, Search*, Describe*)
func WithRetry(maxAttempts int, backoff RetryBackoffConfig, retryableCodes []string) Option {
	return func(c *config) {
		c.retryEnabled = true
		c.retryMaxAttempts = maxAttempts
		c.retryBackoff = backoff
		c.retryableCodes = retryableCodes
	}
}

// WithCircuitBreaker включает circuit breaker с указанными параметрами
// Circuit breaker отслеживает состояние на уровне target (host:port)
func WithCircuitBreaker(cbConfig CircuitBreakerConfig) Option {
	return func(c *config) {
		c.circuitBreakerEnabled = true
		c.circuitBreakerConfig = cbConfig
	}
}

// WithInsecure отключает TLS и использует незащищенное подключение
func WithInsecure() Option {
	return func(c *config) {
		c.insecure = true
	}
}

// WithUnaryInterceptors добавляет пользовательские unary interceptors
// Эти interceptors будут применены ПОСЛЕ встроенных (timeout, circuit breaker, retry)
func WithUnaryInterceptors(interceptors ...grpc.UnaryClientInterceptor) Option {
	return func(c *config) {
		c.unaryInterceptors = append(c.unaryInterceptors, interceptors...)
	}
}

// WithStreamInterceptors добавляет пользовательские stream interceptors
func WithStreamInterceptors(interceptors ...grpc.StreamClientInterceptor) Option {
	return func(c *config) {
		c.streamInterceptors = append(c.streamInterceptors, interceptors...)
	}
}

// WithDefaultRetry включает retry с дефолтными настройками
// maxAttempts: 3
// backoff: base=100ms, max=2s, jitter=true
// retryable codes: UNAVAILABLE, DEADLINE_EXCEEDED, RESOURCE_EXHAUSTED, ABORTED
func WithDefaultRetry() Option {
	return WithRetry(
		3,
		RetryBackoffConfig{Base: 100 * time.Millisecond, Max: 2 * time.Second, Jitter: true},
		[]string{"UNAVAILABLE", "DEADLINE_EXCEEDED", "RESOURCE_EXHAUSTED", "ABORTED"},
	)
}

// WithDefaultCircuitBreaker включает circuit breaker с дефолтными настройками
// failures: 5 за 30 секунд
// half-open: max 5 вызовов
// open state: 60 секунд
func WithDefaultCircuitBreaker() Option {
	return WithCircuitBreaker(CircuitBreakerConfig{
		FailuresForOpen:  5,
		Window:           30 * time.Second,
		HalfOpenMaxCalls: 5,
		OpenStateFor:     60 * time.Second,
	})
}
