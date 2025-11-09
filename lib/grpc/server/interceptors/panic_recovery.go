package interceptors

import (
	"context"
	"log"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PanicRecoveryUnaryInterceptor перехватывает panic в обработчиках gRPC и возвращает Internal ошибку
// Интерсептор всегда включен для обеспечения стабильности сервиса
func PanicRecoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				// Получаем стек вызовов
				stack := debug.Stack()

				// Логируем panic с полным стеком
				log.Printf("PANIC recovered in gRPC handler: method=%s, error=%v\nStack trace:\n%s",
					info.FullMethod, r, string(stack))

				// Возвращаем Internal ошибку клиенту без раскрытия деталей
				err = status.Error(codes.Internal, "internal server error")
				resp = nil
			}
		}()

		// Выполняем handler
		return handler(ctx, req)
	}
}
