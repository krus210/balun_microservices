# lib/authmw

Библиотека для JWT аутентификации в микросервисах. Предоставляет middleware для автоматической валидации JWT токенов от auth сервиса.

## Возможности

- **JWKS кеширование** - автоматическое обновление публичных ключей от auth сервиса
- **JWT валидация** - проверка подписи, issuer, audience, expiration
- **gRPC interceptors** - unary и stream interceptors для gRPC сервисов
- **HTTP middleware** - middleware для gateway
- **User ID в context** - автоматическое извлечение user_id из токена

## Использование в gRPC сервисе

```go
import (
    "github.com/sskorolev/balun_microservices/lib/authmw"
)

// 1. Создаем JWKS cache
cache, err := authmw.NewJWKSCache(authmw.JWKSCacheConfig{
    JWKSURL:       "http://auth:8082/jwks",
    RefreshPeriod: 5 * time.Minute,
})
if err != nil {
    log.Fatal(err)
}
defer cache.Stop()

// 2. Создаем validator
validator := authmw.NewValidator(authmw.ValidatorConfig{
    JWKSCache:        cache,
    ExpectedIssuer:   "balun-auth-service",
    ExpectedAudience: "users", // имя вашего сервиса
})

// 3. Добавляем interceptor к gRPC серверу
server := grpc.NewServer(
    grpc.UnaryInterceptor(
        authmw.UnaryServerInterceptor(
            validator,
            "/api.users.v1.UsersService/HealthCheck", // методы без auth
        ),
    ),
)
```

## Использование в HTTP gateway

```go
import (
    "github.com/sskorolev/balun_microservices/lib/authmw"
)

// Создаем cache и validator (как выше)

// Добавляем middleware
mux := http.NewServeMux()
mux.HandleFunc("/api/users", usersHandler)

handler := authmw.HTTPMiddleware(validator, "/health")(mux)
http.ListenAndServe(":8080", handler)
```

## Получение user_id в handler

```go
func MyHandler(ctx context.Context, req *pb.MyRequest) (*pb.MyResponse, error) {
    userID, ok := authmw.GetUserID(ctx)
    if !ok {
        return nil, status.Error(codes.Internal, "user_id not found in context")
    }

    // Используем userID...
}
```

## Конфигурация

### JWKS Cache

- `JWKSURL` - URL JWKS endpoint auth сервиса
- `RefreshPeriod` - период обновления ключей (по умолчанию 5 минут)
- `HTTPTimeout` - таймаут HTTP запросов (по умолчанию 10 секунд)

### Validator

- `JWKSCache` - экземпляр JWKS cache
- `ExpectedIssuer` - ожидаемый issuer в токене (должен совпадать с auth.issuer в auth сервисе)
- `ExpectedAudience` - имя вашего сервиса (должно быть в auth.audience в auth сервисе)

## Пропуск аутентификации

Для методов/путей, которые не требуют аутентификации (например, health checks):

**gRPC:**
```go
authmw.UnaryServerInterceptor(validator,
    "/api.users.v1.UsersService/HealthCheck",
    "/grpc.health.v1.Health/Check",
)
```

**HTTP:**
```go
authmw.HTTPMiddleware(validator, "/health", "/metrics", "/swagger/")
```
