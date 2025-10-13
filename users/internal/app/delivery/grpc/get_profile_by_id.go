package grpc

import (
	"context"
	"log"

	pb "users/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *UsersController) GetProfileByID(ctx context.Context, req *pb.GetProfileByIDRequest) (*pb.GetProfileByIDResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	userProfile, err := h.usecase.GetProfileByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetProfileByIDResponse{
		UserProfile: newPbUserProfileFromUserProfile(userProfile),
	}, nil
}
