package grpc

import (
	"context"
	"log"

	"chat/internal/app/usecase/dto"

	pb "chat/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *ChatController) ListUserChats(ctx context.Context, req *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	chats, err := h.usecase.ListUserChats(ctx, dto.ListUserChatsDto{
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.ListUserChatsResponse{
		Chats: newPbChatsFromChats(chats),
	}, nil
}
