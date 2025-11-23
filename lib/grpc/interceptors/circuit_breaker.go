package interceptors

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mercari/go-circuitbreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// globalCircuitBreakers хранит circuit breaker для каждого target (host:port)
var globalCircuitBreakers = &circuitBreakerRegistry{
	breakers: make(map[string]*circuitbreaker.CircuitBreaker),
}

// circuitBreakerRegistry управляет circuit breakers для разных targets
type circuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*circuitbreaker.CircuitBreaker
}

// getOrCreate возвращает существующий circuit breaker для target или создает новый
func (r *circuitBreakerRegistry) getOrCreate(
	target string,
	failureThreshold int64,
	window time.Duration,
	halfOpenMaxSuccesses int64,
	openTimeout time.Duration,
) *circuitbreaker.CircuitBreaker {
	r.mu.RLock()
	cb, exists := r.breakers[target]
	r.mu.RUnlock()

	if exists {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check после получения write lock
	if cb, exists := r.breakers[target]; exists {
		return cb
	}

	// Создаем новый circuit breaker для этого target
	cb = circuitbreaker.New(
		circuitbreaker.WithFailOnContextCancel(true),
		circuitbreaker.WithFailOnContextDeadline(true),
		circuitbreaker.WithCounterResetInterval(window),
		circuitbreaker.WithTripFunc(circuitbreaker.NewTripFuncConsecutiveFailures(failureThreshold)),
		circuitbreaker.WithOpenTimeout(openTimeout),
		circuitbreaker.WithHalfOpenMaxSuccesses(halfOpenMaxSuccesses),
		circuitbreaker.WithOnStateChangeHookFn(func(from, to circuitbreaker.State) {
			log.Printf("Circuit breaker for %s changed state: %s -> %s", target, stateToString(from), stateToString(to))
		}),
	)

	r.breakers[target] = cb
	log.Printf("Created new circuit breaker for target %s (failures: %d, window: %v, half-open successes: %d, open timeout: %v)",
		target, failureThreshold, window, halfOpenMaxSuccesses, openTimeout)

	return cb
}

// CircuitBreakerUnaryInterceptor создает unary interceptor с circuit breaker логикой
// Circuit breaker отслеживает состояние на уровне target (host:port)
func CircuitBreakerUnaryInterceptor(
	target string,
	failureThreshold int,
	window time.Duration,
	halfOpenMaxSuccesses int,
	openTimeout time.Duration,
) (grpc.UnaryClientInterceptor, error) {
	if failureThreshold <= 0 {
		return nil, fmt.Errorf("failureThreshold must be positive, got %d", failureThreshold)
	}
	if window <= 0 {
		return nil, fmt.Errorf("window must be positive, got %v", window)
	}
	if halfOpenMaxSuccesses <= 0 {
		return nil, fmt.Errorf("halfOpenMaxSuccesses must be positive, got %d", halfOpenMaxSuccesses)
	}
	if openTimeout <= 0 {
		return nil, fmt.Errorf("openTimeout must be positive, got %v", openTimeout)
	}

	// Получаем или создаем circuit breaker для этого target
	cb := globalCircuitBreakers.getOrCreate(
		target,
		int64(failureThreshold),
		window,
		int64(halfOpenMaxSuccesses),
		openTimeout,
	)

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Выполняем вызов через circuit breaker
		_, err := cb.Do(ctx, func() (interface{}, error) {
			err := invoker(ctx, method, req, reply, cc, opts...)
			return nil, err
		})
		if err != nil {
			// Проверяем, не circuit breaker ли это ошибка
			if errors.Is(err, circuitbreaker.ErrOpen) {
				log.Printf("Circuit breaker is OPEN for target %s, rejecting call to %s", target, method)
				// Возвращаем gRPC ошибку UNAVAILABLE когда circuit breaker открыт
				return status.Errorf(codes.Unavailable, "circuit breaker is open for %s", target)
			}
		}

		return err
	}, nil
}

// stateToString конвертирует состояние circuit breaker в строку для логирования
func stateToString(state circuitbreaker.State) string {
	switch state {
	case circuitbreaker.StateClosed:
		return "CLOSED"
	case circuitbreaker.StateHalfOpen:
		return "HALF_OPEN"
	case circuitbreaker.StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}
