package grpc

import (
	"context"
	"log"

	"chat/internal/app/usecase/dto"

	pb "chat/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *ChatController) ListChatMembers(ctx context.Context, req *pb.ListChatMembersRequest) (*pb.ListChatMembersResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	userIDs, err := h.usecase.ListChatMembers(ctx, dto.ListChatMembersDto{
		UserID: 1, // TODO: брать из хедера
		ChatID: req.ChatId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.ListChatMembersResponse{
		UserIds: userIDs,
	}, nil
}
