package interceptors

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryUnaryInterceptor создает unary interceptor который повторяет неудачные RPC вызовы
// Retry применяется ТОЛЬКО для идемпотентных операций (методы с префиксами: Get, List, Search, Describe)
// Для не-идемпотентных операций retry отключен
func RetryUnaryInterceptor(
	maxAttempts int,
	baseBackoff time.Duration,
	maxBackoff time.Duration,
	jitter bool,
	retryableCodes []string,
) grpc.UnaryClientInterceptor {
	// Преобразуем строковые коды в codes.Code
	retryableGRPCCodes := make(map[codes.Code]struct{}, len(retryableCodes))
	for _, codeStr := range retryableCodes {
		if code, ok := parseGRPCCode(codeStr); ok {
			retryableGRPCCodes[code] = struct{}{}
		}
	}

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Проверяем идемпотентность метода
		if !isIdempotent(method) {
			// Для не-идемпотентных операций retry отключен
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		var lastErr error
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			// Выполняем вызов
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				// Успех
				if attempt > 1 {
					log.Printf("gRPC retry succeeded on attempt %d/%d for method %s", attempt, maxAttempts, method)
				}
				return nil
			}

			lastErr = err

			// Проверяем, стоит ли делать retry
			st, ok := status.FromError(err)
			if !ok {
				// Не gRPC ошибка - не делаем retry
				return err
			}

			if _, shouldRetry := retryableGRPCCodes[st.Code()]; !shouldRetry {
				// Код ошибки не подлежит retry
				return err
			}

			// Это последняя попытка?
			if attempt >= maxAttempts {
				log.Printf("gRPC retry exhausted all %d attempts for method %s: %v", maxAttempts, method, err)
				return err
			}

			// Проверяем, не истек ли контекст
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Вычисляем время ожидания с exponential backoff
			backoff := calculateBackoff(attempt, baseBackoff, maxBackoff, jitter)
			log.Printf("gRPC retry attempt %d/%d for method %s after %v (error: %v)", attempt, maxAttempts, method, backoff, st.Code())

			// Ждем перед следующей попыткой
			select {
			case <-time.After(backoff):
				// Продолжаем retry
			case <-ctx.Done():
				// Контекст отменен
				return ctx.Err()
			}
		}

		return lastErr
	}
}

// isIdempotent проверяет, является ли метод идемпотентным по имени
// Идемпотентные операции: Get*, List*, Search*, Describe*
func isIdempotent(method string) bool {
	// Извлекаем имя метода из полного пути (например, "/api.users.v1.UsersService/GetUser" -> "GetUser")
	parts := strings.Split(method, "/")
	if len(parts) == 0 {
		return false
	}
	methodName := parts[len(parts)-1]

	// Проверяем префиксы идемпотентных операций
	idempotentPrefixes := []string{"Get", "List", "Search", "Describe", "Find", "Query"}
	for _, prefix := range idempotentPrefixes {
		if strings.HasPrefix(methodName, prefix) {
			return true
		}
	}

	return false
}

// calculateBackoff вычисляет время ожидания с exponential backoff и опциональным jitter
func calculateBackoff(attempt int, base, max time.Duration, useJitter bool) time.Duration {
	// Exponential backoff: base * 2^(attempt-1)
	backoff := base
	for i := 1; i < attempt; i++ {
		backoff *= 2
		if backoff > max {
			backoff = max
			break
		}
	}

	// Добавляем jitter для избежания thundering herd problem
	if useJitter {
		// Jitter: random between [backoff/2, backoff]
		halfBackoff := backoff / 2
		jitterRange := backoff - halfBackoff
		backoff = halfBackoff + time.Duration(rand.Int63n(int64(jitterRange)))
	}

	return backoff
}

// parseGRPCCode преобразует строковое представление кода в codes.Code
func parseGRPCCode(codeStr string) (codes.Code, bool) {
	codeStr = strings.ToUpper(codeStr)
	switch codeStr {
	case "OK":
		return codes.OK, true
	case "CANCELED", "CANCELLED":
		return codes.Canceled, true
	case "UNKNOWN":
		return codes.Unknown, true
	case "INVALID_ARGUMENT":
		return codes.InvalidArgument, true
	case "DEADLINE_EXCEEDED":
		return codes.DeadlineExceeded, true
	case "NOT_FOUND":
		return codes.NotFound, true
	case "ALREADY_EXISTS":
		return codes.AlreadyExists, true
	case "PERMISSION_DENIED":
		return codes.PermissionDenied, true
	case "RESOURCE_EXHAUSTED":
		return codes.ResourceExhausted, true
	case "FAILED_PRECONDITION":
		return codes.FailedPrecondition, true
	case "ABORTED":
		return codes.Aborted, true
	case "OUT_OF_RANGE":
		return codes.OutOfRange, true
	case "UNIMPLEMENTED":
		return codes.Unimplemented, true
	case "INTERNAL":
		return codes.Internal, true
	case "UNAVAILABLE":
		return codes.Unavailable, true
	case "DATA_LOSS":
		return codes.DataLoss, true
	case "UNAUTHENTICATED":
		return codes.Unauthenticated, true
	default:
		return codes.Unknown, false
	}
}
