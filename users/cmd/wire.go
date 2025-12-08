//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"users/internal/app/repository"
	"users/internal/app/usecase"

	"github.com/google/wire"

	lib "github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/authmw"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/postgres"
)

// AppContainer содержит App, Usecase и JWTValidator для передачи из Wire
type AppContainer struct {
	App          *lib.App
	Usecase      usecase.Usecase
	JWTValidator *authmw.Validator
	JWKSCache    authmw.JWKSProvider
}

// provideApp создает App и инициализирует компоненты
func provideApp(ctx context.Context, cfg *config.StandardServiceConfig) (*lib.App, error) {
	app, err := lib.NewApp(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Инициализируем logger
	if err := app.InitLogger(cfg.Logger, cfg.Service.Name, cfg.Service.Environment); err != nil {
		return nil, err
	}

	// Инициализируем tracer
	if err := app.InitTracer(cfg.Tracer); err != nil {
		return nil, err
	}

	// Инициализируем metrics
	if err := app.InitMetrics(cfg.Metrics, cfg.Service.Name); err != nil {
		return nil, err
	}

	// Инициализируем admin HTTP сервер (метрики и pprof)
	if cfg.Server.Admin != nil {
		if err := app.InitAdminServer(*cfg.Server.Admin); err != nil {
			return nil, err
		}
	}

	// Инициализируем PostgreSQL
	if err := app.InitPostgres(ctx, cfg.Database); err != nil {
		return nil, err
	}

	return app, nil
}

// provideTransactionManager получает transaction manager из App
func provideTransactionManager(app *lib.App) postgres.TransactionManagerAPI {
	return app.TransactionManager()
}

// provideRepository создает репозиторий и возвращает его как интерфейс
func provideRepository(tm postgres.TransactionManagerAPI) usecase.UsersRepository {
	return repository.NewRepository(tm)
}

// provideUsecase создает usecase и возвращает его как интерфейс
func provideUsecase(repo usecase.UsersRepository) usecase.Usecase {
	return usecase.NewUsecase(repo)
}

// provideAuthComponents создает и инициализирует auth компоненты
func provideAuthComponents(ctx context.Context, cfg *config.StandardServiceConfig) (*lib.AuthComponents, func(), error) {
	return lib.InitAuthComponents(ctx, cfg.AuthService, "users")
}

// provideJWKSCache извлекает JWKS кеш из auth компонентов
func provideJWKSCache(authComponents *lib.AuthComponents) authmw.JWKSProvider {
	return authComponents.JWKSCache
}

// provideJWTValidator извлекает JWT validator из auth компонентов
func provideJWTValidator(authComponents *lib.AuthComponents) *authmw.Validator {
	return authComponents.JWTValidator
}

// provideAppContainer создает контейнер с App, Usecase и JWT компонентами
func provideAppContainer(app *lib.App, uc usecase.Usecase, validator *authmw.Validator, cache authmw.JWKSProvider) *AppContainer {
	return &AppContainer{
		App:          app,
		Usecase:      uc,
		JWTValidator: validator,
		JWKSCache:    cache,
	}
}

// InitializeApp - injector функция, которую сгенерирует Wire
// Возвращает AppContainer и cleanup функцию
func InitializeApp(ctx context.Context, cfg *config.StandardServiceConfig) (*AppContainer, func(), error) {
	wire.Build(
		provideApp,
		provideTransactionManager,
		provideRepository,
		provideUsecase,
		provideAuthComponents, // Заменяет provideAuthGRPCConn, provideJWKSCache, provideJWTValidator
		provideJWKSCache,
		provideJWTValidator,
		provideAppContainer,
	)
	return nil, nil, nil
}
