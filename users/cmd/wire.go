//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"lib/postgres"

	"users/internal/app/delivery/grpc"
	"users/internal/app/repository"
	"users/internal/app/usecase"
	"users/internal/config"
	errorsMiddleware "users/internal/middleware/errors"
	pb "users/pkg/api"

	"github.com/google/wire"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// providePostgresConnection создает connection pool для postgres и cleanup функцию
func providePostgresConnection(ctx context.Context, cfg *config.Config) (*postgres.Connection, func(), error) {
	conn, _, err := postgres.New(ctx,
		postgres.WithHost(cfg.Database.Host),
		postgres.WithPort(cfg.Database.Port),
		postgres.WithDatabase(cfg.Database.Name),
		postgres.WithUser(cfg.Database.User),
		postgres.WithPassword(cfg.Database.Password),
		postgres.WithSSLMode(cfg.Database.SSLMode),
		postgres.WithMaxConnIdleTime(cfg.Database.MaxConnIdleTime),
	)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		conn.Close()
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
	server := grpcLib.NewServer(
		grpcLib.ChainUnaryInterceptor(
			errorsMiddleware.ErrorsUnaryInterceptor(),
		),
	)

	pb.RegisterUsersServiceServer(server, controller)
	reflection.Register(server)

	return server
}

// InitializeApp - injector функция, которую сгенерирует Wire
// Возвращает gRPC сервер и cleanup функцию для закрытия ресурсов
func InitializeApp(ctx context.Context, cfg *config.Config) (*grpcLib.Server, func(), error) {
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
