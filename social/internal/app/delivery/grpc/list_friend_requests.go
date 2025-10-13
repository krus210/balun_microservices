package grpc

import (
	"context"
	"log"

	"social/internal/app/models"
	pb "social/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *SocialController) ListRequests(ctx context.Context, req *pb.ListRequestsRequest) (*pb.ListRequestsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	friendRequests, err := h.usecase.ListFriendRequests(ctx, models.UserID(req.GetToUserId()))
	if err != nil {
		return nil, err
	}

	return &pb.ListRequestsResponse{
		Requests: newPbFriendRequestsFromFriendRequests(friendRequests),
	}, nil
}
