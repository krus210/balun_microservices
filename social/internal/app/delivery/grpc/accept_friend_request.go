package grpc

import (
	"context"
	"log"

	"social/internal/app/models"
	"social/internal/app/usecase/dto"

	pb "social/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *SocialController) AcceptFriendRequest(ctx context.Context, req *pb.AcceptFriendRequestRequest) (*pb.AcceptFriendRequestResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	friendRequest, err := h.usecase.AcceptFriendRequest(ctx, dto.ChangeFriendRequestDto{
		UserID:    "1", // TODO: брать из хедера
		RequestID: models.FriendRequestID(req.RequestId),
	})
	if err != nil {
		return nil, err
	}

	return &pb.AcceptFriendRequestResponse{
		FriendRequest: newPbFriendRequestFromFriendRequest(friendRequest),
	}, nil
}
