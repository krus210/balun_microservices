package grpc

import (
	"context"
	"log"

	"auth/internal/app/usecase/dto"

	pb "auth/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *AuthController) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	err := h.validateCredentials(req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	user, err := h.usecase.Register(ctx, dto.RegisterRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		UserId: user.ID,
	}, nil
}
