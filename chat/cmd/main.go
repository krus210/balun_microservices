package main

import (
	"context"
	"fmt"
	"lib/postgres"
	"log"
	"net"
	"os/signal"
	"syscall"

	"chat/internal/app/adapters"
	"chat/internal/app/repository"
	"chat/internal/app/usecase"
	"chat/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	deliveryGrpc "chat/internal/app/delivery/grpc"
	errorsMiddleware "chat/internal/middleware/errors"
	chatPb "chat/pkg/api"
	usersPb "chat/pkg/users/api"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию с интеграцией secrets
	cfg, err := config.LoadWithSecrets(ctx)
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

	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(usersConn))

	// Подключаемся к PostgreSQL
	conn, txMngr, err := postgres.New(ctx,
		postgres.WithHost(cfg.Database.Host),
		postgres.WithPort(cfg.Database.Port),
		postgres.WithDatabase(cfg.Database.Name),
		postgres.WithUser(cfg.Database.User),
		postgres.WithPassword(cfg.Database.Password),
		postgres.WithSSLMode(cfg.Database.SSLMode),
		postgres.WithMaxConnIdleTime(cfg.Database.MaxConnIdleTime),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	repo := repository.NewRepository(txMngr)

	chatUsecase := usecase.NewUsecase(usersClient, repo)

	controller := deliveryGrpc.NewChatController(chatUsecase)

	// Запускаем gRPC сервер
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
	chatPb.RegisterChatServiceServer(server, controller)

	reflection.Register(server)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
