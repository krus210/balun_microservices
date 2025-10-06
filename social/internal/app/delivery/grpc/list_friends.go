package grpc

import (
	"context"
	"log"

	"social/internal/app/usecase/dto"

	pb "social/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *SocialController) ListFriends(ctx context.Context, req *pb.ListFriendsRequest) (*pb.ListFriendsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	friendsResponse, err := h.usecase.ListFriends(ctx, dto.ListFriendsDto{
		UserID: req.UserId,
		Limit:  req.Limit,
		Cursor: req.Cursor,
	})
	if err != nil {
		return nil, err
	}

	return &pb.ListFriendsResponse{
		FriendUserIds: newPbFriendRequestsIDsFromFriendRequests(friendsResponse.Friends),
		NextCursor:    friendsResponse.NextCursor,
	}, nil
}
