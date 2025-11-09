//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"users/internal/app/repository"
	"users/internal/app/usecase"

	"github.com/google/wire"
	lib "github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/postgres"
)

// AppContainer содержит App и Usecase для передачи из Wire
type AppContainer struct {
	App     *lib.App
	Usecase usecase.Usecase
}

// provideApp создает App и инициализирует PostgreSQL
func provideApp(ctx context.Context, cfg *config.StandardServiceConfig) (*lib.App, error) {
	app, err := lib.NewApp(ctx, cfg)
	if err != nil {
		return nil, err
	}

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

// provideAppContainer создает контейнер с App и Usecase
func provideAppContainer(app *lib.App, uc usecase.Usecase) *AppContainer {
	return &AppContainer{
		App:     app,
		Usecase: uc,
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
		provideAppContainer,
	)
	return nil, nil, nil
}
