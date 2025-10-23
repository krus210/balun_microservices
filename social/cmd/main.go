package main

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"social/internal/app/delivery/friend_request_handler"
	"social/pkg/kafka"

	"social/internal/app/adapters"
	"social/internal/app/repository"
	"social/internal/app/usecase"
	"social/pkg/postgres"
	"social/pkg/postgres/transaction_manager"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	usersConn, err := grpc.NewClient("users:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to users service: %v", err)
	}

	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(usersConn))

	conn, err := postgres.NewConnectionPool(ctx, DSN(),
		postgres.WithMaxConnIdleTime(time.Minute),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	producer, err := kafka.NewSyncProducer(strings.Split(kafkaBrokers, ","), nil)
	if err != nil {
		log.Fatal(err)
	}

	friendRequestEventsHanler := friend_request_handler.NewKafkaOrderBatchHandler(producer,
		friend_request_handler.WithMaxBatchSize(100),
		friend_request_handler.WithTopic(kafkaFriendRequestEventTopicName),
	)

	txMngr := transaction_manager.New(conn)

	friendRequestRepo := repository.NewRepository(txMngr)
	outboxRepo := outboxRepository.NewRepository(txMngr)

	worker := outboxProcessor.NewOutboxFriendRequestWorker(outboxRepo, txMngr, friendRequestEventsHanler,
		outboxProcessor.WithBatchSize(10),
		outboxProcessor.WithMaxRetry(10),
		outboxProcessor.WithRetryInterval(30*time.Second),
		outboxProcessor.WithWindow(time.Hour),
	)

	go worker.Run(ctx)

	outboxProc := outboxProcessor.NewProcessor(outboxProcessor.Deps{Repository: outboxRepo})

	socialUsecase := usecase.NewUsecase(usersClient, friendRequestRepo, outboxProc, txMngr)

	controller := deliveryGrpc.NewSocialController(socialUsecase)

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			errorsMiddleware.ErrorsUnaryInterceptor(),
		),
	)
	socialPb.RegisterSocialServiceServer(server, controller) // регистрация обработчиков

	reflection.Register(server) // регистрируем дополнительные обработчики

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
