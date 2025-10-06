package main

import (
	"log"
	"net"

	"social/internal/app/adapters"
	"social/internal/app/repository"
	"social/internal/app/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	deliveryGrpc "social/internal/app/delivery/grpc"
	errorsMiddleware "social/internal/middleware/errors"
	socialPb "social/pkg/api"
	usersPb "social/pkg/users/api"
)

func main() {
	usersConn, err := grpc.NewClient("users:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to users service: %v", err)
	}

	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(usersConn))

	repo := repository.NewInMemorySocialRepository()

	socialUsecase := usecase.NewUsecase(usersClient, repo)

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
