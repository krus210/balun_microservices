package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"google.golang.org/grpc"

	"auth/internal/app/adapters"
	deliveryGrpc "auth/internal/app/delivery/grpc"
	"auth/internal/app/repository"
	"auth/internal/app/usecase"
	errorsMiddleware "auth/internal/middleware/errors"

	authPb "auth/pkg/api"
	usersPb "auth/pkg/users/api"
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

	// Создаем приложение
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	// Подключаемся к Users сервису
	if err := application.InitGRPCClient(ctx, "users", cfg.UsersService); err != nil {
		log.Fatalf("failed to connect to users service: %v", err)
	}

	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(application.GetGRPCClient("users")))

	repo := repository.NewUsersRepositoryStub()

	authUsecase := usecase.NewUsecase(usersClient, repo)

	controller := deliveryGrpc.NewAuthController(authUsecase)

	// Инициализируем gRPC сервер
	application.InitGRPCServer(cfg.Server, errorsMiddleware.ErrorsUnaryInterceptor())

	// Регистрируем gRPC сервисы
	application.RegisterGRPC(func(s *grpc.Server) {
		authPb.RegisterAuthServiceServer(s, controller)
	})

	// Запускаем приложение
	if err := application.Run(ctx, *cfg.Server.GRPC); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Fatalf("failed to serve: %v", err)
		}
	}
}
