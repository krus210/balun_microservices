package main

import (
	"chat/pkg/postgres"
	"chat/pkg/postgres/transaction_manager"
	"context"
	"log"
	"net"
	"time"

	"chat/internal/app/adapters"
	"chat/internal/app/repository"
	"chat/internal/app/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	deliveryGrpc "chat/internal/app/delivery/grpc"
	errorsMiddleware "chat/internal/middleware/errors"
	chatPb "chat/pkg/api"
	usersPb "chat/pkg/users/api"
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

	txMngr := transaction_manager.New(conn)
	repo := repository.NewRepository(txMngr)

	chatUsecase := usecase.NewUsecase(usersClient, repo)

	controller := deliveryGrpc.NewChatController(chatUsecase)

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			errorsMiddleware.ErrorsUnaryInterceptor(),
		),
	)
	chatPb.RegisterChatServiceServer(server, controller) // регистрация обработчиков

	reflection.Register(server) // регистрируем дополнительные обработчики

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
