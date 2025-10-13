package grpc

import (
	"context"
	"log"

	pb "users/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *UsersController) GetProfileByNickname(ctx context.Context, req *pb.GetProfileByNicknameRequest) (*pb.GetProfileByNicknameResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	userProfile, err := h.usecase.GetProfileByNickname(ctx, req.Nickname)
	if err != nil {
		return nil, err
	}

	return &pb.GetProfileByNicknameResponse{
		UserProfile: newPbUserProfileFromUserProfile(userProfile),
	}, nil
}
