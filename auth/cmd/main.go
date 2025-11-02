package main

import (
	"auth/internal/app/adapters"
	deliveryGrpc "auth/internal/app/delivery/grpc"
	"auth/internal/app/repository"
	"auth/internal/app/usecase"
	"auth/internal/config"
	errorsMiddleware "auth/internal/middleware/errors"
	"fmt"
	"log"
	"net"

	authPb "auth/pkg/api"
	usersPb "auth/pkg/users/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Starting %s service (version: %s, environment: %s)",
		cfg.Service.Name, cfg.Service.Version, cfg.Service.Environment)

	usersAddr := fmt.Sprintf("%s:%d", cfg.UsersService.Host, cfg.UsersService.Port)
	usersConn, err := grpc.NewClient(usersAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to users service: %v", err)
	}
	defer usersConn.Close()

	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(usersConn))

	repo := repository.NewUsersRepositoryStub()

	authUsecase := usecase.NewUsecase(usersClient, repo)

	controller := deliveryGrpc.NewAuthController(authUsecase)

	listenAddr := fmt.Sprintf(":%d", cfg.Server.GRPC.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			errorsMiddleware.ErrorsUnaryInterceptor(),
		),
	)
	authPb.RegisterAuthServiceServer(server, controller)

	reflection.Register(server)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
