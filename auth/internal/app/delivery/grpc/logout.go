package grpc

import (
	"context"

	"auth/internal/app/usecase/dto"

	pb "auth/pkg/api"
)

func (h *AuthController) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.usecase.Logout(ctx, dto.LogoutRequest{
		RefreshToken: req.GetRefreshToken(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.LogoutResponse{}, nil
}
