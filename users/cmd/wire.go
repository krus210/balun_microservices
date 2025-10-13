//go:build wireinject
// +build wireinject

package main

import (
	"users/internal/app/delivery/grpc"
	"users/internal/app/repository"
	"users/internal/app/usecase"
	errorsMiddleware "users/internal/middleware/errors"
	pb "users/pkg/api"

	"github.com/google/wire"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// provideRepository создает репозиторий и возвращает его как интерфейс
func provideRepository() usecase.UsersRepository {
	return repository.NewUsersRepositoryStub()
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
func InitializeApp() (*grpcLib.Server, error) {
	wire.Build(
		provideRepository,
		provideUsecase,
		grpc.NewUsersController,
		provideGRPCServer,
	)
	return nil, nil
}
