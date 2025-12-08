package authmw

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// AuthorizationHeader - header для JWT токена
	AuthorizationHeader = "authorization"
	// BearerPrefix - префикс для Bearer токена
	BearerPrefix = "Bearer "
	// UserIDKey - ключ для user_id в context
	UserIDKey = "user_id"
)

// UnaryServerInterceptor создает gRPC unary interceptor для JWT валидации
func UnaryServerInterceptor(validator *Validator, skipMethods ...string) grpc.UnaryServerInterceptor {
	skipMap := make(map[string]bool)
	for _, method := range skipMethods {
		skipMap[method] = true
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Пропускаем методы, которые не требуют аутентификации
		if skipMap[info.FullMethod] {
			return handler(ctx, req)
		}

		// Извлекаем метаданные
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Извлекаем Authorization header
		authHeaders := md.Get(AuthorizationHeader)
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Извлекаем токен из header
		authHeader := authHeaders[0]
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		// Валидируем токен
		claims, err := validator.Validate(ctx, tokenString)
		if err != nil {
			if err == ErrTokenExpired {
				return nil, status.Error(codes.Unauthenticated, "token expired")
			}
			if err == ErrInvalidAudience {
				return nil, status.Error(codes.PermissionDenied, "invalid audience")
			}
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Добавляем user_id в context
		ctx = context.WithValue(ctx, UserIDKey, claims.Subject)

		// Вызываем handler с обогащенным context
		return handler(ctx, req)
	}
}

// StreamServerInterceptor создает gRPC stream interceptor для JWT валидации
func StreamServerInterceptor(validator *Validator, skipMethods ...string) grpc.StreamServerInterceptor {
	skipMap := make(map[string]bool)
	for _, method := range skipMethods {
		skipMap[method] = true
	}

	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Пропускаем методы, которые не требуют аутентификации
		if skipMap[info.FullMethod] {
			return handler(srv, ss)
		}

		ctx := ss.Context()

		// Извлекаем метаданные
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Извлекаем Authorization header
		authHeaders := md.Get(AuthorizationHeader)
		if len(authHeaders) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Извлекаем токен из header
		authHeader := authHeaders[0]
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			return status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		// Валидируем токен
		claims, err := validator.Validate(ctx, tokenString)
		if err != nil {
			if err == ErrTokenExpired {
				return status.Error(codes.Unauthenticated, "token expired")
			}
			if err == ErrInvalidAudience {
				return status.Error(codes.PermissionDenied, "invalid audience")
			}
			return status.Error(codes.Unauthenticated, "invalid token")
		}

		// Обогащаем context
		ctx = context.WithValue(ctx, UserIDKey, claims.Subject)

		// Оборачиваем stream с новым context
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}
}

// wrappedServerStream оборачивает grpc.ServerStream с кастомным context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetUserID извлекает user_id из context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
