package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/postgres"

	"chat/internal/app/adapters"
	"chat/internal/app/repository"
	"chat/internal/app/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	deliveryGrpc "chat/internal/app/delivery/grpc"
	errorsMiddleware "chat/internal/middleware/errors"
	chatPb "chat/pkg/api"
	usersPb "chat/pkg/users/api"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию через lib/config
	cfg, err := config.LoadServiceConfig(ctx, "chat",
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

	// Инициализируем PostgreSQL через lib/app
	conn, cleanup, err := app.InitPostgres(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	defer cleanup()

	// Создаем Transaction Manager
	txMngr := postgres.NewTransactionManager(conn)

	repo := repository.NewRepository(txMngr)

	chatUsecase := usecase.NewUsecase(usersClient, repo)

	controller := deliveryGrpc.NewChatController(chatUsecase)

	// Создаем gRPC сервер через lib/app
	server := app.InitGRPCServer(
		errorsMiddleware.ErrorsUnaryInterceptor(),
	)
	chatPb.RegisterChatServiceServer(server, controller)

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
