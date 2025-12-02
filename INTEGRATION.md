# Интеграция JWT аутентификации

Данная инструкция описывает, как интегрировать lib/authmw в микросервисы.

## Обзор

Все выполнено:
- ✅ Auth Service с JWT токенами (RS256)
- ✅ Refresh token rotation с anti-reuse защитой
- ✅ JWKS endpoint в auth сервисе
- ✅ lib/authmw библиотека с middleware
- ⏳ Интеграция в сервисы (ниже пример)

## Интеграция в gRPC сервис (users, social, chat)

### 1. Добавьте зависимость в go.mod

```bash
cd users  # или social, chat
go get github.com/sskorolev/balun_microservices/lib/authmw
```

### 2. Обновите cmd/main.go

```go
import (
    "github.com/sskorolev/balun_microservices/lib/authmw"
)

func main() {
    // ... существующий код ...

    // Создаем JWKS cache
    jwksCache, err := authmw.NewJWKSCache(authmw.JWKSCacheConfig{
        JWKSURL:       "http://auth:8082/.well-known/jwks.json",  // или настроить через конфиг
        RefreshPeriod: 5 * time.Minute,
    })
    if err != nil {
        logger.FatalKV(ctx, "failed to create JWKS cache", "error", err)
    }
    defer jwksCache.Stop()

    // Создаем JWT validator
    jwtValidator := authmw.NewValidator(authmw.ValidatorConfig{
        JWKSCache:        jwksCache,
        ExpectedIssuer:   "balun-auth-service",
        ExpectedAudience: "users",  // имя сервиса из auth/config.yaml auth.audience
    })

    // Инициализируем gRPC сервер С interceptor
    application.InitGRPCServer(cfg.Server,
        authmw.UnaryServerInterceptor(jwtValidator),  // добавляем JWT interceptor
        errorsMiddleware.ErrorsUnaryInterceptor(),
    )

    // ... остальной код ...
}
```

### 3. Используйте user_id в handlers

```go
func (h *UsersController) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
    // Извлекаем user_id из context
    userID, ok := authmw.GetUserID(ctx)
    if !ok {
        return nil, status.Error(codes.Internal, "user_id not found")
    }

    // Используем userID для бизнес-логики
    profile, err := h.usecase.GetProfile(ctx, userID)
    // ...
}
```

### 4. Пропуск аутентификации для public методов

Если есть методы, которые не требуют аутентификации:

```go
authmw.UnaryServerInterceptor(jwtValidator,
    "/api.users.v1.UsersService/HealthCheck",
    "/grpc.health.v1.Health/Check",
)
```

## Интеграция в Gateway

### 1. Добавьте зависимость

```bash
cd gateway
go get github.com/sskorolev/balun_microservices/lib/authmw
```

### 2. Обновите cmd/main.go

```go
import (
    "github.com/sskorolev/balun_microservices/lib/authmw"
)

func main() {
    // ... существующий код ...

    // Создаем JWKS cache
    jwksCache, err := authmw.NewJWKSCache(authmw.JWKSCacheConfig{
        JWKSURL:       "http://auth:8082/.well-known/jwks.json",
        RefreshPeriod: 5 * time.Minute,
    })
    if err != nil {
        logger.FatalKV(ctx, "failed to create JWKS cache", "error", err)
    }
    defer jwksCache.Stop()

    // Создаем JWT validator
    jwtValidator := authmw.NewValidator(authmw.ValidatorConfig{
        JWKSCache:        jwksCache,
        ExpectedIssuer:   "balun-auth-service",
        ExpectedAudience: "gateway",
    })

    // Оборачиваем HTTP handler middleware
    httpHandler := authmw.HTTPMiddleware(jwtValidator,
        "/auth/register",     // public endpoints
        "/auth/login",
        "/health",
        "/metrics",
        "/swagger/",
    )(mux)

    // Запускаем HTTP сервер
    http.ListenAndServe(":8080", httpHandler)
}
```

## URL JWKS endpoint

Auth сервис должен предоставлять JWKS на одном из путей:
- `GET /jwks` (через gRPC gateway)
- `GET /.well-known/jwks.json` (стандартный путь)

## Конфигурация auth сервиса

Убедитесь что `auth/config.yaml` содержит все сервисы в `auth.audience`:

```yaml
auth:
  issuer: balun-auth-service
  audience:
    - users
    - social
    - chat
    - gateway
```

## Тестирование

### 1. Получите токен

```bash
# Регистрация
grpc_cli call localhost:8081 AuthService/Register '{"email":"test@test.com", "password":"123456"}'

# Логин
grpc_cli call localhost:8081 AuthService/Login '{"email":"test@test.com", "password":"123456"}'
# Ответ: {"userId":"...", "accessToken":"...", "refreshToken":"..."}
```

### 2. Используйте токен в запросах

```bash
# gRPC с metadata
grpc_cli call --metadata authorization:"Bearer <access_token>" localhost:8082 UsersService/GetProfile '{}'

# HTTP
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/users/profile
```

## Troubleshooting

### "missing authorization header"
- Убедитесь что передаете header `authorization: Bearer <token>`

### "invalid token"
- Проверьте что токен валидный (не истек)
- Проверьте что issuer совпадает в токене и validator

### "invalid audience"
- Проверьте что audience сервиса присутствует в списке auth.audience в auth/config.yaml
- Проверьте что ExpectedAudience в validator совпадает с именем сервиса

### "key with KID not found"
- Убедитесь что JWKS cache успешно загружен
- Проверьте доступность auth сервиса
- Проверьте логи auth сервиса на наличие ошибок при генерации ключей
