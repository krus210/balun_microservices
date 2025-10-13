package grpc

import (
	"context"
	"log"

	"users/internal/app/usecase/dto"

	pb "users/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *UsersController) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	userProfile, err := h.usecase.UpdateProfile(ctx, dto.UpdateProfileRequest{
		UserID:    req.UserId,
		Nickname:  req.Nickname,
		Bio:       req.Bio,
		AvatarURL: req.AvatarUrl,
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateProfileResponse{
		UserProfile: newPbUserProfileFromUserProfile(userProfile),
	}, nil
}
