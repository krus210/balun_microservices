package main

import (
	"log"
	"net"

	"users/internal/app/repository"
	"users/internal/app/usecase"
	pb "users/pkg/api"

	deliveryGrpc "users/internal/app/delivery/grpc"
	errorsMiddleware "users/internal/middleware/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	repo := repository.NewUsersRepositoryStub()

	usersUsecase := usecase.NewUsecase(repo)

	controller := deliveryGrpc.NewUsersController(usersUsecase)

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			errorsMiddleware.ErrorsUnaryInterceptor(),
		),
	)
	pb.RegisterUsersServiceServer(server, controller) // регистрация обработчиков

	reflection.Register(server) // регистрируем дополнительные обработчики

	log.Printf("server listening at %v", lis.Addr())
}
