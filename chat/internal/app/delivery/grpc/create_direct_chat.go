package grpc

import (
	"context"
	"log"

	"chat/internal/app/usecase/dto"

	pb "chat/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *ChatController) CreateDirectChat(ctx context.Context, req *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	chat, err := h.usecase.CreateDirectChat(ctx, dto.CreateDirectChatDto{
		UserID:        1, // TODO: брать из хедера
		ParticipantID: req.ParticipantId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateDirectChatResponse{
		ChatId: chat.ID,
	}, nil
}
