package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"

	"chat/internal/app/adapters"
	"chat/internal/app/repository"
	"chat/internal/app/usecase"

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

	// Создаем приложение
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	// Инициализируем PostgreSQL
	if err := application.InitPostgres(ctx, cfg.Database); err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}

	// Подключаемся к Users сервису
	if err := application.InitGRPCClient(ctx, "users", cfg.UsersService); err != nil {
		log.Fatalf("failed to connect to users service: %v", err)
	}

	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(application.GetGRPCClient("users")))

	repo := repository.NewRepository(application.TransactionManager())

	chatUsecase := usecase.NewUsecase(usersClient, repo)

	controller := deliveryGrpc.NewChatController(chatUsecase)

	// Инициализируем gRPC сервер
	application.InitGRPCServer(cfg.Server, errorsMiddleware.ErrorsUnaryInterceptor())

	// Регистрируем gRPC сервисы
	application.RegisterGRPC(func(s *grpc.Server) {
		chatPb.RegisterChatServiceServer(s, controller)
	})

	// Запускаем приложение
	slog.Info("starting chat service", "grpc_port", cfg.Server.GRPC.Port)

	if err := application.Run(ctx, *cfg.Server.GRPC); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Fatalf("failed to serve: %v", err)
		}
	}
	slog.Info("chat service stopped gracefully")
}
