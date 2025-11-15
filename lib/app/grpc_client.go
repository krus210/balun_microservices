package app

import (
	"context"

	"github.com/sskorolev/balun_microservices/lib/config"
	grpcclient "github.com/sskorolev/balun_microservices/lib/grpc"
	"google.golang.org/grpc"
)

// InitGRPCClient создает gRPC клиент из конфигурации target сервиса
// Это helper функция для упрощения создания клиентов в сервисах
//
// Автоматически включает:
// - Tracing (OpenTelemetry через stats handler) - для распространения trace context
// - Timeout (если настроен)
// - Circuit Breaker (если настроен)
// - Retry (если настроен)
func InitGRPCClient(ctx context.Context, targetCfg *config.TargetServiceConfig) (*grpc.ClientConn, func(), error) {
	opts := []grpcclient.Option{
		grpcclient.WithInsecure(),
	}

	// Добавляем timeout если указан
	if targetCfg.GRPCClient != nil && targetCfg.GRPCClient.Timeout > 0 {
		opts = append(opts, grpcclient.WithTimeout(targetCfg.GRPCClient.Timeout))
	}

	// Добавляем retry если настроен
	if targetCfg.GRPCClient != nil && targetCfg.GRPCClient.Retry != nil {
		opts = append(opts, grpcclient.WithRetry(
			targetCfg.GRPCClient.Retry.MaxAttempts,
			grpcclient.RetryBackoffConfig{
				Base:   targetCfg.GRPCClient.Retry.Backoff.Base,
				Max:    targetCfg.GRPCClient.Retry.Backoff.Max,
				Jitter: targetCfg.GRPCClient.Retry.Backoff.Jitter,
			},
			targetCfg.GRPCClient.Retry.RetryableCodes,
		))
	}

	// Добавляем circuit breaker если настроен
	if targetCfg.GRPCClient != nil && targetCfg.GRPCClient.CircuitBreaker != nil {
		opts = append(opts, grpcclient.WithCircuitBreaker(grpcclient.CircuitBreakerConfig{
			FailuresForOpen:  targetCfg.GRPCClient.CircuitBreaker.FailuresForOpen,
			Window:           targetCfg.GRPCClient.CircuitBreaker.Window,
			HalfOpenMaxCalls: targetCfg.GRPCClient.CircuitBreaker.HalfOpenMaxCalls,
			OpenStateFor:     targetCfg.GRPCClient.CircuitBreaker.OpenStateFor,
		}))
	}

	return grpcclient.NewClient(ctx, targetCfg.Address(), opts...)
}
