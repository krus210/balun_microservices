package grpc

import (
	"context"
	"log"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"

	pb "chat/pkg/api"

	"google.golang.org/grpc/metadata"
)

func (h *ChatController) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	message, err := h.usecase.SendMessage(ctx, dto.SendMessageDto{
		UserID: "1", // TODO: брать из хедера
		ChatID: models.ChatID(req.ChatId),
		Text:   req.Text,
	})
	if err != nil {
		return nil, err
	}

	return &pb.SendMessageResponse{
		Message: newPbMessageFromMessage(message),
	}, nil
}
