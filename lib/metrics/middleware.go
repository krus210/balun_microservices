package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	// Если метрики не инициализированы, возвращаем no-op interceptor
	if !initialized || ms.serverMetrics == nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}
	return ms.serverMetrics.UnaryServerInterceptor()
}

// ResponseTimeUnaryInterceptor - ...
func ResponseTimeUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (_ interface{}, err error) {
		start := time.Now()
		defer func() { ResponseTimeHistogramObserve(info.FullMethod, err, time.Since(start)) }()

		IncRequests(info.FullMethod)

		return handler(ctx, req)
	}
}
