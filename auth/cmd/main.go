package main

import (
	"log"
	"net"

	"auth/internal/app/adapters"
	deliveryGrpc "auth/internal/app/delivery/grpc"
	"auth/internal/app/repository"
	"auth/internal/app/usecase"

	authPb "auth/pkg/api"
	usersPb "auth/pkg/users/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	errorsMiddleware "auth/internal/middleware/errors"
)

func main() {
	usersConn, err := grpc.NewClient("users:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to users service: %v", err)
	}

	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(usersConn))

	repo := repository.NewUsersRepositoryStub()

	authUsecase := usecase.NewUsecase(usersClient, repo)

	controller := deliveryGrpc.NewAuthController(authUsecase)

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			errorsMiddleware.ErrorsUnaryInterceptor(),
		),
	)
	authPb.RegisterAuthServiceServer(server, controller) // регистрация обработчиков

	reflection.Register(server) // регистрируем дополнительные обработчики

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	// Register:
	// grpc_cli call --json_input --json_output localhost:8081 AuthService/Register '{"email":"stas@gmail.com", "password":"123456"}'
	// Login:
	// grpc_cli call --json_input --json_output localhost:8081 AuthService/Login '{"email":"stas@gmail.com", "password":"123456"}'
	// Refresh:
	// grpc_cli call --json_input --json_output localhost:8081 AuthService/Refresh '{"refreshToken":"ff9bf764-0bfc-4e5d-b59f-cdaeeececb06"}'
}
