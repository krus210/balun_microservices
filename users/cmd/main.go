package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	deliveryGrpc "users/internal/app/delivery/grpc"
	errorsMiddleware "users/internal/middleware/errors"
	pb "users/pkg/api"

	"github.com/sskorolev/balun_microservices/lib/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию напрямую через lib/config
	cfg, err := config.LoadServiceConfig(ctx, "users")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Инициализируем приложение через Wire
	container, cleanup, err := InitializeApp(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer cleanup()

	// Создаем контроллер
	controller := deliveryGrpc.NewUsersController(container.Usecase)

	// Инициализируем gRPC сервер
	container.App.InitGRPCServer(cfg.Server, errorsMiddleware.ErrorsUnaryInterceptor())

	// Регистрируем gRPC сервисы
	container.App.RegisterGRPC(func(s *grpc.Server) {
		pb.RegisterUsersServiceServer(s, controller)
	})

	// Запускаем приложение
	slog.Info("starting users service", "grpc_port", cfg.Server.GRPC.Port)

	if err := container.App.Run(ctx, *cfg.Server.GRPC); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Fatalf("failed to serve: %v", err)
		}
	}
	slog.Info("users service stopped gracefully")
}
