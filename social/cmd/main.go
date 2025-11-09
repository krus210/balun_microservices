package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

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
	defer producer.Close()

	// Создаем обработчик событий заявок в друзья
	friendRequestEventsHandler := friend_request_handler.NewKafkaFriendRequestBatchHandler(producer,
		friend_request_handler.WithMaxBatchSize(cfg.FriendRequestHandler.BatchSize),
		friend_request_handler.WithTopic(cfg.Kafka.Topics.FriendRequestEvents),
	)

	// Инициализируем репозитории
	friendRequestRepo := repository.NewRepository(application.TransactionManager())
	outboxRepo := outboxRepository.NewRepository(application.TransactionManager())

	// Создаем и запускаем outbox worker
	worker := outboxProcessor.NewOutboxFriendRequestWorker(outboxRepo, application.TransactionManager(), friendRequestEventsHandler,
		outboxProcessor.WithBatchSize(cfg.Outbox.Processor.BatchSize),
		outboxProcessor.WithMaxRetry(cfg.Outbox.Processor.MaxRetry),
		outboxProcessor.WithRetryInterval(cfg.Outbox.Processor.RetryInterval),
		outboxProcessor.WithWindow(cfg.Outbox.Processor.Window),
	)

	go worker.Run(ctx)

	// Создаем use cases и controller
	outboxProc := outboxProcessor.NewProcessor(outboxProcessor.Deps{Repository: outboxRepo})
	socialUsecase := usecase.NewUsecase(usersClient, friendRequestRepo, outboxProc, application.TransactionManager())
	controller := deliveryGrpc.NewSocialController(socialUsecase)

	// Инициализируем gRPC сервер
	application.InitGRPCServer(errorsMiddleware.ErrorsUnaryInterceptor())

	// Регистрируем gRPC сервисы
	application.RegisterGRPC(func(s *grpc.Server) {
		socialPb.RegisterSocialServiceServer(s, controller)
	})

	// Запускаем приложение
	if err := application.Run(ctx, *cfg.Server.GRPC); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Fatalf("failed to serve: %v", err)
		}
	}
}
