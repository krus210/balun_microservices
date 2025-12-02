package grpc

import (
	"context"
	"log"

	"auth/internal/app/usecase/dto"

	pb "auth/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *AuthController) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	user, err := h.usecase.Refresh(ctx, dto.RefreshRequest{
		RefreshToken: req.RefreshToken,
		DeviceID:     req.GetDeviceId(),
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
