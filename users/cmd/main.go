package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	deliveryGrpc "users/internal/app/delivery/grpc"
	errorsMiddleware "users/internal/middleware/errors"
	pb "users/pkg/api"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/authmw"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/logger"
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

	// Останавливаем JWKS кеш при завершении
	defer container.JWKSCache.Stop()

	// Создаем контроллер
	controller := deliveryGrpc.NewUsersController(container.Usecase)

	// Инициализируем gRPC сервер с JWT и errors middleware
	container.App.InitGRPCServer(
		cfg.Server,
		errorsMiddleware.ErrorsUnaryInterceptor(),
		authmw.UnaryServerInterceptor(container.JWTValidator),
	)

	// Регистрируем gRPC сервисы
	container.App.RegisterGRPC(func(s *grpc.Server) {
		pb.RegisterUsersServiceServer(s, controller)
	})

	// Запускаем приложение через errgroup
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.InfoKV(gCtx, "starting users service gRPC server", "grpc_port", cfg.Server.GRPC.Port)
		if err := container.App.Run(gCtx, *cfg.Server.GRPC); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	// Запускаем admin HTTP сервер
	if cfg.Server.Admin != nil {
		g.Go(func() error {
			logger.InfoKV(gCtx, "starting admin HTTP server", "admin_port", cfg.Server.Admin.Port)
			if err := container.App.ServeAdmin(gCtx); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		})
	}

	waitErr := app.WaitForShutdown(ctx, g.Wait, app.GracefulShutdownTimeout)
	switch {
	case waitErr == nil || errors.Is(waitErr, context.Canceled):
		logger.InfoKV(ctx, "users service components stopped")
	case errors.Is(waitErr, context.DeadlineExceeded):
		logger.WarnKV(ctx, "graceful shutdown timeout exceeded, forcing cleanup")
	default:
		logger.FatalKV(ctx, "failed to serve", "error", waitErr.Error())
	}

	logger.InfoKV(ctx, "users service shutdown complete")
}
