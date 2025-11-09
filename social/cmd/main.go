package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/postgres"

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

	log.Printf("Starting %s service (version: %s, environment: %s)",
		cfg.Service.Name, cfg.Service.Version, cfg.Service.Environment)

	// Подключаемся к Users сервису
	usersConn, usersCleanup, err := app.InitGRPCClient(ctx, cfg.UsersService)
	if err != nil {
		log.Fatalf("failed to connect to users service: %v", err)
	}
	defer usersCleanup()

	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(usersConn))

	// Инициализируем PostgreSQL через lib/app
	conn, cleanup, err := app.InitPostgres(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	defer cleanup()

	// Создаем Transaction Manager
	txMngr := postgres.NewTransactionManager(conn)

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
	friendRequestRepo := repository.NewRepository(txMngr)
	outboxRepo := outboxRepository.NewRepository(txMngr)

	// Создаем и запускаем outbox worker
	worker := outboxProcessor.NewOutboxFriendRequestWorker(outboxRepo, txMngr, friendRequestEventsHandler,
		outboxProcessor.WithBatchSize(cfg.Outbox.Processor.BatchSize),
		outboxProcessor.WithMaxRetry(cfg.Outbox.Processor.MaxRetry),
		outboxProcessor.WithRetryInterval(cfg.Outbox.Processor.RetryInterval),
		outboxProcessor.WithWindow(cfg.Outbox.Processor.Window),
	)

	go worker.Run(ctx)

	// Создаем use cases и controller
	outboxProc := outboxProcessor.NewProcessor(outboxProcessor.Deps{Repository: outboxRepo})
	socialUsecase := usecase.NewUsecase(usersClient, friendRequestRepo, outboxProc, txMngr)
	controller := deliveryGrpc.NewSocialController(socialUsecase)

	// Инициализируем gRPC сервер через lib/app
	server := app.InitGRPCServer(errorsMiddleware.ErrorsUnaryInterceptor())
	socialPb.RegisterSocialServiceServer(server, controller)

	// Запускаем gRPC сервер
	listenAddr := fmt.Sprintf(":%d", cfg.Server.GRPC.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
