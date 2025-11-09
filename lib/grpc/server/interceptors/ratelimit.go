package interceptors

import (
	"context"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sskorolev/balun_microservices/lib/config"
)

// RateLimitUnaryInterceptor ограничивает количество запросов к gRPC серверу
// Использует глобальный rate limiter для всего сервиса (все методы делят один лимит)
func RateLimitUnaryInterceptor(cfg config.RateLimitConfig) grpc.UnaryServerInterceptor {
	// Создаем глобальный rate limiter
	// Burst устанавливаем равным reqPerSec для обработки коротких всплесков
	limiter := rate.NewLimiter(rate.Limit(cfg.ReqPerSec), cfg.ReqPerSec)

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Если rate limit disabled - пропускаем
		if !cfg.Enabled {
			return handler(ctx, req)
		}

		// Проверяем ignore list
		if isInRateLimitIgnoreList(info.FullMethod, cfg.Ignore) {
			return handler(ctx, req)
		}

		// Проверяем лимит (глобальный для всех методов)
		if !limiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		// Выполняем handler
		return handler(ctx, req)
	}
}

// isInRateLimitIgnoreList проверяет находится ли метод в списке игнорируемых
func isInRateLimitIgnoreList(method string, ignoreList []string) bool {
	for _, ignored := range ignoreList {
		if ignored == method {
			return true
		}
	}
	return false
}
