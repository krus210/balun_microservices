package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"

	"auth/internal/app/adapters"
	deliveryGrpc "auth/internal/app/delivery/grpc"
	"auth/internal/app/repository"
	"auth/internal/app/usecase"
	errorsMiddleware "auth/internal/middleware/errors"

	authPb "auth/pkg/api"
	usersPb "auth/pkg/users/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию через lib/config
	cfg, err := config.LoadServiceConfig(ctx, "auth",
		config.WithUsersService("users", 8082),
	)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Starting %s service (version: %s, environment: %s)",
		cfg.Service.Name, cfg.Service.Version, cfg.Service.Environment)

	// Подключаемся к Users сервису
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

	// Создаем gRPC сервер через lib/app
	server := app.InitGRPCServer(
		errorsMiddleware.ErrorsUnaryInterceptor(),
	)
	authPb.RegisterAuthServiceServer(server, controller)

	log.Printf("gRPC server listening on port %d", cfg.Server.GRPC.Port)

	// Запускаем gRPC сервер через lib/app
	if err := app.ServeGRPC(ctx, server, *cfg.Server.GRPC); err != nil {
		if err == context.Canceled {
			log.Println("shutdown")
		} else {
			log.Fatalf("failed to serve: %v", err)
		}
	}
}
