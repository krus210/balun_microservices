package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sskorolev/balun_microservices/lib/authmw"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/logger"
	"google.golang.org/grpc"
)

// AuthComponents содержит JWT компоненты для аутентификации
type AuthComponents struct {
	JWKSCache    authmw.JWKSProvider
	JWTValidator *authmw.Validator
	AuthConn     *grpc.ClientConn
}

// InitAuthComponents создает и инициализирует auth компоненты
// Используется в сервисах users, social, chat для JWT аутентификации
func InitAuthComponents(
	ctx context.Context,
	authServiceCfg *config.TargetServiceConfig,
	audience string,
) (*AuthComponents, func(), error) {
	// Создаем gRPC соединение к auth сервису
	authConn, connCleanup, err := InitGRPCClient(ctx, authServiceCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	logger.InfoKV(ctx, "connected to auth service", "address", authServiceCfg.Address())

	// Создаем wrapper для вызова GetJWKS через gRPC
	authWrapper := authmw.NewGRPCClientWrapper(authConn)

	// Создаем JWKS кеш с автообновлением каждые 5 минут
	jwksCache, err := authmw.NewJWKSCacheGRPC(authmw.JWKSCacheGRPCConfig{
		Client:        authWrapper,
		RefreshPeriod: 5 * time.Minute,
		GRPCTimeout:   10 * time.Second,
	})
	if err != nil {
		connCleanup()
		return nil, nil, fmt.Errorf("failed to create JWKS cache: %w", err)
	}
	logger.InfoKV(ctx, "JWKS cache initialized")

	// Создаем JWT validator
	jwtValidator := authmw.NewValidator(authmw.ValidatorConfig{
		JWKSCache:        jwksCache,
		ExpectedIssuer:   "balun-auth-service",
		ExpectedAudience: audience,
	})
	logger.InfoKV(ctx, "JWT validator initialized", "audience", audience)

	components := &AuthComponents{
		JWKSCache:    jwksCache,
		JWTValidator: jwtValidator,
		AuthConn:     authConn,
	}

	// Cleanup функция для graceful shutdown
	cleanup := func() {
		jwksCache.Stop()
		connCleanup()
	}

	return components, cleanup, nil
}
