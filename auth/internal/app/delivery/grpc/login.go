package grpc

import (
	"context"
	"log"

	"auth/internal/app/usecase/dto"

	pb "auth/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *AuthController) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
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

	user, err := h.usecase.Login(ctx, dto.LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		UserId:       user.ID,
		AccessToken:  user.Token.AccessToken,
		RefreshToken: user.Token.RefreshToken,
	}, nil
}
