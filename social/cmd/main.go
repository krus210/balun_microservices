package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"

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

	// Создаем Kafka producer
	producer, err := kafka.NewSyncProducer([]string{cfg.Kafka.GetBrokers()}, cfg.Kafka.ClientID, nil)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}

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

	// Инициализируем gRPC сервер
	application.InitGRPCServer(cfg.Server, errorsMiddleware.ErrorsUnaryInterceptor())

	// Регистрируем gRPC сервисы
	application.RegisterGRPC(func(s *grpc.Server) {
		socialPb.RegisterSocialServiceServer(s, controller)
	})

	// Запускаем gRPC сервер и outbox worker через errgroup
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("starting social service gRPC server", "grpc_port", cfg.Server.GRPC.Port)
		if err := application.Run(gCtx, *cfg.Server.GRPC); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		slog.Info("starting outbox worker")
		worker.Run(gCtx)
		slog.Info("outbox worker stopped")
		return nil
	})

	slog.Info("starting social service", "grpc_port", cfg.Server.GRPC.Port)

	waitErr := app.WaitForShutdown(ctx, g.Wait, app.GracefulShutdownTimeout)
	switch {
	case waitErr == nil || errors.Is(waitErr, context.Canceled):
		slog.Info("social service components stopped")
	case errors.Is(waitErr, context.DeadlineExceeded):
		slog.Warn("graceful shutdown timeout exceeded, forcing cleanup")
	default:
		log.Fatalf("failed to serve: %v", waitErr)
	}

	slog.Info("closing kafka producer...")
	if err := producer.Close(); err != nil {
		slog.Error("failed to close kafka producer", "error", err)
	}

	slog.Info("social service shutdown complete")
}
