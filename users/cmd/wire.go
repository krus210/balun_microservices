//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"time"

	"users/internal/app/delivery/grpc"
	"users/internal/app/repository"
	"users/internal/app/usecase"
	errorsMiddleware "users/internal/middleware/errors"
	pb "users/pkg/api"
	"users/pkg/postgres"
	"users/pkg/postgres/transaction_manager"

	"github.com/google/wire"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// providePostgresConnection создает connection pool для postgres и cleanup функцию
func providePostgresConnection(ctx context.Context) (*postgres.Connection, func(), error) {
	conn, err := postgres.NewConnectionPool(ctx, DSN(),
		postgres.WithMaxConnIdleTime(time.Minute),
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
func provideTransactionManager(conn *postgres.Connection) transaction_manager.TransactionManagerAPI {
	return transaction_manager.New(conn)
}

// provideRepository создает репозиторий и возвращает его как интерфейс
func provideRepository(tm transaction_manager.TransactionManagerAPI) usecase.UsersRepository {
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
func InitializeApp(ctx context.Context) (*grpcLib.Server, func(), error) {
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
