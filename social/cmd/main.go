package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/authmw"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/logger"

	"social/internal/app/adapters"
	"social/internal/app/delivery/friend_request_handler"
	deliveryGrpc "social/internal/app/delivery/grpc"
	outboxProcessor "social/internal/app/outbox/processor"
	outboxRepository "social/internal/app/outbox/repository"
	"social/internal/app/repository"
	"social/internal/app/usecase"
	errorsMiddleware "social/internal/middleware/errors"
	"social/pkg/kafka"

	socialPb "social/pkg/api"
	usersPb "social/pkg/users/api"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию через lib/config с явным указанием всех компонентов
	cfg, err := config.LoadServiceConfig(ctx, "social")
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

	logger.InfoKV(ctx, "starting social service",
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

	// Инициализируем auth компоненты (JWKS кеш и JWT validator)
	authComponents, authCleanup, err := app.InitAuthComponents(
		ctx,
		cfg.AuthService,
		"social", // audience для social сервиса
	)
	if err != nil {
		logger.FatalKV(ctx, "failed to initialize auth components", "error", err.Error())
	}
	defer authCleanup()

	// Создаем Kafka producer
	producer, err := kafka.NewSyncProducer([]string{cfg.Kafka.GetBrokers()}, cfg.Kafka.ClientID, nil)
	if err != nil {
		logger.FatalKV(ctx, "failed to create kafka producer", "error", err.Error())
	}

	// Добавляем cleanup для Kafka producer
	defer func() {
		logger.InfoKV(ctx, "closing kafka producer")
		if err := producer.Close(); err != nil {
			logger.ErrorKV(ctx, "failed to close kafka producer", "error", err.Error())
		}
	}()

	// Создаем обработчик событий заявок в друзья
	friendRequestEventsHandler := friend_request_handler.NewKafkaFriendRequestBatchHandler(producer,
		friend_request_handler.WithMaxBatchSize(cfg.FriendRequestHandler.BatchSize),
		friend_request_handler.WithTopic(cfg.Kafka.Topics.FriendRequestEvents),
	)

	// Инициализируем репозитории
	friendRequestRepo := repository.NewRepository(application.TransactionManager())
	outboxRepo := outboxRepository.NewRepository(application.TransactionManager())

	// Создаем outbox worker
	worker := outboxProcessor.NewOutboxFriendRequestWorker(outboxRepo, application.TransactionManager(), friendRequestEventsHandler,
		outboxProcessor.WithBatchSize(cfg.Outbox.Processor.BatchSize),
		outboxProcessor.WithMaxRetry(cfg.Outbox.Processor.MaxRetry),
		outboxProcessor.WithRetryInterval(cfg.Outbox.Processor.RetryInterval),
		outboxProcessor.WithWindow(cfg.Outbox.Processor.Window),
	)

	// Создаем use cases и controller
	outboxProc := outboxProcessor.NewProcessor(outboxProcessor.Deps{Repository: outboxRepo})
	socialUsecase := usecase.NewUsecase(usersClient, friendRequestRepo, outboxProc, application.TransactionManager())
	controller := deliveryGrpc.NewSocialController(socialUsecase)

	// Инициализируем gRPC сервер с JWT и errors middleware
	application.InitGRPCServer(
		cfg.Server,
		errorsMiddleware.ErrorsUnaryInterceptor(),
		authmw.UnaryServerInterceptor(authComponents.JWTValidator),
	)

	// Регистрируем gRPC сервисы
	application.RegisterGRPC(func(s *grpc.Server) {
		socialPb.RegisterSocialServiceServer(s, controller)
	})

	// Запускаем gRPC сервер и outbox worker через errgroup
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.InfoKV(gCtx, "starting social service gRPC server", "grpc_port", cfg.Server.GRPC.Port)
		if err := application.Run(gCtx, *cfg.Server.GRPC); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		logger.InfoKV(gCtx, "starting outbox worker")
		worker.Run(gCtx)
		logger.InfoKV(gCtx, "outbox worker stopped")
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
		logger.InfoKV(ctx, "social service components stopped")
	case errors.Is(waitErr, context.DeadlineExceeded):
		logger.WarnKV(ctx, "graceful shutdown timeout exceeded, forcing cleanup")
	default:
		logger.FatalKV(ctx, "failed to serve", "error", waitErr.Error())
	}

	logger.InfoKV(ctx, "social service shutdown complete")
}
