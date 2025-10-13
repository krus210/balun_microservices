package grpc

import (
	"context"
	"log"

	"chat/internal/app/models"
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
		UserID: models.UserID(1), // TODO: брать из хедера
		ChatID: models.ChatID(req.ChatId),
	})
	if err != nil {
		return nil, err
	}

	// Конвертация []models.UserID в []int64
	userIDsInt64 := make([]int64, len(userIDs))
	for i, id := range userIDs {
		userIDsInt64[i] = int64(id)
	}

	return &pb.ListChatMembersResponse{
		UserIds: userIDsInt64,
	}, nil
}
