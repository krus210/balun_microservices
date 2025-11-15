package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/logger"

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
		logger.FatalKV(ctx, "failed to load config", "error", err.Error())
	}

	// Создаем приложение
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		logger.FatalKV(ctx, "failed to create app", "error", err.Error())
	}

	// Инициализируем logger
	if err := application.InitLogger(cfg.Logger, cfg.Service.Name, cfg.Service.Environment); err != nil {
		logger.FatalKV(ctx, "failed to initialize logger", "error", err.Error())
	}

	// Инициализируем tracer
	if err := application.InitTracer(cfg.Tracer); err != nil {
		logger.FatalKV(ctx, "failed to initialize tracer", "error", err.Error())
	}

	// Инициализируем metrics
	if err := application.InitMetrics(cfg.Metrics, cfg.Service.Name); err != nil {
		logger.FatalKV(ctx, "failed to initialize metrics", "error", err.Error())
	}

	// Инициализируем admin HTTP сервер (метрики и pprof)
	if cfg.Server.Admin != nil {
		if err := application.InitAdminServer(*cfg.Server.Admin); err != nil {
			logger.FatalKV(ctx, "failed to initialize admin server", "error", err.Error())
		}
	}

	logger.InfoKV(ctx, "starting chat service",
		"version", cfg.Service.Version,
		"environment", cfg.Service.Environment,
		"grpc_port", cfg.Server.GRPC.Port,
	)

	// Инициализируем PostgreSQL
	if err := application.InitPostgres(ctx, cfg.Database); err != nil {
		logger.FatalKV(ctx, "failed to init postgres", "error", err.Error())
	}

	// Подключаемся к Users сервису
	if err := application.InitGRPCClient(ctx, "users", cfg.UsersService); err != nil {
		logger.FatalKV(ctx, "failed to connect to users service", "error", err.Error())
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

	// Запускаем gRPC сервер через errgroup
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.InfoKV(gCtx, "starting chat service gRPC server", "grpc_port", cfg.Server.GRPC.Port)
		if err := application.Run(gCtx, *cfg.Server.GRPC); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	// Запускаем admin HTTP сервер
	if cfg.Server.Admin != nil {
		g.Go(func() error {
			logger.InfoKV(gCtx, "starting admin HTTP server", "admin_port", cfg.Server.Admin.Port)
			if err := application.ServeAdmin(gCtx); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		})
	}

	waitErr := app.WaitForShutdown(ctx, g.Wait, app.GracefulShutdownTimeout)
	switch {
	case waitErr == nil || errors.Is(waitErr, context.Canceled):
		logger.InfoKV(ctx, "chat service components stopped")
	case errors.Is(waitErr, context.DeadlineExceeded):
		logger.WarnKV(ctx, "graceful shutdown timeout exceeded, forcing cleanup")
	default:
		logger.FatalKV(ctx, "failed to serve", "error", waitErr.Error())
	}

	logger.InfoKV(ctx, "chat service shutdown complete")
}
