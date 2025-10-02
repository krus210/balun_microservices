package grpc

import (
	"auth/internal/usecase/dto"
	pb "auth/pkg/api"
	"context"
	"log"

	"google.golang.org/grpc/metadata"
)

func (h *AuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	user, err := h.usecases.Refresh(ctx, dto.RefreshRequest{
		UserID:       1, // TODO: получать из хедера
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	return &pb.RefreshResponse{
		UserId:       user.ID,
		AccessToken:  user.Token.AccessToken,
		RefreshToken: user.Token.RefreshToken,
	}, nil
}
