package grpc

import (
	"context"
	"log"

	"social/internal/app/models"
	"social/internal/app/usecase/dto"

	pb "social/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *SocialController) SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestRequest) (*pb.SendFriendRequestResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	friendRequest, err := h.usecase.SendFriendRequest(ctx, dto.FriendRequestDto{
		FromUserID: models.UserID(1), // TODO: брать из хедера
		ToUserID:   models.UserID(req.ToUserId),
	})
	if err != nil {
		return nil, err
	}

	return &pb.SendFriendRequestResponse{
		FriendRequest: newPbFriendRequestFromFriendRequest(friendRequest),
	}, nil
}
