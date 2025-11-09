package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/config"
)

// TimeoutUnaryInterceptor устанавливает таймаут на обработку gRPC запроса
// Таймаут НЕ применяется для streaming RPC
func TimeoutUnaryInterceptor(cfg config.TimeoutConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Если timeout disabled - пропускаем
		if !cfg.Enabled {
			return handler(ctx, req)
		}

		// Проверяем ignore list
		if isInIgnoreList(info.FullMethod, cfg.Ignore) {
			return handler(ctx, req)
		}

		// Определяем таймаут для метода
		timeoutMs := cfg.TimeoutMs
		for _, pathCfg := range cfg.Paths {
			if pathCfg.Path == info.FullMethod {
				timeoutMs = pathCfg.TimeoutMs
				break
			}
		}

		// Если таймаут не указан - используем контекст как есть
		if timeoutMs <= 0 {
			return handler(ctx, req)
		}

		// Создаем новый контекст с таймаутом
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()

		// Выполняем handler с таймаутом
		return handler(timeoutCtx, req)
	}
}

// isInIgnoreList проверяет находится ли метод в списке игнорируемых
func isInIgnoreList(method string, ignoreList []string) bool {
	for _, ignored := range ignoreList {
		if ignored == method {
			return true
		}
	}
	return false
}
