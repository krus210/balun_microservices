package grpc

import (
	"context"
	"log"

	"social/internal/app/usecase/dto"

	pb "social/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *SocialController) DeclineFriendRequest(ctx context.Context, req *pb.DeclineFriendRequestRequest) (*pb.DeclineFriendRequestResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	friendRequest, err := h.usecase.DeclineFriendRequest(ctx, dto.ChangeFriendRequestDto{
		UserID:    1, // TODO: брать из хедера
		RequestID: req.RequestId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.DeclineFriendRequestResponse{
		FriendRequest: newPbFriendRequestFromFriendRequest(friendRequest),
	}, nil
}
