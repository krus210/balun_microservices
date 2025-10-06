package grpc

import (
	"context"
	"log"

	"social/internal/app/usecase/dto"

	pb "social/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *SocialController) RemoveFriend(ctx context.Context, req *pb.RemoveFriendRequest) (*pb.RemoveFriendResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	err := h.usecase.RemoveFriend(ctx, dto.FriendRequestDto{
		FromUserID: 1, // TODO: брать из хедера
		ToUserID:   req.UserId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.RemoveFriendResponse{}, nil
}
