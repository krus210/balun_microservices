package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// TimeoutUnaryInterceptor создает unary interceptor который устанавливает timeout для каждого RPC вызова
func TimeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Проверяем, есть ли уже deadline в контексте
		if _, ok := ctx.Deadline(); !ok {
			// Если нет - устанавливаем наш timeout
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		// Вызываем следующий interceptor или сам RPC метод
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
