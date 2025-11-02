package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	"social/internal/app/delivery/friend_request_handler"
	"social/internal/config"
	"social/pkg/kafka"

	"social/internal/app/adapters"
	"social/internal/app/repository"
	"social/internal/app/usecase"

	"lib/postgres"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	deliveryGrpc "social/internal/app/delivery/grpc"
	outboxProcessor "social/internal/app/outbox/processor"
	outboxRepository "social/internal/app/outbox/repository"
	errorsMiddleware "social/internal/middleware/errors"
	socialPb "social/pkg/api"
	usersPb "social/pkg/users/api"
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

	// Создаем Kafka producer
	producer, err := kafka.NewSyncProducer([]string{cfg.Kafka.GetBrokers()}, cfg.Kafka.ClientID, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем обработчик событий заявок в друзья
	friendRequestEventsHandler := friend_request_handler.NewKafkaFriendRequestBatchHandler(producer,
		friend_request_handler.WithMaxBatchSize(cfg.FriendRequestHandler.BatchSize),
		friend_request_handler.WithTopic(cfg.Kafka.Topics.FriendRequestEvents),
	)

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

	outboxProc := outboxProcessor.NewProcessor(outboxProcessor.Deps{Repository: outboxRepo})

	socialUsecase := usecase.NewUsecase(usersClient, friendRequestRepo, outboxProc, txMngr)

	controller := deliveryGrpc.NewSocialController(socialUsecase)

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
	socialPb.RegisterSocialServiceServer(server, controller)

	reflection.Register(server)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
