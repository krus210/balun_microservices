package grpc

import (
	"context"
	"log"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"

	pb "chat/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *ChatController) GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.GetChatResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	chat, err := h.usecase.GetChat(ctx, dto.GetChatDto{
		UserID: "1", // TODO: брать из хедера
		ChatID: models.ChatID(req.ChatId),
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetChatResponse{
		Chat: newPbChatFromChat(chat),
	}, nil
}
