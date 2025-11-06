//go:build wireinject
// +build wireinject

package main

import (
	"context"

	lib "github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/postgres"
	"users/internal/app/delivery/grpc"
	"users/internal/app/repository"
	"users/internal/app/usecase"
	errorsMiddleware "users/internal/middleware/errors"
	pb "users/pkg/api"

	"github.com/google/wire"
	grpcLib "google.golang.org/grpc"
)

// providePostgresConnection создает connection pool для postgres и cleanup функцию
func providePostgresConnection(ctx context.Context, cfg *config.StandardServiceConfig) (*postgres.Connection, func(), error) {
	conn, cleanup, err := lib.InitPostgres(ctx, cfg.Database)
	if err != nil {
		return nil, nil, err
	}
	return conn, cleanup, nil
}

// provideTransactionManager создает transaction manager
func provideTransactionManager(conn *postgres.Connection) postgres.TransactionManagerAPI {
	return postgres.NewTransactionManager(conn)
}

// provideRepository создает репозиторий и возвращает его как интерфейс
func provideRepository(tm postgres.TransactionManagerAPI) usecase.UsersRepository {
	return repository.NewRepository(tm)
}

// provideUsecase создает usecase и возвращает его как интерфейс
func provideUsecase(repo usecase.UsersRepository) usecase.Usecase {
	return usecase.NewUsecase(repo)
}

// provideGRPCServer создает gRPC сервер с middleware и регистрирует контроллер
func provideGRPCServer(controller *grpc.UsersController) *grpcLib.Server {
	server := lib.InitGRPCServer(
		errorsMiddleware.ErrorsUnaryInterceptor(),
	)

	pb.RegisterUsersServiceServer(server, controller)

	return server
}

// InitializeApp - injector функция, которую сгенерирует Wire
// Возвращает gRPC сервер и cleanup функцию для закрытия ресурсов
func InitializeApp(ctx context.Context, cfg *config.StandardServiceConfig) (*grpcLib.Server, func(), error) {
	wire.Build(
		providePostgresConnection,
		provideTransactionManager,
		provideRepository,
		provideUsecase,
		grpc.NewUsersController,
		provideGRPCServer,
	)
	return nil, nil, nil
}
